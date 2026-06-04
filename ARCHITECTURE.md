# LumaGo Architecture

## Design Intent

LumaGo separates the engine into a Go-native game layer and a Vulkan renderer backend.

```text
Game code -> Scene -> Render commands -> Renderer interface -> Vulkan backend
```

## Engine Layer

The engine layer owns:

- App lifecycle
- Main loop
- Scene data
- Sprite definitions
- Material references
- Texture IDs
- Camera
- Lights
- Occluders
- Render command generation

## Renderer Layer

The renderer layer owns:

- GPU device resources
- Swapchain
- Command buffers
- Descriptor sets
- Pipelines
- Render targets
- Shader execution
- Final presentation

## Rule: No Vulkan Types in Game-Facing API

Game code should not import the Vulkan backend directly.

Good:

```go
scene.AddLight(graphics.Light2D{...})
```

Bad:

```go
scene.AddLight(vulkan.LightDescriptor{...})
```

## Render Command Flow

```text
Scene
  |
  |-- Build sprite commands
  |-- Build light commands
  |-- Build occluder commands
  |
Renderer
  |
  |-- Upload instance data
  |-- Execute color pass
  |-- Execute normal pass
  |-- Execute shadow pass
  |-- Execute lighting pass
  |-- Execute composite pass
```
