# Vulkan Renderer Backend

This package will contain the Vulkan renderer implementation.

Current status: scaffold only.

## Planned Files

- `instance.go` — Vulkan instance creation
- `device.go` — physical/logical device selection
- `swapchain.go` — swapchain creation and recreation
- `command.go` — command pool and command buffer management
- `sync.go` — semaphores and fences
- `buffer.go` — GPU buffer helpers
- `image.go` — image and image view helpers
- `descriptor.go` — descriptor set layouts and descriptor pools
- `pipeline.go` — graphics pipeline creation

## Planned Passes

- Sprite color pass
- Sprite normal pass
- Shadow map pass
- Light accumulation pass
- SDF shadow pass
- Composite pass

## Design Rule

The Vulkan backend must implement `engine/renderer.Renderer` and should not leak Vulkan types into `engine/graphics`, `engine/scene`, or `cmd/sandbox`.
