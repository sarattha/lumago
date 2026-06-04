# LumaGo

**LumaGo** is a Go-native 2D game engine prototype focused on sprite batching, normal-mapped lighting, dynamic shadows, and Vulkan-powered rendering.

> Status: **pre-alpha / architecture scaffold**  
> Target platform: **PC first**  
> Rendering backend: **Vulkan-first, renderer-abstracted**  
> Engine focus: **2D games only**

## Vision

LumaGo is designed around a clean separation:

```text
Go engine decides what should be rendered.
Vulkan backend decides how it is rendered on the GPU.
```

The game-facing API should stay simple. Game developers should work with sprites, lights, materials, cameras, scenes, and occluders. Vulkan details should remain inside the renderer backend.

## Core Goals

- Go-native 2D engine API
- PC-first runtime
- Vulkan renderer backend
- Sprite batching
- Texture atlas support
- Normal-mapped 2D lighting
- Per-light 2D shadow maps
- Experimental SDF-based soft shadows
- Debug views for color, normals, lights, shadows, and SDF
- Minimal per-frame allocation in the hot path

## Non-Goals for MVP

- 3D engine support
- Mobile support
- Full editor
- Networking
- Physics engine
- Audio engine
- Scripting
- Asset marketplace
- Plugin system

## Architecture

```text
Go 2D Game Engine
  |
  |-- Scene / ECS
  |-- Sprite system
  |-- Animation system
  |-- Tilemap system
  |-- Camera
  |-- Light components
  |-- Occluder components
  |-- Render command generation
  |
  v
Vulkan Renderer Backend
  |
  |-- Swapchain
  |-- Command buffers
  |-- GPU buffers
  |-- Texture upload
  |-- Sprite batch pipelines
  |-- Normal-map lighting pipeline
  |-- Shadow-map pipeline
  |-- SDF shadow pipeline
  |-- Final composite pass
```

## MVP Rendering Pipeline

```text
Frame
  |
  |-- Go Update()
  |     |-- update sprites
  |     |-- update animations
  |     |-- update lights
  |     |-- update occluders
  |
  |-- Go BuildRenderCommands()
  |     |-- SpriteDrawCommand[]
  |     |-- Light2D[]
  |     |-- Occluder2D[]
  |
  |-- Vulkan Pass 1: Scene color
  |     |-- render albedo sprites
  |
  |-- Vulkan Pass 2: Scene normal
  |     |-- render normal-map sprites
  |
  |-- Vulkan Pass 3: Shadow maps
  |     |-- build per-light shadow maps from occluders
  |
  |-- Vulkan Pass 4: Light accumulation
  |     |-- render lights
  |     |-- sample normal texture
  |     |-- sample shadow map
  |
  |-- Vulkan Pass 5: SDF shadow mode, optional
  |     |-- sample SDF texture
  |     |-- raymarch light visibility
  |
  |-- Vulkan Pass 6: Composite
  |     |-- sceneColor * lightBuffer + emissive
  |
  |-- Present
```

## Requirements

- Go 1.23+
- GitHub CLI, optional
- Vulkan SDK, later, when the real Vulkan backend is implemented

## Quick Start

```bash
go test ./...
go run ./cmd/sandbox
```

The current scaffold uses a `NopRenderer` so the repository can compile before the Vulkan backend is implemented.

## Development Commands

```bash
make fmt
make test
make run
```

## Current Package API Sketch

```go
game := app.NewGame(app.Config{
    Width:  1280,
    Height: 720,
    Title:  "LumaGo Lighting Sandbox",
})

scene := scene.New()

scene.AddLight(graphics.Light2D{
    Position:    lmath.Vec2{X: 300, Y: 220},
    Radius:      400,
    Color:       lmath.Color{R: 1.0, G: 0.85, B: 0.55, A: 1.0},
    Intensity:   2.0,
    CastShadows: true,
})

game.SetScene(scene)
game.Run()
```

## MVP Acceptance Target

A PC demo room with:

- 1,000 sprites
- 4 normal-mapped materials
- 4 dynamic point lights
- 2 shadow-casting lights
- 20 rectangle/segment occluders
- Camera movement
- Debug views for color, normal, light, shadow, and SDF

## License

MIT License.
