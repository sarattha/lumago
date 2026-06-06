# LumaGo Roadmap

## Vulkan Binding Policy

Every phase that needs Vulkan must extend a narrow internal Vulkan shim only for
the required calls in that phase. Do not expand gameplay APIs around Vulkan types,
and do not assume a broad generated binding is reliable enough for macOS/MoltenVK.
Keep MoltenVK portability details backend-local and documented.

## Phase 0 — PC Vulkan Foundation

Goal: create the minimum desktop runtime that opens a window and renders one textured quad through Vulkan.

Checklist:

- [x] Create desktop window
- [x] Create Vulkan instance
- [x] Select physical device
- [x] Create logical device
- [x] Create swapchain
- [x] Create command buffers
- [x] Create synchronization objects
- [x] Render one textured quad
- [ ] Capture frame in RenderDoc
- [x] Isolate required Vulkan/MoltenVK calls behind the internal shim

## Phase 1 — Sprite Engine and Batching

Goal: let Go handle sprites, animation, transforms, camera, layer sorting, and batch command generation.

Checklist:

- [x] Texture loader
- [x] Texture atlas
- [x] Sprite command buffer
- [x] Instance buffer
- [x] Layer sorting
- [x] Camera transform
- [x] Batch rendering
- [x] 1,000+ sprite demo
- [x] Extend the Vulkan shim only for texture upload, descriptor, buffer, and draw calls needed by batching

## Phase 2 — Normal-Mapped 2D Lighting

Goal: sprites react to dynamic lights using normal maps.

Checklist:

- [x] Material supports albedo + normal textures
- [x] Scene color render target
- [x] Scene normal render target
- [x] Light buffer render target
- [x] Point light shader
- [x] Ambient light
- [x] Final composite pass
- [x] Normal debug view
- [x] Extend the Vulkan shim only for render targets, descriptors, pipeline state, and pass transitions needed by lighting

## Phase 3 — Per-Light 2D Shadow Maps

Goal: lights are blocked by 2D occluders.

Checklist:

- [x] Occluder component
- [x] Segment extraction
- [x] Nearby occluder culling
- [x] Per-light shadow texture
- [x] Shadow lookup in light shader
- [x] Hard shadows
- [x] Debug shadow visualization
- [x] Extend the Vulkan shim only for shadow textures, shadow pass resources, and shader inputs needed by hard shadows

## Phase 4 — SDF-Based Experimental Shadows

Goal: use signed distance fields to create soft, stylized 2D shadows.

Checklist:

- [x] Generate static SDF from level geometry
- [x] Upload SDF texture
- [x] Raymarch from light to pixel
- [x] Render soft shadow factor
- [x] Compare with shadow-map mode
- [x] Debug SDF view
- [x] Extend the Vulkan shim only for SDF texture upload, sampling, and debug rendering

## Phase 5 — Polish and Performance

Goal: make the MVP stable, measurable, and testable.

Checklist:

- [x] Frame timing overlay
- [x] Draw call counter
- [x] Sprite counter
- [x] Light counter
- [x] GPU timing markers
- [x] Memory allocation checks
- [x] Shader reload in development mode
- [x] Basic config file
- [x] Extend the Vulkan shim only for timing markers, debug labels, and runtime toggles needed by measurement
