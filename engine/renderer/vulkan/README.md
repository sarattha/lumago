# Vulkan Renderer Backend

This package will contain the Vulkan renderer implementation.

Current status: phase-00 foundation.

The backend opens a GLFW desktop surface through `engine/platform/desktop`, creates
a Vulkan instance/device/swapchain, and renders one checker-textured quad using
SPIR-V shaders from `shaders/bin`.

On macOS, the current `vulkan-go` binding needs small Darwin shims for instance,
device, surface, and graphics pipeline creation because its generated create-info
packing does not work reliably with Homebrew MoltenVK portability requirements.

Future Vulkan work should continue this direction deliberately: add a narrow
internal shim for only the Vulkan calls required by the active phase, instead of
depending on a broad generated binding surface.

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

## Running

```bash
make shaders
make run
```

Set `LUMAGO_RENDERER=nop` or use `make run-nop` for the fallback renderer. Set
`LUMAGO_VULKAN_DEBUG=1` to print requested Vulkan instance/device extensions.
