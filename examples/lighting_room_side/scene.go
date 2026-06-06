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
	sideTargetWidth      = 1920
	sideTargetHeight     = 1080
	sideTargetFPS        = 60
	sideTileSize         = 36
	sideTileColumns      = 42
	sideTileRows         = 24
	sideOriginX          = 204
	sideOriginY          = 120
	sideMaterialCount    = 8
	sideLightCount       = 5
	sideShadowLightCount = 3
	sideOccluderCount    = 18
)

type sideSpriteRole int

const (
	sideRoleBackground sideSpriteRole = iota
	sideRoleWall
	sideRoleFloor
	sideRolePlatform
	sideRoleProp
	sideRoleLamp
	sideRoleCharacter
	sideRoleOverlay
)

type sideMaterialSet struct {
	Background graphics.Material2D
	Wall       graphics.Material2D
	Floor      graphics.Material2D
	Platform   graphics.Material2D
	Crate      graphics.Material2D
	Lamp       graphics.Material2D
	Character  graphics.Material2D
	Overlay    graphics.Material2D
}

type sideSpriteSpec struct {
	Role     sideSpriteRole
	Material graphics.Material2D
	Src      lmath.Rect
	Position lmath.Vec2
	Scale    lmath.Vec2
	Rotation float32
	Color    lmath.Color
	Layer    int
}

func buildSideLightingRoom(game *app.Game, config demoConfig) *scene.Scene {
	world := scene.New()
	world.SetLightingConfig(graphics.LightingConfig2D{
		Ambient:    lmath.Color{R: 0.24, G: 0.25, B: 0.28, A: 1},
		DebugView:  config.DebugView,
		ShadowMode: config.ShadowMode,
	})
	materials := sideMaterials(game)
	addSideSprites(world, materials)
	addSideOccluders(world)
	updateSideLights(world, 0)
	return world
}

func sideMaterials(game *app.Game) sideMaterialSet {
	return sideMaterialSet{
		Background: sideMaterial(game, "background", 0.95, 0),
		Wall:       sideMaterial(game, "wall", 0.62, 0.16),
		Floor:      sideMaterial(game, "floor", 0.50, 0.28),
		Platform:   sideMaterial(game, "platform", 0.45, 0.34),
		Crate:      sideMaterial(game, "crate", 0.42, 0.58),
		Lamp:       sideMaterial(game, "lamp", 0.20, 3.0),
		Character:  sideMaterial(game, "character", 0.36, 0.78),
		Overlay:    sideMaterial(game, "overlay", 0.18, 4.0),
	}
}

func addSideSprites(world *scene.Scene, materials sideMaterialSet) {
	for y := 2; y < sideTileRows-2; y += 2 {
		for x := 2; x < sideTileColumns-2; x += 2 {
			addSideSprite(world, sideSpriteSpec{
				Role:     sideRoleBackground,
				Material: materials.Background,
				Src:      fullSpriteSrc(),
				Position: sideTileCenter(x, y),
				Scale:    sideSpriteScale(2.1, 2.1),
				Color:    sideBackgroundColor(x, y),
				Layer:    0,
			})
		}
	}
	addSideSolid(world, materials.Overlay, sideRoleOverlay, sideTileCenter(21, 12), 36.5, 16.5, lmath.Color{R: 0.06, G: 0.07, B: 0.09, A: 1}, 1)
	for x := 0; x < sideTileColumns; x++ {
		for _, y := range []int{0, 1, sideTileRows - 2, sideTileRows - 1} {
			addSideTile(world, materials.Wall, sideRoleWall, x, y, sideWallColor(x, y), 2)
		}
	}
	for y := 2; y < sideTileRows-2; y++ {
		for _, x := range []int{0, 1, sideTileColumns - 2, sideTileColumns - 1} {
			addSideTile(world, materials.Wall, sideRoleWall, x, y, sideWallColor(x, y), 2)
		}
	}
	for y := sideTileRows - 5; y < sideTileRows-2; y++ {
		for x := 2; x < sideTileColumns-2; x++ {
			addSideTile(world, materials.Floor, sideRoleFloor, x, y, sideFloorColor(x, y), 3)
		}
	}
	addPlatformRun(world, materials, 6, 15, 11)
	addPlatformRun(world, materials, 20, 11, 12)
	addPlatformRun(world, materials, 28, 17, 9)
	addPlatformRun(world, materials, 10, 8, 8)
	addPropStack(world, materials, 6, sideTileRows-7, 3)
	addPropStack(world, materials, 33, sideTileRows-7, 4)
	addPropStack(world, materials, 25, 8, 2)
	addPropStack(world, materials, 13, 13, 2)
	addLamps(world, materials)
	addCharacters(world, materials)
	addOverlayLegend(world, materials)
}

func addSideTile(world *scene.Scene, material graphics.Material2D, role sideSpriteRole, x, y int, color lmath.Color, layer int) {
	addSideSprite(world, sideSpriteSpec{
		Role:     role,
		Material: material,
		Src:      fullSpriteSrc(),
		Position: sideTileCenter(x, y),
		Scale:    sideSpriteScale(1, 1),
		Color:    color,
		Layer:    layer,
	})
}

func addPlatformRun(world *scene.Scene, materials sideMaterialSet, left, y, width int) {
	addSideSolid(world, materials.Overlay, sideRoleOverlay, sideTileCenter(left+width/2, y), float32(width)+0.35, 1.15, lmath.Color{R: 0.07, G: 0.06, B: 0.05, A: 1}, 3)
	for x := left; x < left+width; x++ {
		addSideSprite(world, sideSpriteSpec{
			Role:     sideRolePlatform,
			Material: materials.Platform,
			Src:      fullSpriteSrc(),
			Position: sideTileCenter(x, y),
			Scale:    sideSpriteScale(1.05, 0.75),
			Color:    lmath.Color{R: 0.92, G: 0.82, B: 0.68, A: 1},
			Layer:    4,
		})
	}
}

func addPropStack(world *scene.Scene, materials sideMaterialSet, x, bottomY, height int) {
	for i := 0; i < height; i++ {
		addSideSolid(world, materials.Overlay, sideRoleOverlay, sideTileCenter(x, bottomY-i), 1.35, 1.35, lmath.Color{R: 0.06, G: 0.04, B: 0.03, A: 1}, 4)
		addSideSprite(world, sideSpriteSpec{
			Role:     sideRoleProp,
			Material: materials.Crate,
			Src:      fullSpriteSrc(),
			Position: sideTileCenter(x, bottomY-i),
			Scale:    sideSpriteScale(1.55, 1.55),
			Rotation: float32((i%2)*2-1) * 0.05,
			Color:    lmath.Color{R: 1.00, G: 0.88, B: 0.70, A: 1},
			Layer:    5,
		})
	}
}

func addLamps(world *scene.Scene, materials sideMaterialSet) {
	for _, lamp := range sideLampPositions() {
		addSideSolid(world, materials.Overlay, sideRoleOverlay, sideTileCenter(lamp.X, lamp.Y), 1.75, 0.28, lamp.Color, 6)
		addSideSolid(world, materials.Overlay, sideRoleOverlay, sideTileCenter(lamp.X, lamp.Y), 0.28, 1.75, lamp.Color, 6)
		addSideSprite(world, sideSpriteSpec{
			Role:     sideRoleLamp,
			Material: materials.Lamp,
			Src:      fullSpriteSrc(),
			Position: sideTileCenter(lamp.X, lamp.Y),
			Scale:    sideSpriteScale(1.75, 1.75),
			Color:    lmath.White(),
			Layer:    7,
		})
		addSideSprite(world, sideSpriteSpec{
			Role:     sideRoleOverlay,
			Material: materials.Overlay,
			Src:      solidSpriteSrc(),
			Position: sideTileCenter(lamp.X, lamp.Y+1),
			Scale:    sideSpriteScaleForSrc(0.38, 0.38, solidSpriteSrc()),
			Color:    lamp.Color,
			Layer:    8,
		})
	}
}

func addCharacters(world *scene.Scene, materials sideMaterialSet) {
	for _, character := range []struct {
		X, Y int
		C    lmath.Color
	}{
		{X: 14, Y: sideTileRows - 6, C: lmath.Color{R: 0.75, G: 0.95, B: 1.00, A: 1}},
		{X: 23, Y: 10, C: lmath.Color{R: 0.95, G: 0.78, B: 1.00, A: 1}},
	} {
		addSideSolid(world, materials.Overlay, sideRoleOverlay, sideTileCenter(character.X, character.Y), 1.55, 1.85, lmath.Color{R: 0.03, G: 0.04, B: 0.06, A: 1}, 8)
		addSideSprite(world, sideSpriteSpec{
			Role:     sideRoleCharacter,
			Material: materials.Character,
			Src:      fullSpriteSrc(),
			Position: sideTileCenter(character.X, character.Y),
			Scale:    sideSpriteScale(1.95, 1.95),
			Color:    character.C,
			Layer:    9,
		})
	}
}

func addOverlayLegend(world *scene.Scene, materials sideMaterialSet) {
	swatches := []struct {
		X, Y int
		C    lmath.Color
	}{
		{X: 4, Y: 3, C: lmath.Color{R: 0.25, G: 0.45, B: 0.75, A: 1}},
		{X: 6, Y: 3, C: lmath.Color{R: 0.92, G: 0.82, B: 0.44, A: 1}},
		{X: 8, Y: 3, C: lmath.Color{R: 0.90, G: 0.60, B: 0.34, A: 1}},
		{X: 10, Y: 3, C: lmath.Color{R: 0.40, G: 0.85, B: 1.00, A: 1}},
	}
	for _, swatch := range swatches {
		addSideSprite(world, sideSpriteSpec{
			Role:     sideRoleOverlay,
			Material: materials.Overlay,
			Src:      solidSpriteSrc(),
			Position: sideTileCenter(swatch.X, swatch.Y),
			Scale:    sideSpriteScaleForSrc(0.8, 0.8, solidSpriteSrc()),
			Color:    swatch.C,
			Layer:    10,
		})
	}
}

type sideLamp struct {
	X, Y  int
	Color lmath.Color
}

func sideLampPositions() []sideLamp {
	return []sideLamp{
		{X: 8, Y: 4, Color: lmath.Color{R: 1.00, G: 0.72, B: 0.26, A: 1}},
		{X: 18, Y: 5, Color: lmath.Color{R: 0.62, G: 0.88, B: 1.00, A: 1}},
		{X: 31, Y: 6, Color: lmath.Color{R: 0.92, G: 0.44, B: 1.00, A: 1}},
		{X: 13, Y: 14, Color: lmath.Color{R: 1.00, G: 0.88, B: 0.52, A: 1}},
		{X: 35, Y: 15, Color: lmath.Color{R: 0.52, G: 1.00, B: 0.62, A: 1}},
	}
}

func updateSideLights(world *scene.Scene, t float32) {
	world.SetLights([]graphics.Light2D{
		sideShadowLight(8, 4, 250, lmath.Color{R: 1.00, G: 0.72, B: 0.26, A: 1}, 1.25),
		sideLight(18+int(2*float32(math.Sin(float64(t*1.2)))), 5, 240, lmath.Color{R: 0.62, G: 0.88, B: 1.00, A: 1}, 1.15),
		sideShadowLight(31, 6, 270, lmath.Color{R: 0.92, G: 0.44, B: 1.00, A: 1}, 1.10),
		sideLight(13, 14, 230, lmath.Color{R: 1.00, G: 0.88, B: 0.52, A: 1}, 1.05),
		sideShadowLight(35, 15, 240, lmath.Color{R: 0.52, G: 1.00, B: 0.62, A: 1}, 1.10),
	})
}

func sideLight(x, y int, radius float32, color lmath.Color, intensity float32) graphics.Light2D {
	return graphics.Light2D{
		Position:  sideTileCenter(x, y),
		Radius:    radius,
		Color:     color,
		Intensity: intensity,
		Falloff:   1.5,
	}
}

func sideShadowLight(x, y int, radius float32, color lmath.Color, intensity float32) graphics.Light2D {
	light := sideLight(x, y, radius, color, intensity)
	light.CastShadows = true
	return light
}

func addSideOccluders(world *scene.Scene) {
	rects := []lmath.Rect{
		sideTileRect(0, 0, sideTileColumns, 2),
		sideTileRect(0, sideTileRows-2, sideTileColumns, 2),
		sideTileRect(0, 0, 2, sideTileRows),
		sideTileRect(sideTileColumns-2, 0, 2, sideTileRows),
		sideTileRect(2, sideTileRows-5, sideTileColumns-4, 3),
		sideTileRect(6, 15, 11, 1),
		sideTileRect(20, 11, 12, 1),
		sideTileRect(28, 17, 9, 1),
		sideTileRect(10, 8, 8, 1),
		sideTileRect(6, sideTileRows-9, 1, 3),
		sideTileRect(33, sideTileRows-10, 1, 4),
		sideTileRect(25, 7, 1, 2),
		sideTileRect(13, 12, 1, 2),
		sideTileRect(22, 13, 3, 1),
	}
	for i, rect := range rects {
		occluder := graphics.RectOccluder2D(rect, 1)
		if i >= 9 {
			occluder.Caster = graphics.ShadowCaster2D{ID: "side-prop", Dynamic: true}
		}
		world.AddOccluder(occluder)
	}
	for _, segment := range []graphics.Segment2D{
		{A: sideTileCenter(5, sideTileRows-6), B: sideTileCenter(9, sideTileRows-10)},
		{A: sideTileCenter(18, 8), B: sideTileCenter(23, 11)},
		{A: sideTileCenter(29, 15), B: sideTileCenter(35, 12)},
		{A: sideTileCenter(37, 6), B: sideTileCenter(39, 12)},
	} {
		world.AddOccluder(graphics.SegmentOccluder2D(segment.A, segment.B, 1))
	}
}

func sideTileRect(left, top, width, height int) lmath.Rect {
	return lmath.Rect{
		X: float32(sideOriginX + left*sideTileSize - sideTileSize/2),
		Y: float32(sideOriginY + top*sideTileSize - sideTileSize/2),
		W: float32(width * sideTileSize),
		H: float32(height * sideTileSize),
	}
}

func sideTileCenter(x, y int) lmath.Vec2 {
	return lmath.Vec2{
		X: float32(sideOriginX + x*sideTileSize),
		Y: float32(sideOriginY + y*sideTileSize),
	}
}

func sideSpriteScale(x, y float32) lmath.Vec2 {
	return lmath.Vec2{X: sideTileSize * x, Y: sideTileSize * y}
}

func sideSpriteScaleForSrc(x, y float32, src lmath.Rect) lmath.Vec2 {
	return lmath.Vec2{X: sideTileSize * x / src.W, Y: sideTileSize * y / src.H}
}

func addSideSolid(world *scene.Scene, material graphics.Material2D, role sideSpriteRole, position lmath.Vec2, widthTiles, heightTiles float32, color lmath.Color, layer int) {
	addSideSprite(world, sideSpriteSpec{
		Role:     role,
		Material: material,
		Src:      solidSpriteSrc(),
		Position: position,
		Scale:    sideSpriteScaleForSrc(widthTiles, heightTiles, solidSpriteSrc()),
		Color:    color,
		Layer:    layer,
	})
}

func sideBackgroundColor(x, y int) lmath.Color {
	base := lmath.Color{R: 0.24, G: 0.30, B: 0.38, A: 1}
	if y > 12 {
		base = lmath.Color{R: 0.20, G: 0.24, B: 0.30, A: 1}
	}
	return sideJitter(base, x, y, 0.006)
}

func sideWallColor(x, y int) lmath.Color {
	return sideJitter(lmath.Color{R: 0.62, G: 0.68, B: 0.78, A: 1}, x, y, 0.010)
}

func sideFloorColor(x, y int) lmath.Color {
	if y == sideTileRows-5 {
		return sideJitter(lmath.Color{R: 0.96, G: 0.90, B: 0.64, A: 1}, x, y, 0.012)
	}
	return sideJitter(lmath.Color{R: 0.72, G: 0.56, B: 0.42, A: 1}, x, y, 0.010)
}

func sideJitter(color lmath.Color, x, y int, amount float32) lmath.Color {
	jitter := float32((x*7+y*5)%5) * amount
	return lmath.Color{
		R: min1(color.R + jitter),
		G: min1(color.G + jitter),
		B: min1(color.B + jitter),
		A: color.A,
	}
}

func sideMaterial(game *app.Game, name string, roughness, emissive float32) graphics.Material2D {
	return graphics.Material2D{
		Albedo:    game.Assets.LoadTexture(sideAssetPath(name + ".png")),
		Normal:    game.Assets.LoadTexture(sideAssetPath(name + "_n.png")),
		Roughness: roughness,
		Emissive:  emissive,
	}
}

func sideAssetPath(name string) string {
	path := filepath.Join("examples", "lighting_room_side", "assets", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join("assets", name)
}

func addSideSprite(world *scene.Scene, spec sideSpriteSpec) {
	world.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Material: spec.Material,
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

func fullSpriteSrc() lmath.Rect {
	return lmath.Rect{W: 1, H: 1}
}

func solidSpriteSrc() lmath.Rect {
	return lmath.Rect{X: 0.5, Y: 0.5, W: 1.0 / 64, H: 1.0 / 64}
}

func min1(value float32) float32 {
	if value > 1 {
		return 1
	}
	return value
}
