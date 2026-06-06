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
	acceptanceTileSize     = 32
	acceptanceTileSpacing  = 38
	acceptanceOriginX      = 220
	acceptanceOriginY      = 120
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
		{R: 0.72, G: 0.76, B: 0.68, A: 1},
		{R: 0.54, G: 0.58, B: 0.64, A: 1},
		{R: 0.93, G: 0.78, B: 0.58, A: 1},
		{R: 0.48, G: 0.82, B: 0.75, A: 1},
	}
	for y := 0; y < acceptanceTileRows; y++ {
		for x := 0; x < acceptanceTileColumns; x++ {
			index := y*acceptanceTileColumns + x
			materialIndex := index % len(materials)
			jitter := float32((x*y)%7) * 0.01
			addSprite(
				world,
				materials[materialIndex],
				lmath.Rect{W: acceptanceTileSize, H: acceptanceTileSize},
				lmath.Vec2{
					X: acceptanceOriginX + float32(x*acceptanceTileSpacing),
					Y: acceptanceOriginY + float32(y*acceptanceTileSpacing),
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
		shadowLight(560+130*float32(math.Sin(float64(t*1.1))), 260+90*float32(math.Cos(float64(t*0.8))), 460, lmath.Color{R: 1.00, G: 0.78, B: 0.45, A: 1}, 1.9),
		shadowLight(1210+160*float32(math.Cos(float64(t*0.9))), 360+95*float32(math.Sin(float64(t*1.4))), 420, lmath.Color{R: 0.45, G: 0.68, B: 1.00, A: 1}, 1.4),
		light(820+120*float32(math.Sin(float64(t*1.7))), 790+70*float32(math.Sin(float64(t*0.7))), 360, lmath.Color{R: 0.85, G: 0.42, B: 1.00, A: 1}, 1.1),
		light(1480+150*float32(math.Cos(float64(t*0.6))), 780+90*float32(math.Sin(float64(t*1.2))), 500, lmath.Color{R: 0.55, G: 1.00, B: 0.70, A: 1}, 1.2),
	})
}

func addAcceptanceOccluders(world *scene.Scene) {
	rects := []lmath.Rect{
		{X: 420, Y: 210, W: 260, H: 20},
		{X: 420, Y: 420, W: 260, H: 20},
		{X: 420, Y: 210, W: 20, H: 230},
		{X: 660, Y: 210, W: 20, H: 230},
		{X: 900, Y: 260, W: 260, H: 22},
		{X: 900, Y: 520, W: 260, H: 22},
		{X: 900, Y: 260, W: 22, H: 282},
		{X: 1138, Y: 260, W: 22, H: 282},
		{X: 1260, Y: 700, W: 220, H: 22},
		{X: 500, Y: 740, W: 240, H: 20},
		{X: 760, Y: 680, W: 22, H: 150},
		{X: 1520, Y: 340, W: 24, H: 190},
		{X: 300, Y: 600, W: 180, H: 18},
		{X: 1420, Y: 160, W: 160, H: 18},
		{X: 700, Y: 500, W: 96, H: 96},
		{X: 1180, Y: 620, W: 90, H: 120},
	}
	for i, rect := range rects {
		occluder := graphics.RectOccluder2D(rect, 1)
		if i >= 14 {
			occluder.Caster = graphics.ShadowCaster2D{ID: "dynamic", Dynamic: true}
		}
		world.AddOccluder(occluder)
	}

	segments := []graphics.Segment2D{
		{A: lmath.Vec2{X: 340, Y: 340}, B: lmath.Vec2{X: 460, Y: 470}},
		{A: lmath.Vec2{X: 760, Y: 320}, B: lmath.Vec2{X: 900, Y: 450}},
		{A: lmath.Vec2{X: 1020, Y: 820}, B: lmath.Vec2{X: 1180, Y: 950}},
		{A: lmath.Vec2{X: 1520, Y: 650}, B: lmath.Vec2{X: 1680, Y: 820}},
	}
	for _, segment := range segments {
		world.AddOccluder(graphics.SegmentOccluder2D(segment.A, segment.B, 1))
	}
}

func min1(value float32) float32 {
	if value > 1 {
		return 1
	}
	return value
}
