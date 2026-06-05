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

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			world.AddSprite(graphics.SpriteDrawCommand{
				Sprite: graphics.Sprite{
					Material: floorMaterial,
					Src:      lmath.Rect{X: 0, Y: 0, W: 24, H: 24},
					Color: lmath.Color{
						R: 0.55 + float32(x%8)*0.05,
						G: 0.75,
						B: 0.85 - float32(y%8)*0.04,
						A: 1,
					},
				},
				Transform: graphics.Transform2D{
					Position: lmath.Vec2{X: float32(x * 30), Y: float32(y * 30)},
					Scale:    lmath.Vec2{X: 1, Y: 1},
					Z:        float32((x + y) % 4),
				},
				Layer: y % 3,
			})
		}
	}

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
	cameraT := float32(0)
	game.SetUpdateFunc(func(dt time.Duration) error {
		cameraT += float32(dt.Seconds())
		camera := world.Camera()
		camera.Position = lmath.Vec2{
			X: 320 + 160*float32(math.Sin(float64(cameraT*0.7))),
			Y: 280 + 120*float32(math.Cos(float64(cameraT*0.5))),
		}
		camera.Zoom = 1.2
		world.SetCamera(camera)
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

	fmt.Println("LumaGo sandbox finished.")
}
