package main

import (
	"fmt"
	"os"
	"runtime"

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
		Title:  "LumaGo Lighting Sandbox",
	})

	world := scene.New()

	floorAlbedo := game.Assets.LoadTexture("examples/lighting_room/assets/floor.png")
	floorNormal := game.Assets.LoadTexture("examples/lighting_room/assets/floor_n.png")

	floorMaterial := graphics.Material2D{
		Albedo:    floorAlbedo,
		Normal:    floorNormal,
		Roughness: 0.75,
		Emissive:  0.0,
	}

	world.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Material: floorMaterial,
			Src:      lmath.Rect{X: 0, Y: 0, W: 256, H: 256},
			Color:    lmath.White(),
		},
		Transform: graphics.Transform2D{
			Position: lmath.Vec2{X: 0, Y: 0},
			Scale:    lmath.Vec2{X: 1, Y: 1},
			Rotation: 0,
			Z:        0,
		},
		Layer: 0,
	})

	world.AddLight(graphics.Light2D{
		Position:    lmath.Vec2{X: 300, Y: 220},
		Radius:      400,
		Color:       lmath.Color{R: 1.0, G: 0.85, B: 0.55, A: 1.0},
		Intensity:   2.0,
		Falloff:     1.0,
		CastShadows: true,
	})

	world.AddOccluder(graphics.Occluder2D{
		Points: []lmath.Vec2{
			{X: 200, Y: 200},
			{X: 400, Y: 200},
			{X: 400, Y: 260},
			{X: 200, Y: 260},
		},
		Layer: 0,
	})

	game.SetScene(world)

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

	fmt.Println("LumaGo sandbox finished.")
}
