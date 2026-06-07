package main

import (
	"image"
	"os"
	"testing"

	"github.com/sarattha/lumago/engine/app"
	engineassets "github.com/sarattha/lumago/engine/assets"
	"github.com/sarattha/lumago/engine/graphics"
)

func TestRunnerStartsWithJump(t *testing.T) {
	state := newRunnerState()

	state.Step(1.0/runnerTargetFPS, runnerInput{Start: true, Jump: true})

	if !state.Started {
		t.Fatalf("runner did not start")
	}
	if state.PlayerVelY <= 0 {
		t.Fatalf("player velocity=%.2f, want upward jump in renderer coordinates", state.PlayerVelY)
	}
}

func TestRunnerStartsRunningByDefault(t *testing.T) {
	state := newRunnerState()

	if !state.Started {
		t.Fatal("runner should start by default so the HUD score is visibly live")
	}
}

func TestRunnerJumpMovesPlayerUpThenBackToGround(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	state.Obstacles = nil

	state.Step(1.0/runnerTargetFPS, runnerInput{Jump: true})
	if state.PlayerBottom <= runnerGroundY {
		t.Fatalf("player bottom=%.2f, want above ground %.2f after jump", state.PlayerBottom, float32(runnerGroundY))
	}

	for i := 0; i < runnerTargetFPS; i++ {
		state.Step(1.0/runnerTargetFPS, runnerInput{})
	}
	if !state.grounded() {
		t.Fatalf("player did not return to ground: bottom=%.2f velocity=%.2f", state.PlayerBottom, state.PlayerVelY)
	}
}

func TestRunnerDuckUsesLowerHitbox(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	standing := state.playerRect()

	state.Step(1.0/runnerTargetFPS, runnerInput{Duck: true})
	ducking := state.playerRect()

	if !state.Ducking {
		t.Fatalf("runner did not enter ducking state")
	}
	if ducking.H >= standing.H {
		t.Fatalf("ducking hitbox height=%.2f, want less than standing %.2f", ducking.H, standing.H)
	}
	if ducking.W <= standing.W {
		t.Fatalf("ducking hitbox width=%.2f, want wider than standing %.2f", ducking.W, standing.W)
	}
}

func TestRunnerAdvancesScoreAndSpeed(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	state.Obstacles = nil

	for i := 0; i < runnerTargetFPS; i++ {
		state.Step(1.0/runnerTargetFPS, runnerInput{})
	}

	if state.Score <= 0 {
		t.Fatalf("score=%d, want progress after running", state.Score)
	}
	if state.Speed <= runnerStartSpeed {
		t.Fatalf("speed=%.2f, want acceleration from %.2f", state.Speed, float32(runnerStartSpeed))
	}
}

func TestRunnerScoreDigitsAreLeftToRight(t *testing.T) {
	if got := runnerScoreDigits(43); got != [5]int{0, 0, 0, 4, 3} {
		t.Fatalf("score digits=%v, want 00043", got)
	}
	if got := runnerScoreDigits(123456); got != [5]int{2, 3, 4, 5, 6} {
		t.Fatalf("score digits=%v, want wrapped 23456", got)
	}
	if got := runnerScoreDigits(-1); got != [5]int{0, 0, 0, 0, 0} {
		t.Fatalf("score digits=%v, want clamped zero", got)
	}
}

func TestRunnerScoreHUDChangesWithScore(t *testing.T) {
	config := defaultRunnerConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	zero := newRunnerState()
	scored := newRunnerState()
	scored.Score = 43

	zeroSegments := scoreHUDSegmentStates(buildRunnerScene(game, zero, config).Sprites())
	scoredSegments := scoreHUDSegmentStates(buildRunnerScene(game, scored, config).Sprites())

	if len(zeroSegments) != 35 || len(scoredSegments) != 35 {
		t.Fatalf("score HUD segment counts zero=%d scored=%d, want 35 each", len(zeroSegments), len(scoredSegments))
	}
	if equalBoolSlices(zeroSegments, scoredSegments) {
		t.Fatalf("score HUD segments did not change between 00000 and 00043")
	}
}

func TestRunnerCollisionEndsRunAndRestartResets(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	state.Obstacles = []runnerObstacle{{Kind: runnerObstacleCactus, X: runnerDinoX}}

	state.Step(1.0/runnerTargetFPS, runnerInput{})

	if !state.GameOver {
		t.Fatalf("collision did not end the run")
	}
	state.Step(1.0/runnerTargetFPS, runnerInput{Restart: true})
	if state.GameOver || !state.Started {
		t.Fatalf("restart failed: started=%t gameOver=%t", state.Started, state.GameOver)
	}
	if state.Score != 0 {
		t.Fatalf("score=%d, want reset", state.Score)
	}
}

func TestRunnerSceneUsesReadableSpriteRolesWithoutLighting(t *testing.T) {
	config := defaultRunnerConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	state := newRunnerState()
	world := buildRunnerScene(game, state, config)

	if len(world.Sprites()) < 45 {
		t.Fatalf("sprites=%d, want composed runner graphics", len(world.Sprites()))
	}
	if len(world.Lights()) != 0 {
		t.Fatalf("lights=%d, want no runner scene lights", len(world.Lights()))
	}
	if len(world.Occluders()) != 0 {
		t.Fatalf("occluders=%d, want no runner shadow occluders", len(world.Occluders()))
	}
	counts := countRunnerLayers(world.Sprites())
	if counts[13] < 1 {
		t.Fatalf("dino sprite missing: layers=%v", counts)
	}
	if counts[9] < len(state.Obstacles) {
		t.Fatalf("rock obstacle sprites=%d, want at least %d", counts[9], len(state.Obstacles))
	}
	if counts[5] < 2 {
		t.Fatalf("road sprites=%d, want scrolling textured road", counts[5])
	}
	if counts[22] < 20 {
		t.Fatalf("score digit sprites=%d, want visible seven-segment score", counts[22])
	}
	if countRunnerSunMoonSprites(world.Sprites()) < 1 {
		t.Fatalf("sun/moon sprite missing")
	}
}

func TestRunnerRoadSpritesOverlap(t *testing.T) {
	config := defaultRunnerConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	state := newRunnerState()
	state.Distance = 123
	world := buildRunnerScene(game, state, config)

	var roads []graphics.SpriteDrawCommand
	for _, sprite := range world.Sprites() {
		if sprite.Layer == 5 {
			roads = append(roads, sprite)
		}
	}
	if len(roads) < 3 {
		t.Fatalf("road sprites=%d, want repeated road segments", len(roads))
	}
	for i := 1; i < len(roads); i++ {
		prevRight := roads[i-1].Transform.Position.X + roads[i-1].Transform.Scale.X*0.5
		nextLeft := roads[i].Transform.Position.X - roads[i].Transform.Scale.X*0.5
		if overlap := prevRight - nextLeft; overlap < runnerRoadOverlap-0.001 {
			t.Fatalf("road sprites %d/%d overlap=%.2f, want at least %.2f", i-1, i, overlap, float32(runnerRoadOverlap))
		}
	}
}

func TestRunnerLoadsVisualAssetTextures(t *testing.T) {
	config := defaultRunnerConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	buildRunnerScene(game, newRunnerState(), config)

	for _, asset := range []struct {
		Name   string
		Kind   string
		Src    image.Rectangle
		Width  int
		Height int
	}{
		{Name: "sky/dawn-sky.png", Kind: "sky", Src: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36},
		{Name: "sky/noon-sky.png", Kind: "sky", Src: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36},
		{Name: "sky/evening-sky.png", Kind: "sky", Src: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36},
		{Name: "sky/night-sky.png", Kind: "sky", Src: image.Rect(0, 0, 1672, 941), Width: 64, Height: 36},
		{Name: "sun.png", Kind: "sun", Src: image.Rect(130, 70, 930, 960), Width: 56, Height: 64},
		{Name: "moon.png", Kind: "moon", Src: image.Rect(120, 80, 930, 940), Width: 56, Height: 64},
		{Name: "road.png", Kind: "road", Src: image.Rect(0, 390, 1536, 650), Width: 64, Height: 14},
		{Name: "rock.png", Kind: "rock", Src: image.Rect(300, 285, 1220, 820), Width: 56, Height: 40},
		{Name: "dino.png", Kind: "dino", Src: image.Rect(180, 170, 1260, 840), Width: 64, Height: 40},
	} {
		path := runnerProcessedAssetPath(asset.Name, asset.Kind, asset.Src, asset.Width, asset.Height)
		info, ok := game.Assets.TextureByPath(path)
		if !ok {
			t.Fatalf("%s was not loaded", path)
		}
		if info.Width != asset.Width || info.Height != asset.Height {
			t.Fatalf("%s size=%dx%d, want %dx%d", path, info.Width, info.Height, asset.Width, asset.Height)
		}
		if info.Width*info.Height > 64*64 {
			t.Fatalf("%s has %d texels, want Vulkan texel lighting path limit <=4096", path, info.Width*info.Height)
		}
	}
}

func TestRunnerLoadsAssetManifest(t *testing.T) {
	catalog, err := loadRunnerAssetCatalog(defaultAssetMetadataPath)
	if err != nil {
		t.Fatalf("loadRunnerAssetCatalog returned error: %v", err)
	}
	if !catalog.Ready {
		t.Fatal("asset catalog is not ready")
	}
	if len(catalog.Manifest.Textures) != 9 || len(catalog.Manifest.Sprites) != 9 {
		t.Fatalf("manifest textures=%d sprites=%d, want 9 each", len(catalog.Manifest.Textures), len(catalog.Manifest.Sprites))
	}
	sprite, ok := catalog.SpritesByName["dino_run"]
	if !ok {
		t.Fatal("dino_run sprite missing from manifest")
	}
	if sprite.Rect != (engineassets.AssetRect{X: 180, Y: 170, W: 1080, H: 670}) {
		t.Fatalf("dino_run rect=%+v", sprite.Rect)
	}
	texture, ok := catalog.TexturesByID[sprite.TextureID]
	if !ok {
		t.Fatalf("texture id %q missing for dino_run", sprite.TextureID)
	}
	if texture.Filter != "nearest" || texture.Wrap != "clamp_to_edge" {
		t.Fatalf("dino texture sampling=%s/%s, want nearest/clamp_to_edge", texture.Filter, texture.Wrap)
	}
	if texture.TileSize.W != 16 || texture.TileSize.H != 16 || texture.PixelsPerUnit != 16 {
		t.Fatalf("dino texture profile ppu=%d tile=%+v, want 16 and 16x16", texture.PixelsPerUnit, texture.TileSize)
	}
	if len(catalog.Manifest.Atlases) != 1 || catalog.Manifest.Atlases[0].Padding != 2 || catalog.Manifest.Atlases[0].Extrusion != 1 {
		t.Fatalf("atlas bleed metadata missing: %+v", catalog.Manifest.Atlases)
	}
	if len(catalog.Manifest.NormalMaps) != 9 {
		t.Fatalf("neutral fallback normals=%d, want 9", len(catalog.Manifest.NormalMaps))
	}
}

func TestRunnerConfigAcceptsAssetMetadataOverride(t *testing.T) {
	t.Setenv("LUMAGO_ASSET_METADATA", "custom/assets.json")

	config := defaultRunnerConfig()
	applyRunnerEnvironment(&config)

	if config.AssetMetadata != "custom/assets.json" {
		t.Fatalf("asset metadata path=%q, want environment override", config.AssetMetadata)
	}
}

func TestPrepareRunnerAssetsEnablesDevelopmentHotReload(t *testing.T) {
	config := defaultRunnerConfig()
	config.Development = true

	catalog, reloader, err := prepareRunnerAssets(config)
	if err != nil {
		t.Fatalf("prepareRunnerAssets returned error: %v", err)
	}
	if !catalog.Ready {
		t.Fatal("catalog is not ready")
	}
	if reloader == nil {
		t.Fatal("development mode did not create a hot reloader")
	}
	if _, err := os.Stat(reloader.BaseDir()); err != nil {
		t.Fatalf("reloader base dir is invalid: %v", err)
	}
}

func TestRunnerGeneratedCutoutsHaveTransparentPixels(t *testing.T) {
	config := defaultRunnerConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	buildRunnerScene(game, newRunnerState(), config)

	for _, asset := range []struct {
		Name   string
		Kind   string
		Src    image.Rectangle
		Width  int
		Height int
	}{
		{Name: "sun.png", Kind: "sun", Src: image.Rect(130, 70, 930, 960), Width: 56, Height: 64},
		{Name: "moon.png", Kind: "moon", Src: image.Rect(120, 80, 930, 940), Width: 56, Height: 64},
		{Name: "rock.png", Kind: "rock", Src: image.Rect(300, 285, 1220, 820), Width: 56, Height: 40},
		{Name: "dino.png", Kind: "dino", Src: image.Rect(180, 170, 1260, 840), Width: 64, Height: 40},
	} {
		path := runnerProcessedAssetPath(asset.Name, asset.Kind, asset.Src, asset.Width, asset.Height)
		info, ok := game.Assets.TextureByPath(path)
		if !ok {
			t.Fatalf("%s was not loaded", path)
		}
		data, ok := graphics.RegisteredTextureData(info.ID)
		if !ok {
			t.Fatalf("%s texture data was not registered", path)
		}
		transparent := 0
		opaque := 0
		for _, pixel := range data.Pixels {
			if pixel.A <= 0.01 {
				transparent++
			}
			if pixel.A > 0.99 {
				opaque++
			}
		}
		if transparent == 0 || opaque == 0 {
			t.Fatalf("%s transparent=%d opaque=%d, want cutout with both", path, transparent, opaque)
		}
	}
}

func TestRunnerTimeOfDayMovesSunAndMoonRightToLeft(t *testing.T) {
	sunDawn := runnerSunPosition(0)
	sunNoon := runnerSunPosition(runnerDayCycle * 0.25)
	sunEvening := runnerSunPosition(runnerDayCycle * 0.50)
	if !(sunDawn.X > sunNoon.X && sunNoon.X > sunEvening.X) {
		t.Fatalf("sun x positions dawn/noon/evening = %.1f/%.1f/%.1f, want right-to-left", sunDawn.X, sunNoon.X, sunEvening.X)
	}
	if sunNoon.Y <= sunDawn.Y || sunNoon.Y <= sunEvening.Y {
		t.Fatalf("sun arc y dawn/noon/evening = %.1f/%.1f/%.1f, want highest near noon", sunDawn.Y, sunNoon.Y, sunEvening.Y)
	}

	moonEvening := runnerMoonPosition(runnerDayCycle * 0.50)
	moonNight := runnerMoonPosition(runnerDayCycle * 0.75)
	moonDawn := runnerMoonPosition(runnerDayCycle * 0.99)
	if !(moonEvening.X > moonNight.X && moonNight.X > moonDawn.X) {
		t.Fatalf("moon x positions evening/night/dawn = %.1f/%.1f/%.1f, want right-to-left", moonEvening.X, moonNight.X, moonDawn.X)
	}
}

func countRunnerLayers(sprites []graphics.SpriteDrawCommand) map[int]int {
	counts := map[int]int{}
	for _, sprite := range sprites {
		counts[sprite.Layer]++
	}
	return counts
}

func countRunnerSunMoonSprites(sprites []graphics.SpriteDrawCommand) int {
	count := 0
	for _, sprite := range sprites {
		if sprite.Layer == 3 && sprite.Sprite.Material.Albedo != graphics.InvalidTexture && sprite.Sprite.Material.Emissive > 1 {
			count++
		}
	}
	return count
}

func scoreHUDSegmentStates(sprites []graphics.SpriteDrawCommand) []bool {
	states := []bool{}
	for _, sprite := range sprites {
		if sprite.Layer == 22 {
			states = append(states, sprite.Sprite.Color.R > 0.5)
		}
	}
	return states
}

func equalBoolSlices(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
