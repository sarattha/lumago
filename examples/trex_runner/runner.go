package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"

	"github.com/sarattha/lumago/engine/app"
	engineassets "github.com/sarattha/lumago/engine/assets"
	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	"github.com/sarattha/lumago/engine/scene"
)

const (
	runnerTargetWidth  = 1280
	runnerTargetHeight = 720
	runnerTargetFPS    = 60
	runnerGroundY      = 220
	runnerDinoX        = 210
	runnerGravity      = -3000
	runnerJumpVelocity = 1060
	runnerStartSpeed   = 520
	runnerMaxSpeed     = 930
	runnerDayCycle     = 48
	runnerRoadWidth    = 760
	runnerRoadOverlap  = 12
)

type runnerInput struct {
	Start   bool
	Jump    bool
	Duck    bool
	Restart bool
}

type runnerState struct {
	Started      bool
	GameOver     bool
	Time         float32
	Distance     float32
	Score        int
	Speed        float32
	PlayerBottom float32
	PlayerVelY   float32
	Ducking      bool
	Obstacles    []runnerObstacle
	Clouds       []runnerCloud
	spawnTimer   float32
	spawnIndex   int
}

type runnerObstacleKind uint8

const (
	runnerObstacleCactus runnerObstacleKind = iota
	runnerObstacleCactusCluster
	runnerObstacleBird
)

type runnerObstacle struct {
	Kind runnerObstacleKind
	X    float32
}

type runnerCloud struct {
	X     float32
	Y     float32
	Scale float32
}

func newRunnerState() runnerState {
	state := runnerState{}
	state.Reset()
	return state
}

func (s *runnerState) Reset() {
	s.Started = true
	s.GameOver = false
	s.Time = 0
	s.Distance = 0
	s.Score = 0
	s.Speed = runnerStartSpeed
	s.PlayerBottom = runnerGroundY
	s.PlayerVelY = 0
	s.Ducking = false
	s.spawnTimer = 1.05
	s.spawnIndex = 0
	s.Obstacles = []runnerObstacle{
		{Kind: runnerObstacleCactus, X: 780},
		{Kind: runnerObstacleCactusCluster, X: 1120},
	}
	s.Clouds = []runnerCloud{
		{X: 170, Y: 575, Scale: 1.0},
		{X: 520, Y: 620, Scale: 1.25},
		{X: 915, Y: 550, Scale: 0.9},
	}
}

func (s *runnerState) Step(dt float32, input runnerInput) {
	if dt <= 0 {
		return
	}
	if s.GameOver {
		if input.Restart || input.Start || input.Jump {
			s.Reset()
		}
		return
	}
	if !s.Started {
		s.Ducking = input.Duck
		if input.Start || input.Jump {
			s.Started = true
			if input.Jump {
				s.PlayerVelY = runnerJumpVelocity
			}
		}
		return
	}

	s.Time += dt
	s.Speed = minRunner(runnerMaxSpeed, s.Speed+18*dt)
	s.Distance += s.Speed * dt
	s.Score = int(s.Distance / 12)

	grounded := s.grounded()
	s.Ducking = input.Duck && grounded
	if input.Jump && grounded && !s.Ducking {
		s.PlayerVelY = runnerJumpVelocity
	}
	s.PlayerBottom += s.PlayerVelY * dt
	s.PlayerVelY += runnerGravity * dt
	if s.PlayerBottom <= runnerGroundY {
		s.PlayerBottom = runnerGroundY
		s.PlayerVelY = 0
	}

	for i := range s.Obstacles {
		s.Obstacles[i].X -= s.Speed * dt
	}
	s.trimObstacles()
	s.spawnTimer -= dt
	if s.spawnTimer <= 0 {
		s.spawnObstacle()
	}
	for i := range s.Clouds {
		s.Clouds[i].X -= s.Speed * dt * (0.12 + 0.03*float32(i%2))
		if s.Clouds[i].X < -120 {
			s.Clouds[i].X = runnerTargetWidth + 120 + float32(i*90)
		}
	}

	player := s.playerRect()
	for _, obstacle := range s.Obstacles {
		if rectsOverlap(player, obstacle.Rect()) {
			s.GameOver = true
			s.Started = false
			return
		}
	}
}

func (s *runnerState) grounded() bool {
	return s.PlayerBottom <= runnerGroundY+0.001 && s.PlayerVelY == 0
}

func (s *runnerState) playerRect() lmath.Rect {
	if s.Ducking {
		return lmath.Rect{X: runnerDinoX - 47, Y: s.PlayerBottom, W: 94, H: 50}
	}
	return lmath.Rect{X: runnerDinoX - 34, Y: s.PlayerBottom, W: 68, H: 92}
}

func (s *runnerState) trimObstacles() {
	keep := s.Obstacles[:0]
	for _, obstacle := range s.Obstacles {
		if obstacle.X > -120 {
			keep = append(keep, obstacle)
		}
	}
	s.Obstacles = keep
}

func (s *runnerState) spawnObstacle() {
	kinds := []runnerObstacleKind{
		runnerObstacleCactus,
		runnerObstacleCactusCluster,
		runnerObstacleCactus,
		runnerObstacleCactusCluster,
	}
	kind := kinds[s.spawnIndex%len(kinds)]
	s.spawnIndex++
	s.Obstacles = append(s.Obstacles, runnerObstacle{Kind: kind, X: runnerTargetWidth + 95})
	gap := 0.95 + float32((s.spawnIndex%3))*0.18
	s.spawnTimer = gap
}

func (o runnerObstacle) Rect() lmath.Rect {
	switch o.Kind {
	case runnerObstacleCactusCluster:
		return lmath.Rect{X: o.X - 48, Y: runnerGroundY, W: 96, H: 78}
	default:
		return lmath.Rect{X: o.X - 32, Y: runnerGroundY, W: 64, H: 66}
	}
}

type runnerMaterialSet struct {
	SkyDawn    graphics.Material2D
	SkyNoon    graphics.Material2D
	SkyEvening graphics.Material2D
	SkyNight   graphics.Material2D
	Sun        graphics.Material2D
	Moon       graphics.Material2D
	Road       graphics.Material2D
	Rock       graphics.Material2D
	Dino       graphics.Material2D
}

func buildRunnerScene(game *app.Game, state runnerState, config runnerConfig) *scene.Scene {
	world := scene.New()
	materials := runnerMaterials(game, config)
	addRunnerSky(world, state, materials)
	addRunnerTrack(world, state, materials)
	addRunnerObstacles(world, state, materials)
	addRunnerDino(world, state, materials)
	addRunnerScore(world, state)
	return world
}

type runnerAssetCatalog struct {
	Ready         bool
	BaseDir       string
	Manifest      engineassets.AssetManifest
	TexturesByID  map[string]engineassets.ManifestTexture
	SpritesByName map[string]engineassets.ManifestSprite
}

type runnerMaterialSpec struct {
	Sprite         string
	Variant        string
	FallbackSource string
	FallbackSrc    image.Rectangle
	Width          int
	Height         int
	Roughness      float32
	Emissive       float32
	Alpha          runnerAlphaFunc
}

func runnerMaterials(game *app.Game, config runnerConfig) runnerMaterialSet {
	catalog := config.AssetCatalog
	if !catalog.Ready {
		if loaded, err := loadRunnerAssetCatalog(config.AssetMetadata); err == nil {
			catalog = loaded
		}
	}
	return runnerMaterialSet{
		SkyDawn: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "sky_dawn_full", Variant: "sky", FallbackSource: "sky/dawn-sky.png", FallbackSrc: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36, Roughness: 0.92, Emissive: 0.10, Alpha: runnerKeepOpaque,
		}),
		SkyNoon: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "sky_noon_full", Variant: "sky", FallbackSource: "sky/noon-sky.png", FallbackSrc: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36, Roughness: 0.92, Emissive: 0.16, Alpha: runnerKeepOpaque,
		}),
		SkyEvening: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "sky_evening_full", Variant: "sky", FallbackSource: "sky/evening-sky.png", FallbackSrc: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36, Roughness: 0.92, Emissive: 0.12, Alpha: runnerKeepOpaque,
		}),
		SkyNight: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "sky_night_full", Variant: "sky", FallbackSource: "sky/night-sky.png", FallbackSrc: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36, Roughness: 0.92, Emissive: 0.04, Alpha: runnerKeepOpaque,
		}),
		Sun: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "sun_disc", Variant: "sun", FallbackSource: "sun.png", FallbackSrc: image.Rect(130, 70, 930, 960), Width: 56, Height: 64, Roughness: 0.25, Emissive: 2.8, Alpha: runnerBrightAlpha,
		}),
		Moon: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "moon_disc", Variant: "moon", FallbackSource: "moon.png", FallbackSrc: image.Rect(120, 80, 930, 940), Width: 56, Height: 64, Roughness: 0.30, Emissive: 1.8, Alpha: runnerBrightAlpha,
		}),
		Road: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "road_strip", Variant: "road", FallbackSource: "road.png", FallbackSrc: image.Rect(0, 390, 1536, 650), Width: 64, Height: 14, Roughness: 0.58, Emissive: 0.03, Alpha: runnerKeepOpaque,
		}),
		Rock: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "rock_obstacle", Variant: "rock", FallbackSource: "rock.png", FallbackSrc: image.Rect(300, 285, 1220, 820), Width: 56, Height: 40, Roughness: 0.68, Emissive: 0.04, Alpha: runnerRockAlpha,
		}),
		Dino: runnerMaterialFromSpec(game, catalog, runnerMaterialSpec{
			Sprite: "dino_run", Variant: "dino", FallbackSource: "dino.png", FallbackSrc: image.Rect(180, 170, 1260, 840), Width: 64, Height: 40, Roughness: 0.42, Emissive: 0.10, Alpha: runnerDinoAlpha,
		}),
	}
}

type runnerAlphaFunc func(x, y, width, height int, color lmath.Color) float32

func loadRunnerAssetCatalog(path string) (runnerAssetCatalog, error) {
	metadataPath := resolveRunnerAssetMetadataPath(path)
	manifest, err := engineassets.ImportAssetMetadataFile(metadataPath)
	if err != nil {
		return runnerAssetCatalog{}, err
	}
	return runnerAssetCatalogFromManifest(manifest, filepath.Dir(metadataPath)), nil
}

func runnerAssetCatalogFromManifest(manifest engineassets.AssetManifest, baseDir string) runnerAssetCatalog {
	catalog := runnerAssetCatalog{
		Ready:         true,
		BaseDir:       baseDir,
		Manifest:      manifest,
		TexturesByID:  make(map[string]engineassets.ManifestTexture, len(manifest.Textures)),
		SpritesByName: make(map[string]engineassets.ManifestSprite, len(manifest.Sprites)),
	}
	for _, texture := range manifest.Textures {
		catalog.TexturesByID[texture.ID] = texture
	}
	for _, sprite := range manifest.Sprites {
		catalog.SpritesByName[sprite.Name] = sprite
	}
	return catalog
}

func resolveRunnerAssetMetadataPath(path string) string {
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if _, err := os.Stat(defaultAssetMetadataPath); err == nil {
		return defaultAssetMetadataPath
	}
	return fallbackAssetMetadataPath
}

func runnerMaterialFromSpec(game *app.Game, catalog runnerAssetCatalog, spec runnerMaterialSpec) graphics.Material2D {
	if catalog.Ready {
		if sprite, ok := catalog.SpritesByName[spec.Sprite]; ok {
			if texture, ok := catalog.TexturesByID[sprite.TextureID]; ok {
				path := runnerCatalogAssetPath(catalog, texture.Source)
				src := image.Rect(sprite.Rect.X, sprite.Rect.Y, sprite.Rect.X+sprite.Rect.W, sprite.Rect.Y+sprite.Rect.H)
				return runnerMaterialRegion(game, texture.Source, path, spec.Variant, src, spec.Width, spec.Height, spec.Roughness, spec.Emissive, spec.Alpha)
			}
		}
	}
	return runnerMaterialRegion(game, spec.FallbackSource, runnerAssetPath(spec.FallbackSource), spec.Variant, spec.FallbackSrc, spec.Width, spec.Height, spec.Roughness, spec.Emissive, spec.Alpha)
}

func runnerCatalogAssetPath(catalog runnerAssetCatalog, source string) string {
	if filepath.IsAbs(source) {
		return source
	}
	return filepath.Join(catalog.BaseDir, filepath.FromSlash(source))
}

func runnerMaterialRegion(game *app.Game, name, path, variant string, src image.Rectangle, width, height int, roughness, emissive float32, alpha runnerAlphaFunc) graphics.Material2D {
	key := runnerProcessedAssetPath(name, variant, src, width, height)
	if info, ok := game.Assets.TextureByPath(key); ok {
		return graphics.Material2D{
			Albedo:    info.ID,
			Roughness: roughness,
			Emissive:  emissive,
		}
	}
	return graphics.Material2D{
		Albedo:    game.Assets.RegisterGeneratedTexture(key, width, height, runnerProcessedPixels(path, src, width, height, alpha)),
		Roughness: roughness,
		Emissive:  emissive,
	}
}

func runnerProcessedAssetPath(name, variant string, src image.Rectangle, width, height int) string {
	return filepath.Join("generated", "trex_runner", variant, name) + "#" + src.String() + "@" + runnerTextureSizeKey(width, height)
}

func runnerTextureSizeKey(width, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

func runnerProcessedPixels(path string, src image.Rectangle, width, height int, alpha runnerAlphaFunc) []lmath.Color {
	file, err := os.Open(path)
	if err != nil {
		return runnerFallbackPixels(width, height)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return runnerFallbackPixels(width, height)
	}
	bounds := img.Bounds()
	src = src.Intersect(bounds)
	if src.Empty() {
		src = bounds
	}
	pixels := make([]lmath.Color, 0, width*height)
	for y := 0; y < height; y++ {
		sy := src.Min.Y + int((float32(y)+0.5)*float32(src.Dy())/float32(height))
		if sy >= src.Max.Y {
			sy = src.Max.Y - 1
		}
		for x := 0; x < width; x++ {
			sx := src.Min.X + int((float32(x)+0.5)*float32(src.Dx())/float32(width))
			if sx >= src.Max.X {
				sx = src.Max.X - 1
			}
			r, g, b, a := img.At(sx, sy).RGBA()
			color := lmath.Color{
				R: float32(r) / 65535,
				G: float32(g) / 65535,
				B: float32(b) / 65535,
				A: float32(a) / 65535,
			}
			if alpha != nil {
				color.A *= alpha(x, y, width, height, color)
			}
			pixels = append(pixels, color)
		}
	}
	return pixels
}

func runnerFallbackPixels(width, height int) []lmath.Color {
	pixels := make([]lmath.Color, width*height)
	for i := range pixels {
		pixels[i] = lmath.Color{R: 1, G: 0, B: 1, A: 1}
	}
	return pixels
}

func runnerKeepOpaque(x, y, width, height int, color lmath.Color) float32 {
	return 1
}

func runnerBrightAlpha(x, y, width, height int, color lmath.Color) float32 {
	if maxColor(color) < 0.10 {
		return 0
	}
	return 1
}

func runnerDinoAlpha(x, y, width, height int, color lmath.Color) float32 {
	nx := (float32(x) + 0.5) / float32(width)
	ny := (float32(y) + 0.5) / float32(height)
	if nx < 0.03 || nx > 0.98 || ny < 0.08 || ny > 0.96 {
		return 0
	}
	if colorSaturation(color) > 0.18 || color.G > color.R*1.08 || color.R > 0.45 && color.G > 0.32 && color.B < 0.25 {
		return 1
	}
	if ny > 0.78 && nx > 0.33 && nx < 0.83 && maxColor(color) > 0.18 {
		return 1
	}
	return 0
}

func runnerRockAlpha(x, y, width, height int, color lmath.Color) float32 {
	nx := (float32(x) + 0.5) / float32(width)
	ny := (float32(y) + 0.5) / float32(height)
	center := float32(math.Abs(float64(nx - 0.52)))
	top := 0.13 + center*0.72
	if nx < 0.06 || nx > 0.95 || ny < top || ny > 0.92 {
		return 0
	}
	if maxColor(color) < 0.08 {
		return 0
	}
	return 1
}

func maxColor(color lmath.Color) float32 {
	return maxRunner(color.R, maxRunner(color.G, color.B))
}

func minColor(color lmath.Color) float32 {
	return minRunner(color.R, minRunner(color.G, color.B))
}

func colorSaturation(color lmath.Color) float32 {
	max := maxColor(color)
	if max == 0 {
		return 0
	}
	return (max - minColor(color)) / max
}

func runnerAssetPath(name string) string {
	path := filepath.Join("examples", "trex_runner", "assets", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join("assets", name)
}

func addRunnerSky(world *scene.Scene, state runnerState, materials runnerMaterialSet) {
	material, color := runnerSkyMaterialAndTint(materials, state.Time)
	addRunnerSprite(world, material, lmath.Rect{W: 1, H: 1}, 640, 360, 1280, 720, color, 0, 0)
	for _, cloud := range state.Clouds {
		addRunnerCloud(world, cloud)
	}
	addRunnerSunMoon(world, state, materials)
}

func addRunnerCloud(world *scene.Scene, cloud runnerCloud) {
	c := lmath.Color{R: 0.78, G: 0.86, B: 0.96, A: 1}
	addRunnerRect(world, cloud.X, cloud.Y, 92*cloud.Scale, 22*cloud.Scale, c, 2, 0.45)
	addRunnerRect(world, cloud.X-28*cloud.Scale, cloud.Y-12*cloud.Scale, 34*cloud.Scale, 28*cloud.Scale, c, 2, 0.45)
	addRunnerRect(world, cloud.X+16*cloud.Scale, cloud.Y-18*cloud.Scale, 44*cloud.Scale, 36*cloud.Scale, c, 2, 0.45)
}

func addRunnerSunMoon(world *scene.Scene, state runnerState, materials runnerMaterialSet) {
	if runnerSunVisible(state.Time) {
		pos := runnerSunPosition(state.Time)
		addRunnerSprite(world, materials.Sun, lmath.Rect{W: 1, H: 1}, pos.X, pos.Y, 118, 118, lmath.White(), 3, 2.8)
	}
	if runnerMoonVisible(state.Time) {
		pos := runnerMoonPosition(state.Time)
		addRunnerSprite(world, materials.Moon, lmath.Rect{W: 1, H: 1}, pos.X, pos.Y, 108, 108, lmath.Color{R: 0.86, G: 0.92, B: 1.00, A: 1}, 3, 1.8)
	}
}

func addRunnerTrack(world *scene.Scene, state runnerState, materials runnerMaterialSet) {
	addRunnerRect(world, 640, runnerGroundY-76, 1280, 118, lmath.Color{R: 0.14, G: 0.11, B: 0.09, A: 1}, 4, 0)
	roadStep := runnerRoadWidth - runnerRoadOverlap
	offset := float32(math.Mod(float64(state.Distance), float64(roadStep)))
	for x := -offset - float32(roadStep); x < runnerTargetWidth+runnerRoadWidth; x += float32(roadStep) {
		addRunnerSprite(world, materials.Road, lmath.Rect{W: 1, H: 1}, x+runnerRoadWidth/2, runnerGroundY-58, runnerRoadWidth, 164, lmath.White(), 5, 0.04)
	}
}

func addRunnerObstacles(world *scene.Scene, state runnerState, materials runnerMaterialSet) {
	for _, obstacle := range state.Obstacles {
		switch obstacle.Kind {
		case runnerObstacleCactusCluster:
			addRunnerRock(world, materials, obstacle.X-24, 82, 0.96, -0.04)
			addRunnerRock(world, materials, obstacle.X+26, 66, 0.78, 0.06)
		default:
			addRunnerRock(world, materials, obstacle.X, 72, 0.82, 0)
		}
	}
}

func addRunnerRock(world *scene.Scene, materials runnerMaterialSet, x, height, scale, rotation float32) {
	addRunnerRect(world, x, runnerGroundY-5, 78*scale, 18*scale, lmath.Color{R: 0.03, G: 0.03, B: 0.03, A: 0.45}, 7, 0)
	addRunnerSpriteRotated(world, materials.Rock, lmath.Rect{W: 1, H: 1}, x, runnerGroundY+height/2-3, height*1.28*scale, height*scale, rotation, lmath.White(), 9, 0.02)
}

func addRunnerDino(world *scene.Scene, state runnerState, materials runnerMaterialSet) {
	rect := state.playerRect()
	x := rect.X + rect.W/2
	addRunnerRect(world, x, state.PlayerBottom-10, 126, 18, lmath.Color{R: 0.00, G: 0.00, B: 0.00, A: 0.45}, 7, 0)
	if state.Ducking {
		addRunnerSprite(world, materials.Dino, lmath.Rect{W: 1, H: 1}, x+6, state.PlayerBottom+48, 168, 118, lmath.White(), 13, 0.12)
		return
	}
	addRunnerSprite(world, materials.Dino, lmath.Rect{W: 1, H: 1}, x+5, state.PlayerBottom+62, 178, 144, lmath.White(), 13, 0.12)
}

func addRunnerScore(world *scene.Scene, state runnerState) {
	addRunnerRect(world, 1075, 650, 250, 52, lmath.Color{R: 0.03, G: 0.04, B: 0.05, A: 1}, 20, 0)
	for i, digit := range runnerScoreDigits(state.Score) {
		addRunnerDigit(world, float32(1052+i*32), 650, digit)
	}
	if !state.Started && !state.GameOver {
		addRunnerRect(world, 640, 294, 320, 52, lmath.Color{R: 0.05, G: 0.06, B: 0.08, A: 1}, 20, 0)
		addRunnerRect(world, 640, 294, 258, 8, lmath.Color{R: 0.95, G: 0.90, B: 0.66, A: 1}, 21, 2.4)
		addRunnerRect(world, 640, 313, 180, 6, lmath.Color{R: 0.45, G: 0.85, B: 1.00, A: 1}, 21, 2.0)
	}
	if state.GameOver {
		addRunnerRect(world, 640, 294, 330, 58, lmath.Color{R: 0.12, G: 0.03, B: 0.05, A: 1}, 20, 0)
		addRunnerRect(world, 640, 286, 260, 9, lmath.Color{R: 1.00, G: 0.28, B: 0.25, A: 1}, 21, 2.4)
		addRunnerRect(world, 640, 309, 210, 7, lmath.Color{R: 1.00, G: 0.84, B: 0.48, A: 1}, 21, 2.0)
	}
}

func runnerScoreDigits(score int) [5]int {
	if score < 0 {
		score = 0
	}
	score %= 100000
	digits := [5]int{}
	for i := len(digits) - 1; i >= 0; i-- {
		digits[i] = score % 10
		score /= 10
	}
	return digits
}

func addRunnerDigit(world *scene.Scene, x, y float32, digit int) {
	segments := [10][7]bool{
		{true, true, true, true, true, true, false},
		{false, true, true, false, false, false, false},
		{true, true, false, true, true, false, true},
		{true, true, true, true, false, false, true},
		{false, true, true, false, false, true, true},
		{true, false, true, true, false, true, true},
		{true, false, true, true, true, true, true},
		{true, true, true, false, false, false, false},
		{true, true, true, true, true, true, true},
		{true, true, true, true, false, true, true},
	}
	color := lmath.Color{R: 0.92, G: 0.96, B: 1.00, A: 1}
	dim := lmath.Color{R: 0.17, G: 0.20, B: 0.25, A: 1}
	for i, on := range segments[digit] {
		c := dim
		if on {
			c = color
		}
		switch i {
		case 0:
			addRunnerRect(world, x, y-18, 20, 5, c, 22, 1.8)
		case 1:
			addRunnerRect(world, x+12, y-8, 5, 18, c, 22, 1.8)
		case 2:
			addRunnerRect(world, x+12, y+12, 5, 18, c, 22, 1.8)
		case 3:
			addRunnerRect(world, x, y+22, 20, 5, c, 22, 1.8)
		case 4:
			addRunnerRect(world, x-12, y+12, 5, 18, c, 22, 1.8)
		case 5:
			addRunnerRect(world, x-12, y-8, 5, 18, c, 22, 1.8)
		case 6:
			addRunnerRect(world, x, y+2, 20, 5, c, 22, 1.8)
		}
	}
}

func addRunnerRect(world *scene.Scene, x, y, w, h float32, color lmath.Color, layer int, emissive float32) {
	addRunnerSprite(world, graphics.Material2D{Roughness: 0.55, Emissive: emissive}, lmath.Rect{W: 1, H: 1}, x, y, w, h, color, layer, emissive)
}

func addRunnerSprite(world *scene.Scene, material graphics.Material2D, src lmath.Rect, x, y, w, h float32, color lmath.Color, layer int, emissive float32) {
	addRunnerSpriteRotated(world, material, src, x, y, w, h, 0, color, layer, emissive)
}

func addRunnerSpriteRotated(world *scene.Scene, material graphics.Material2D, src lmath.Rect, x, y, w, h, rotation float32, color lmath.Color, layer int, emissive float32) {
	if src.W == 0 {
		src.W = 1
	}
	if src.H == 0 {
		src.H = 1
	}
	material.Emissive = emissive
	world.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Material: material,
			Src:      src,
			Color:    color,
		},
		Transform: graphics.Transform2D{
			Position: lmath.Vec2{X: x, Y: y},
			Scale:    lmath.Vec2{X: w / src.W, Y: h / src.H},
			Rotation: rotation,
			Z:        float32(layer),
		},
		Layer: layer,
	})
}

func runnerSkyMaterialAndTint(materials runnerMaterialSet, t float32) (graphics.Material2D, lmath.Color) {
	phase := runnerDayPhase(t)
	switch {
	case phase < 0.25:
		return materials.SkyDawn, lerpColor(runnerDawnTint(), runnerNoonTint(), phase/0.25)
	case phase < 0.50:
		return materials.SkyNoon, lerpColor(runnerNoonTint(), runnerEveningTint(), (phase-0.25)/0.25)
	case phase < 0.75:
		return materials.SkyEvening, lerpColor(runnerEveningTint(), runnerNightTint(), (phase-0.50)/0.25)
	default:
		return materials.SkyNight, lerpColor(runnerNightTint(), runnerDawnTint(), (phase-0.75)/0.25)
	}
}

func runnerAmbient(t float32) lmath.Color {
	phase := runnerDayPhase(t)
	switch {
	case phase < 0.25:
		return lerpColor(lmath.Color{R: 0.30, G: 0.25, B: 0.23, A: 1}, lmath.Color{R: 0.52, G: 0.50, B: 0.45, A: 1}, phase/0.25)
	case phase < 0.50:
		return lerpColor(lmath.Color{R: 0.52, G: 0.50, B: 0.45, A: 1}, lmath.Color{R: 0.36, G: 0.29, B: 0.25, A: 1}, (phase-0.25)/0.25)
	case phase < 0.75:
		return lerpColor(lmath.Color{R: 0.36, G: 0.29, B: 0.25, A: 1}, lmath.Color{R: 0.13, G: 0.15, B: 0.24, A: 1}, (phase-0.50)/0.25)
	default:
		return lerpColor(lmath.Color{R: 0.13, G: 0.15, B: 0.24, A: 1}, lmath.Color{R: 0.30, G: 0.25, B: 0.23, A: 1}, (phase-0.75)/0.25)
	}
}

func runnerDawnTint() lmath.Color {
	return lmath.Color{R: 1.00, G: 0.92, B: 0.84, A: 1}
}

func runnerNoonTint() lmath.Color {
	return lmath.Color{R: 1.00, G: 1.00, B: 0.96, A: 1}
}

func runnerEveningTint() lmath.Color {
	return lmath.Color{R: 1.00, G: 0.78, B: 0.62, A: 1}
}

func runnerNightTint() lmath.Color {
	return lmath.Color{R: 0.48, G: 0.56, B: 0.84, A: 1}
}

func runnerSunVisible(t float32) bool {
	return runnerDayPhase(t) < 0.58
}

func runnerMoonVisible(t float32) bool {
	return runnerDayPhase(t) >= 0.42
}

func runnerSunPosition(t float32) lmath.Vec2 {
	progress := clampRunner(runnerDayPhase(t)/0.58, 0, 1)
	return runnerArcPosition(progress)
}

func runnerMoonPosition(t float32) lmath.Vec2 {
	phase := runnerDayPhase(t)
	progress := (phase - 0.42) / 0.58
	if progress < 0 {
		progress += 1
	}
	return runnerArcPosition(clampRunner(progress, 0, 1))
}

func runnerArcPosition(progress float32) lmath.Vec2 {
	x := 1220 - progress*1540
	y := 438 + 178*float32(math.Sin(float64(progress*math.Pi)))
	return lmath.Vec2{X: x, Y: y}
}

func runnerDayPhase(t float32) float32 {
	if t <= 0 {
		return 0
	}
	return float32(math.Mod(float64(t), runnerDayCycle)) / runnerDayCycle
}

func lerpColor(a, b lmath.Color, t float32) lmath.Color {
	t = clampRunner(t, 0, 1)
	return lmath.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}

func clampRunner(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func rectsOverlap(a, b lmath.Rect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func minRunner(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxRunner(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
