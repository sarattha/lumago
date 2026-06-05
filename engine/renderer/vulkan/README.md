# Vulkan Renderer Backend

This package will contain the Vulkan renderer implementation.

Current status: phase-00 foundation.

The backend opens a GLFW desktop surface through `engine/platform/desktop`, creates
a Vulkan instance/device/swapchain, uploads one quad vertex buffer, one index
buffer, and one small texture, then renders the textured quad using SPIR-V
shaders from `shaders/bin`.

The backend uses a backend-local Vulkan layer in `internal/vk`. It exposes only
the Vulkan calls and structs needed by the active renderer phase and calls the
system Vulkan SDK or MoltenVK directly through cgo.

Future Vulkan work should continue this direction deliberately: add only the
Vulkan/MoltenVK calls required by the active phase to the local layer, instead
of depending on a broad generated binding surface.

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

## RenderDoc Capture

Use the phase-00 capture helper after installing RenderDoc:

```bash
scripts/renderdoc_capture_phase00.sh
```

The script recompiles shaders, runs `make run` under `renderdoccmd capture`, and
writes captures under `captures/phase00-first-frame*.rdc`. On macOS it accepts
either `renderdoccmd` on `PATH` or
`/Applications/RenderDoc.app/Contents/MacOS/renderdoccmd`.
