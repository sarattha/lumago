package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	"github.com/sarattha/lumago/engine/platform/desktop"
	"github.com/sarattha/lumago/engine/renderer"
	vulkanrenderer "github.com/sarattha/lumago/engine/renderer/vulkan"
	"github.com/sarattha/lumago/engine/scene"
)

func main() {
	runtime.LockOSThread()

	game := app.NewGame(app.Config{
		Width:  1280,
		Height: 720,
		Title:  "LumaGo Lighting Room Example",
	})

	world := scene.New()
	world.SetLightingConfig(graphics.LightingConfig2D{
		Ambient:   lmath.Color{R: 0.10, G: 0.10, B: 0.13, A: 1},
		DebugView: debugViewFromEnv(),
	})

	floor := material(game, "floor", 0.80, 0)
	wall := material(game, "wall", 0.65, 0)
	character := material(game, "character", 0.45, 0.05)
	prop := material(game, "prop", 0.50, 0.02)

	addSprite(world, floor, lmath.Rect{W: 256, H: 160}, lmath.Vec2{X: 520, Y: 460}, lmath.Color{R: 0.75, G: 0.78, B: 0.70, A: 1}, 0)
	addSprite(world, wall, lmath.Rect{W: 256, H: 96}, lmath.Vec2{X: 520, Y: 260}, lmath.Color{R: 0.58, G: 0.62, B: 0.68, A: 1}, 1)
	addSprite(world, character, lmath.Rect{W: 64, H: 96}, lmath.Vec2{X: 440, Y: 395}, lmath.Color{R: 0.95, G: 0.82, B: 0.62, A: 1}, 2)
	addSprite(world, prop, lmath.Rect{W: 80, H: 64}, lmath.Vec2{X: 610, Y: 405}, lmath.Color{R: 0.52, G: 0.84, B: 0.78, A: 1}, 2)

	updateLights(world, 0)
	game.SetScene(world)
	lightTime := float32(0)
	game.SetUpdateFunc(func(dt time.Duration) error {
		lightTime += float32(dt.Seconds())
		updateLights(world, lightTime)
		return nil
	})

	if os.Getenv("LUMAGO_RENDERER") == "nop" {
		game.SetRenderer(renderer.NewNopRenderer())
	} else {
		window, err := desktop.NewWindow(game.Config.Width, game.Config.Height, game.Config.Title)
		if err != nil {
			panic(err)
		}
		vulkanRenderer, err := vulkanrenderer.NewRenderer(vulkanrenderer.Config{
			Window:          window,
			ShaderDirectory: "shaders/bin",
			Validation:      os.Getenv("LUMAGO_VULKAN_VALIDATION") == "1",
		})
		if err != nil {
			window.Close()
			panic(err)
		}
		game.SetWindow(window)
		game.SetRenderer(vulkanRenderer)
	}

	if err := game.Run(); err != nil {
		panic(err)
	}
	fmt.Println("LumaGo lighting room finished.")
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

func updateLights(world *scene.Scene, t float32) {
	world.SetLights([]graphics.Light2D{
		light(380+90*float32(math.Sin(float64(t*1.1))), 300+70*float32(math.Cos(float64(t*0.8))), 340, lmath.Color{R: 1.00, G: 0.78, B: 0.45, A: 1}, 1.9),
		light(710+110*float32(math.Cos(float64(t*0.9))), 315+60*float32(math.Sin(float64(t*1.4))), 280, lmath.Color{R: 0.45, G: 0.68, B: 1.00, A: 1}, 1.4),
		light(470+75*float32(math.Sin(float64(t*1.7))), 500+45*float32(math.Sin(float64(t*0.7))), 220, lmath.Color{R: 0.85, G: 0.42, B: 1.00, A: 1}, 1.1),
		light(830+100*float32(math.Cos(float64(t*0.6))), 470+80*float32(math.Sin(float64(t*1.2))), 380, lmath.Color{R: 0.55, G: 1.00, B: 0.70, A: 1}, 1.2),
	})
}

func debugViewFromEnv() graphics.DebugView2D {
	switch os.Getenv("LUMAGO_DEBUG_VIEW") {
	case "color":
		return graphics.DebugViewSceneColor
	case "normal":
		return graphics.DebugViewSceneNormal
	case "light":
		return graphics.DebugViewLightBuffer
	default:
		return graphics.DebugViewFinalComposite
	}
}
