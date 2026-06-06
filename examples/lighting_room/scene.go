package main

import (
	"math"
	"os"
	"path/filepath"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	"github.com/sarattha/lumago/engine/scene"
)

const (
	acceptanceSprites      = 1000
	acceptanceMaterials    = 5
	acceptanceLights       = 4
	acceptanceShadowLights = 2
	acceptanceOccluders    = 20
	acceptanceTargetFPS    = 60
	acceptanceTargetWidth  = 1920
	acceptanceTargetHeight = 1080
	acceptanceTileColumns  = 40
	acceptanceTileRows     = 25
	acceptanceTileSize     = 34
	acceptanceTileSpacingX = 43
	acceptanceTileSpacingY = 36
	acceptanceOriginX      = 122
	acceptanceOriginY      = 91
	authoredTileSpacing    = 34
	authoredOriginX        = 297
	authoredOriginY        = 132
)

type spriteRole int

const (
	spriteRoleFloor spriteRole = iota
	spriteRoleWall
	spriteRoleProp
	spriteRoleLightMarker
	spriteRoleLegend
)

type roomSpriteSpec struct {
	MaterialIndex int
	Role          spriteRole
	Src           lmath.Rect
	Position      lmath.Vec2
	Scale         lmath.Vec2
	Rotation      float32
	Color         lmath.Color
	Layer         int
}

func buildLightingRoom(game *app.Game, config demoConfig) *scene.Scene {
	world := scene.New()
	world.SetLightingConfig(graphics.LightingConfig2D{
		Ambient:    lmath.Color{R: 0.10, G: 0.10, B: 0.13, A: 1},
		DebugView:  config.DebugView,
		ShadowMode: config.ShadowMode,
	})

	materials := []graphics.Material2D{
		material(game, "floor", 0.80, 0),
		material(game, "wall", 0.65, 0),
		material(game, "character", 0.45, 0.05),
		material(game, "prop", 0.50, 0.02),
		material(game, "floor", 0.20, 3.0),
	}
	addAcceptanceSprites(world, materials)
	addAcceptanceOccluders(world)
	updateLights(world, 0)
	return world
}

func addAcceptanceSprites(world *scene.Scene, materials []graphics.Material2D) {
	for y := 0; y < acceptanceTileRows; y++ {
		for x := 0; x < acceptanceTileColumns; x++ {
			spec := roomSpriteSpecAt(x, y)
			addSprite(world, materials[spec.MaterialIndex], spec)
		}
	}
}

func roomSpriteSpecAt(x, y int) roomSpriteSpec {
	spec := roomSpriteSpec{
		MaterialIndex: 0,
		Role:          spriteRoleFloor,
		Src:           lmath.Rect{W: 1, H: 1},
		Position:      lmath.Vec2{X: tileCenterX(x), Y: tileCenterY(y)},
		Scale:         spriteScale(0.92, 0.92),
		Color:         floorTileColor(x, y),
		Layer:         0,
	}
	if color, ok := legendTileColor(x, y); ok {
		spec.MaterialIndex = 4
		spec.Role = spriteRoleLegend
		spec.Src = solidSpriteSrc()
		spec.Scale = spriteScaleForSrc(1.08, 1.08, spec.Src)
		spec.Color = color
		spec.Layer = 5
		return spec
	}
	if marker, ok := lightMarkerTileColor(x, y); ok {
		spec.MaterialIndex = 4
		spec.Role = spriteRoleLightMarker
		spec.Src = solidSpriteSrc()
		spec.Scale = marker.Scale
		spec.Color = marker.Color
		spec.Layer = 4
		return spec
	}
	if label, ok := labelTileAt(x, y); ok {
		spec.MaterialIndex = 4
		spec.Role = spriteRoleLegend
		spec.Src = solidSpriteSrc()
		spec.Scale = spriteScaleForSrc(1.12, 1.12, spec.Src)
		spec.Color = label.Color
		spec.Layer = 5
		return spec
	}
	if isWallTile(x, y) {
		spec.MaterialIndex = 1
		spec.Role = spriteRoleWall
		spec.Scale = wallTileScale(x, y)
		spec.Color = wallTileColor(x, y)
		spec.Layer = 3
		return spec
	}
	if color, ok := propTileColor(x, y); ok {
		spec.MaterialIndex = propMaterialIndex(x, y)
		spec.Role = spriteRoleProp
		spec.Scale = spriteScale(1.08, 1.08)
		spec.Rotation = propTileRotation(x, y)
		spec.Color = color
		spec.Layer = 2
		return spec
	}
	return spec
}

func roomMaterialIndex(x, y int) int {
	return roomSpriteSpecAt(x, y).MaterialIndex
}

func isWallTile(x, y int) bool {
	return x <= 1 || y <= 1 || x >= acceptanceTileColumns-2 || y >= acceptanceTileRows-2 ||
		((y == 5 || y == 6) && x > 4 && x < 16) ||
		((y == 5 || y == 6) && x > 24 && x < 36) ||
		((x == 12 || x == 13) && y > 5 && y < 14) ||
		((x == 28 || x == 29) && y > 11 && y < 21)
}

func inRectTile(x, y, left, top, width, height int) bool {
	return x >= left && x < left+width && y >= top && y < top+height
}

func tileJitter(x, y int) float32 {
	return float32((x*3+y*5)%6) * 0.008
}

func spriteScale(x, y float32) lmath.Vec2 {
	return lmath.Vec2{X: acceptanceTileSize * x, Y: acceptanceTileSize * y}
}

func spriteScaleForSrc(x, y float32, src lmath.Rect) lmath.Vec2 {
	return lmath.Vec2{X: acceptanceTileSize * x / src.W, Y: acceptanceTileSize * y / src.H}
}

func solidSpriteSrc() lmath.Rect {
	return lmath.Rect{X: 0.5, Y: 0.5, W: 1.0 / 64, H: 1.0 / 64}
}

func floorTileColor(x, y int) lmath.Color {
	base := lmath.Color{R: 0.43, G: 0.50, B: 0.47, A: 1}
	if (x+y)%2 == 0 {
		base = lmath.Color{R: 0.51, G: 0.58, B: 0.54, A: 1}
	}
	if x%5 == 0 || y%5 == 0 {
		base = lmath.Color{R: 0.35, G: 0.40, B: 0.39, A: 1}
	}
	return addJitter(base, x, y, 0.016)
}

func wallTileScale(x, y int) lmath.Vec2 {
	if x <= 1 || x >= acceptanceTileColumns-2 {
		return spriteScale(1.18, 1.05)
	}
	if y <= 1 || y >= acceptanceTileRows-2 {
		return spriteScale(1.05, 1.18)
	}
	return spriteScale(1.16, 1.16)
}

func wallTileColor(x, y int) lmath.Color {
	if x <= 1 || y <= 1 || x >= acceptanceTileColumns-2 || y >= acceptanceTileRows-2 {
		return addJitter(lmath.Color{R: 0.24, G: 0.30, B: 0.36, A: 1}, x, y, 0.012)
	}
	return addJitter(lmath.Color{R: 0.18, G: 0.25, B: 0.33, A: 1}, x, y, 0.014)
}

func propTileColor(x, y int) (lmath.Color, bool) {
	switch {
	case inRectTile(x, y, 7, 8, 6, 3):
		return propBlockColor(x, y, 7, 8, 6, 3, lmath.Color{R: 0.84, G: 0.60, B: 0.33, A: 1}), true
	case inRectTile(x, y, 27, 16, 6, 4):
		return propBlockColor(x, y, 27, 16, 6, 4, lmath.Color{R: 0.70, G: 0.46, B: 0.30, A: 1}), true
	case inRectTile(x, y, 18, 10, 4, 5):
		return propBlockColor(x, y, 18, 10, 4, 5, lmath.Color{R: 0.50, G: 0.76, B: 0.60, A: 1}), true
	case inRectTile(x, y, 4, 18, 5, 3):
		return propBlockColor(x, y, 4, 18, 5, 3, lmath.Color{R: 0.80, G: 0.56, B: 0.36, A: 1}), true
	case inRectTile(x, y, 14, 15, 6, 4):
		return propBlockColor(x, y, 14, 15, 6, 4, lmath.Color{R: 0.40, G: 0.72, B: 0.58, A: 1}), true
	case inRectTile(x, y, 24, 7, 5, 3):
		return propBlockColor(x, y, 24, 7, 5, 3, lmath.Color{R: 0.42, G: 0.68, B: 0.76, A: 1}), true
	case ((x+y)%11 == 0 && x > 4 && x < acceptanceTileColumns-5 && y > 4 && y < acceptanceTileRows-5):
		return addJitter(lmath.Color{R: 0.38, G: 0.64, B: 0.50, A: 1}, x, y, 0.012), true
	default:
		return lmath.Color{}, false
	}
}

func propBlockColor(x, y, left, top, width, height int, base lmath.Color) lmath.Color {
	if x == left || x == left+width-1 || y == top || y == top+height-1 {
		return addJitter(lmath.Color{R: base.R * 0.58, G: base.G * 0.58, B: base.B * 0.58, A: 1}, x, y, 0.010)
	}
	if y == top+1 {
		return addJitter(lmath.Color{R: min1(base.R + 0.13), G: min1(base.G + 0.13), B: min1(base.B + 0.13), A: 1}, x, y, 0.012)
	}
	return addJitter(base, x, y, 0.012)
}

func propMaterialIndex(x, y int) int {
	if inRectTile(x, y, 18, 10, 4, 5) || inRectTile(x, y, 14, 15, 6, 4) || inRectTile(x, y, 24, 7, 5, 3) {
		return 3
	}
	return 2
}

func propTileRotation(x, y int) float32 {
	if (x+y)%11 == 0 {
		return 0.18
	}
	return 0
}

type lightMarkerSpec struct {
	Color lmath.Color
	Scale lmath.Vec2
}

func lightMarkerTileColor(x, y int) (lightMarkerSpec, bool) {
	markers := []struct {
		X, Y  int
		Color lmath.Color
	}{
		{X: 9, Y: 9, Color: lmath.Color{R: 1.00, G: 0.78, B: 0.30, A: 1}},
		{X: 32, Y: 9, Color: lmath.Color{R: 0.35, G: 0.56, B: 1.00, A: 1}},
		{X: 19, Y: 19, Color: lmath.Color{R: 0.86, G: 0.38, B: 1.00, A: 1}},
		{X: 35, Y: 18, Color: lmath.Color{R: 0.42, G: 1.00, B: 0.60, A: 1}},
	}
	for _, marker := range markers {
		dx := absInt(x - marker.X)
		dy := absInt(y - marker.Y)
		if dx == 0 && dy == 0 {
			return lightMarkerSpec{
				Color: lmath.Color{R: min1(marker.Color.R + 0.20), G: min1(marker.Color.G + 0.20), B: min1(marker.Color.B + 0.20), A: 1},
				Scale: spriteScaleForSrc(0.72, 0.72, solidSpriteSrc()),
			}, true
		}
		if dx+dy == 1 {
			return lightMarkerSpec{
				Color: marker.Color,
				Scale: spriteScaleForSrc(0.42, 0.42, solidSpriteSrc()),
			}, true
		}
		if dx == 1 && dy == 1 {
			return lightMarkerSpec{
				Color: lmath.Color{R: marker.Color.R * 0.55, G: marker.Color.G * 0.55, B: marker.Color.B * 0.55, A: 1},
				Scale: spriteScaleForSrc(0.28, 0.28, solidSpriteSrc()),
			}, true
		}
	}
	return lightMarkerSpec{}, false
}

func legendTileColor(x, y int) (lmath.Color, bool) {
	if y < 20 || y > 22 || x < 22 || x > 37 {
		return lmath.Color{}, false
	}
	switch {
	case inRectTile(x, y, 22, 20, 2, 2):
		return lmath.Color{R: 0.48, G: 0.56, B: 0.52, A: 1}, true
	case inRectTile(x, y, 26, 20, 2, 2):
		return lmath.Color{R: 0.20, G: 0.28, B: 0.36, A: 1}, true
	case inRectTile(x, y, 30, 20, 2, 2):
		return lmath.Color{R: 0.88, G: 0.62, B: 0.34, A: 1}, true
	case inRectTile(x, y, 34, 20, 2, 2):
		return lmath.Color{R: 1.00, G: 0.82, B: 0.22, A: 1}, true
	case x == 24 && y == 22:
		return lmath.Color{R: 0.74, G: 0.84, B: 0.78, A: 1}, true
	case x == 28 && y == 22:
		return lmath.Color{R: 0.56, G: 0.68, B: 0.82, A: 1}, true
	case x == 32 && y == 22:
		return lmath.Color{R: 0.98, G: 0.74, B: 0.42, A: 1}, true
	case x == 36 && y == 22:
		return lmath.Color{R: 1.00, G: 0.96, B: 0.62, A: 1}, true
	default:
		return lmath.Color{}, false
	}
}

type labelTileSpec struct {
	Color lmath.Color
}

func labelTileAt(x, y int) (labelTileSpec, bool) {
	labels := []struct {
		Text  string
		X, Y  int
		Color lmath.Color
	}{
		{Text: "WALL", X: 3, Y: 2, Color: lmath.Color{R: 0.82, G: 0.92, B: 1.00, A: 1}},
		{Text: "LIGHT", X: 20, Y: 2, Color: lmath.Color{R: 1.00, G: 0.92, B: 0.32, A: 1}},
		{Text: "PROP", X: 15, Y: 10, Color: lmath.Color{R: 1.00, G: 0.72, B: 0.38, A: 1}},
		{Text: "FLOOR", X: 3, Y: 19, Color: lmath.Color{R: 0.78, G: 0.92, B: 0.80, A: 1}},
	}
	for _, label := range labels {
		if labelPixel(x, y, label.X, label.Y, label.Text) {
			return labelTileSpec{Color: label.Color}, true
		}
	}
	return labelTileSpec{}, false
}

func labelPixel(x, y, left, top int, text string) bool {
	localX := x - left
	localY := y - top
	if localY < 0 || localY >= 5 || localX < 0 {
		return false
	}
	charIndex := localX / 4
	charColumn := localX % 4
	if charIndex >= len(text) || charColumn == 3 {
		return false
	}
	return glyphPixel(text[charIndex], charColumn, localY)
}

func glyphPixel(char byte, x, y int) bool {
	rows := glyphRows(char)
	if rows == [5]string{} {
		return false
	}
	return rows[y][x] == '1'
}

func glyphRows(char byte) [5]string {
	switch char {
	case 'A':
		return [5]string{"010", "101", "111", "101", "101"}
	case 'F':
		return [5]string{"111", "100", "110", "100", "100"}
	case 'G':
		return [5]string{"111", "100", "101", "101", "111"}
	case 'H':
		return [5]string{"101", "101", "111", "101", "101"}
	case 'I':
		return [5]string{"111", "010", "010", "010", "111"}
	case 'L':
		return [5]string{"100", "100", "100", "100", "111"}
	case 'O':
		return [5]string{"111", "101", "101", "101", "111"}
	case 'P':
		return [5]string{"111", "101", "111", "100", "100"}
	case 'R':
		return [5]string{"111", "101", "111", "110", "101"}
	case 'T':
		return [5]string{"111", "010", "010", "010", "010"}
	case 'W':
		return [5]string{"101", "101", "101", "111", "101"}
	default:
		return [5]string{}
	}
}

func addJitter(color lmath.Color, x, y int, amount float32) lmath.Color {
	jitter := tileJitter(x, y) * amount / 0.008
	return lmath.Color{
		R: min1(color.R + jitter),
		G: min1(color.G + jitter),
		B: min1(color.B + jitter),
		A: color.A,
	}
}

func tileCenterX(x int) float32 {
	return acceptanceOriginX + float32(x*acceptanceTileSpacingX)
}

func tileCenterY(y int) float32 {
	return acceptanceOriginY + float32(y*acceptanceTileSpacingY)
}

func roomX(x float32) float32 {
	return acceptanceOriginX + (x-authoredOriginX)*float32(acceptanceTileSpacingX)/authoredTileSpacing
}

func roomY(y float32) float32 {
	return acceptanceOriginY + (y-authoredOriginY)*float32(acceptanceTileSpacingY)/authoredTileSpacing
}

func roomRect(rect lmath.Rect) lmath.Rect {
	return lmath.Rect{
		X: roomX(rect.X),
		Y: roomY(rect.Y),
		W: rect.W * float32(acceptanceTileSpacingX) / authoredTileSpacing,
		H: rect.H * float32(acceptanceTileSpacingY) / authoredTileSpacing,
	}
}

func roomPoint(point lmath.Vec2) lmath.Vec2 {
	return lmath.Vec2{X: roomX(point.X), Y: roomY(point.Y)}
}

func material(game *app.Game, name string, roughness, emissive float32) graphics.Material2D {
	return graphics.Material2D{
		Albedo:    game.Assets.LoadTexture(lightingRoomAssetPath(name + ".png")),
		Normal:    game.Assets.LoadTexture(lightingRoomAssetPath(name + "_n.png")),
		Roughness: roughness,
		Emissive:  emissive,
	}
}

func lightingRoomAssetPath(name string) string {
	path := filepath.Join("examples", "lighting_room", "assets", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join("assets", name)
}

func addSprite(world *scene.Scene, material graphics.Material2D, spec roomSpriteSpec) {
	world.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Material: material,
			Src:      spec.Src,
			Color:    spec.Color,
		},
		Transform: graphics.Transform2D{
			Position: spec.Position,
			Scale:    spec.Scale,
			Rotation: spec.Rotation,
			Z:        float32(spec.Layer),
		},
		Layer: spec.Layer,
	})
}

func light(x, y, radius float32, color lmath.Color, intensity float32) graphics.Light2D {
	return graphics.Light2D{
		Position:  lmath.Vec2{X: x, Y: y},
		Radius:    radius,
		Color:     color,
		Intensity: intensity,
		Falloff:   1.4,
	}
}

func shadowLight(x, y, radius float32, color lmath.Color, intensity float32) graphics.Light2D {
	light := light(x, y, radius, color, intensity)
	light.CastShadows = true
	return light
}

func updateLights(world *scene.Scene, t float32) {
	world.SetLights([]graphics.Light2D{
		shadowLight(roomX(610+90*float32(math.Sin(float64(t*1.1)))), roomY(360+65*float32(math.Cos(float64(t*0.8)))), roomRadius(430), lmath.Color{R: 1.00, G: 0.76, B: 0.46, A: 1}, 1.9),
		shadowLight(roomX(1300+120*float32(math.Cos(float64(t*0.9)))), roomY(430+75*float32(math.Sin(float64(t*1.4)))), roomRadius(460), lmath.Color{R: 0.44, G: 0.66, B: 1.00, A: 1}, 1.5),
		light(roomX(880+90*float32(math.Sin(float64(t*1.7)))), roomY(790+55*float32(math.Sin(float64(t*0.7)))), roomRadius(390), lmath.Color{R: 0.84, G: 0.44, B: 1.00, A: 1}, 1.1),
		light(roomX(1510+95*float32(math.Cos(float64(t*0.6)))), roomY(760+70*float32(math.Sin(float64(t*1.2)))), roomRadius(420), lmath.Color{R: 0.54, G: 1.00, B: 0.70, A: 1}, 1.2),
	})
}

func roomRadius(radius float32) float32 {
	scaleX := float32(acceptanceTileSpacingX) / authoredTileSpacing
	scaleY := float32(acceptanceTileSpacingY) / authoredTileSpacing
	return radius * (scaleX + scaleY) * 0.5
}

func addAcceptanceOccluders(world *scene.Scene) {
	rects := []lmath.Rect{
		{X: 280, Y: 114, W: 1360, H: 22},
		{X: 280, Y: 948, W: 1360, H: 22},
		{X: 280, Y: 114, W: 22, H: 856},
		{X: 1618, Y: 114, W: 22, H: 856},
		{X: 470, Y: 264, W: 360, H: 22},
		{X: 1120, Y: 264, W: 370, H: 22},
		{X: 688, Y: 286, W: 22, H: 270},
		{X: 1232, Y: 522, W: 22, H: 306},
		{X: 520, Y: 366, W: 200, H: 92},
		{X: 1210, Y: 674, W: 230, H: 126},
		{X: 910, Y: 468, W: 128, H: 188},
		{X: 430, Y: 760, W: 190, H: 90},
		{X: 750, Y: 676, W: 250, H: 22},
		{X: 1340, Y: 330, W: 220, H: 22},
		{X: 620, Y: 620, W: 96, H: 96},
		{X: 1440, Y: 760, W: 96, H: 96},
	}
	for i, rect := range rects {
		occluder := graphics.RectOccluder2D(roomRect(rect), 1)
		if i >= 14 {
			occluder.Caster = graphics.ShadowCaster2D{ID: "dynamic", Dynamic: true}
		}
		world.AddOccluder(occluder)
	}

	segments := []graphics.Segment2D{
		{A: lmath.Vec2{X: 382, Y: 308}, B: lmath.Vec2{X: 510, Y: 438}},
		{A: lmath.Vec2{X: 760, Y: 380}, B: lmath.Vec2{X: 910, Y: 520}},
		{A: lmath.Vec2{X: 1030, Y: 800}, B: lmath.Vec2{X: 1210, Y: 924}},
		{A: lmath.Vec2{X: 1450, Y: 540}, B: lmath.Vec2{X: 1585, Y: 700}},
	}
	for _, segment := range segments {
		a := roomPoint(segment.A)
		b := roomPoint(segment.B)
		world.AddOccluder(graphics.SegmentOccluder2D(a, b, 1))
	}
}

func min1(value float32) float32 {
	if value > 1 {
		return 1
	}
	return value
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
