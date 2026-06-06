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

- Go 1.26.4+
- GitHub CLI, optional
- Vulkan runtime/tooling for the phase-00 PC backend:
  - `brew install glfw vulkan-loader vulkan-headers vulkan-validationlayers vulkan-tools shaderc`

## Quick Start

```bash
go test ./...
make shaders
make run
```

Use `make run-nop` to run the sandbox through the fallback `NopRenderer`.
Use `make run-lighting` to run the MVP lighting room demo, or
`make run-lighting-nop` to run the same scene through `NopRenderer` diagnostics.

## Development Commands

```bash
make fmt
make test
make vet
make shaders
make run
make run-lighting
```

On macOS/Homebrew, `make run` sets the loader and MoltenVK ICD environment expected by the local Vulkan backend layer. For direct runs, use the same environment shown in the `Makefile`.

## MVP Runtime Requirements

- Target demo resolution: 1920x1080
- Target frame rate: 60 FPS on a mid-range desktop GPU
- Go 1.26.4+ and a desktop Vulkan runtime
- macOS/Homebrew Vulkan stack: `glfw`, `vulkan-loader`, `vulkan-headers`, `vulkan-validationlayers`, `vulkan-tools`, `shaderc`, and MoltenVK
- Shader binaries in `shaders/bin`, generated with `make shaders`

The lighting room reads `examples/lighting_room/lumago.conf` for window size,
renderer mode, debug view, shadow mode, and development toggles. Environment
variables override the file:

- `LUMAGO_RENDERER=nop` runs without Vulkan.
- `LUMAGO_DEBUG_VIEW=color|normal|light|shadow|sdf` selects the debug output.
- `LUMAGO_SHADOW_MODE=sdf` enables experimental SDF shadows.
- `LUMAGO_FRAME_LIMIT=1` is useful for short diagnostic runs.
- `LUMAGO_VULKAN_VALIDATION=1` enables validation layers when present.

When the demo debug overlay is enabled, diagnostics are printed with CPU frame
time, GPU frame time when available, hot-path allocation bytes, draw calls,
sprite count, light count, occluder count, and pass timings for color, normal,
shadow, light, SDF, and composite passes.

## Vulkan Binding Policy

Vulkan integration should grow through a narrow internal backend shim. Each
roadmap phase should add only the Vulkan/MoltenVK calls it needs, keeping Vulkan
handles and portability details out of gameplay-facing packages. Do not add a
generated Vulkan binding dependency for future phases; extend
`engine/renderer/vulkan/internal/vk` with the smallest required call surface.

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
