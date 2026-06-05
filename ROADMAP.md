# LumaGo Roadmap

## Vulkan Binding Policy

Every phase that needs Vulkan must extend a narrow internal Vulkan shim only for
the required calls in that phase. Do not expand gameplay APIs around Vulkan types,
and do not assume a broad generated binding is reliable enough for macOS/MoltenVK.
Keep MoltenVK portability details backend-local and documented.

## Phase 0 — PC Vulkan Foundation

Goal: create the minimum desktop runtime that opens a window and renders one textured quad through Vulkan.

Checklist:

- [ ] Create desktop window
- [ ] Create Vulkan instance
- [ ] Select physical device
- [ ] Create logical device
- [ ] Create swapchain
- [ ] Create command buffers
- [ ] Create synchronization objects
- [ ] Render one textured quad
- [ ] Capture frame in RenderDoc
- [ ] Isolate required Vulkan/MoltenVK calls behind the internal shim

## Phase 1 — Sprite Engine and Batching

Goal: let Go handle sprites, animation, transforms, camera, layer sorting, and batch command generation.

Checklist:

- [ ] Texture loader
- [ ] Texture atlas
- [ ] Sprite command buffer
- [ ] Instance buffer
- [ ] Layer sorting
- [ ] Camera transform
- [ ] Batch rendering
- [ ] 1,000+ sprite demo
- [ ] Extend the Vulkan shim only for texture upload, descriptor, buffer, and draw calls needed by batching

## Phase 2 — Normal-Mapped 2D Lighting

Goal: sprites react to dynamic lights using normal maps.

Checklist:

- [ ] Material supports albedo + normal textures
- [ ] Scene color render target
- [ ] Scene normal render target
- [ ] Light buffer render target
- [ ] Point light shader
- [ ] Ambient light
- [ ] Final composite pass
- [ ] Normal debug view
- [ ] Extend the Vulkan shim only for render targets, descriptors, pipeline state, and pass transitions needed by lighting

## Phase 3 — Per-Light 2D Shadow Maps

Goal: lights are blocked by 2D occluders.

Checklist:

- [ ] Occluder component
- [ ] Segment extraction
- [ ] Nearby occluder culling
- [ ] Per-light shadow texture
- [ ] Shadow lookup in light shader
- [ ] Hard shadows
- [ ] Debug shadow visualization
- [ ] Extend the Vulkan shim only for shadow textures, shadow pass resources, and shader inputs needed by hard shadows

## Phase 4 — SDF-Based Experimental Shadows

Goal: use signed distance fields to create soft, stylized 2D shadows.

Checklist:

- [ ] Generate static SDF from level geometry
- [ ] Upload SDF texture
- [ ] Raymarch from light to pixel
- [ ] Render soft shadow factor
- [ ] Compare with shadow-map mode
- [ ] Debug SDF view
- [ ] Extend the Vulkan shim only for SDF texture upload, sampling, and debug rendering

## Phase 5 — Polish and Performance

Goal: make the MVP stable, measurable, and testable.

Checklist:

- [ ] Frame timing overlay
- [ ] Draw call counter
- [ ] Sprite counter
- [ ] Light counter
- [ ] GPU timing markers
- [ ] Memory allocation checks
- [ ] Shader reload in development mode
- [ ] Basic config file
- [ ] Extend the Vulkan shim only for timing markers, debug labels, and runtime toggles needed by measurement
