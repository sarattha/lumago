package main

import (
	"math"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	"github.com/sarattha/lumago/engine/scene"
)

const (
	acceptanceSprites      = 1000
	acceptanceMaterials    = 4
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
	}
	addAcceptanceSprites(world, materials)
	addAcceptanceOccluders(world)
	updateLights(world, 0)
	return world
}

func addAcceptanceSprites(world *scene.Scene, materials []graphics.Material2D) {
	palette := []lmath.Color{
		{R: 0.54, G: 0.62, B: 0.58, A: 1}, // floor
		{R: 0.36, G: 0.43, B: 0.47, A: 1}, // wall
		{R: 0.86, G: 0.66, B: 0.42, A: 1}, // warm props
		{R: 0.42, G: 0.74, B: 0.58, A: 1}, // cool props/accent floor
	}
	for y := 0; y < acceptanceTileRows; y++ {
		for x := 0; x < acceptanceTileColumns; x++ {
			materialIndex := roomMaterialIndex(x, y)
			jitter := tileJitter(x, y)
			addSprite(
				world,
				materials[materialIndex],
				lmath.Rect{W: acceptanceTileSize, H: acceptanceTileSize},
				lmath.Vec2{
					X: tileCenterX(x),
					Y: tileCenterY(y),
				},
				lmath.Color{
					R: min1(palette[materialIndex].R + jitter),
					G: min1(palette[materialIndex].G + jitter),
					B: min1(palette[materialIndex].B + jitter),
					A: 1,
				},
				materialIndex,
			)
		}
	}
}

func roomMaterialIndex(x, y int) int {
	if isWallTile(x, y) {
		return 1
	}
	if inRectTile(x, y, 7, 7, 6, 3) ||
		inRectTile(x, y, 27, 16, 6, 4) ||
		inRectTile(x, y, 18, 10, 4, 5) ||
		inRectTile(x, y, 4, 18, 5, 3) {
		return 2
	}
	if inRectTile(x, y, 13, 15, 7, 4) ||
		inRectTile(x, y, 24, 6, 5, 3) ||
		((x+y)%9 == 0 && x > 3 && x < acceptanceTileColumns-4 && y > 3 && y < acceptanceTileRows-3) {
		return 3
	}
	return 0
}

func isWallTile(x, y int) bool {
	return x == 0 || y == 0 || x == acceptanceTileColumns-1 || y == acceptanceTileRows-1 ||
		(y == 4 && x > 4 && x < 16) ||
		(y == 4 && x > 24 && x < 36) ||
		(x == 12 && y > 4 && y < 13) ||
		(x == 28 && y > 11 && y < 20)
}

func inRectTile(x, y, left, top, width, height int) bool {
	return x >= left && x < left+width && y >= top && y < top+height
}

func tileJitter(x, y int) float32 {
	return float32((x*3+y*5)%6) * 0.008
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
		Albedo:    game.Assets.LoadTexture("examples/lighting_room/assets/" + name + ".png"),
		Normal:    game.Assets.LoadTexture("examples/lighting_room/assets/" + name + "_n.png"),
		Roughness: roughness,
		Emissive:  emissive,
	}
}

func addSprite(world *scene.Scene, material graphics.Material2D, src lmath.Rect, position lmath.Vec2, color lmath.Color, layer int) {
	world.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Material: material,
			Src:      src,
			Color:    color,
		},
		Transform: graphics.Transform2D{
			Position: position,
			Scale:    lmath.Vec2{X: 1, Y: 1},
			Z:        float32(layer),
		},
		Layer: layer,
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
