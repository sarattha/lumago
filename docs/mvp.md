# LumaGo MVP

## Goal

Create a PC-only 2D engine prototype in Go with Vulkan rendering, sprite batching, normal-mapped lighting, simple shadow-casting lights, and an experimental SDF shadow mode.

## Must-Have

- Desktop window
- Vulkan renderer
- Sprite batching
- Texture atlas
- Albedo + normal map sprites
- Point lights
- Ambient light
- Light accumulation pass
- Rectangle/segment occluders
- Hard shadows using per-light shadow maps
- Final composite pass
- Debug overlays
- Frame timing and draw-count diagnostics
- Basic runtime config for window size, renderer mode, debug views, and development toggles

## Should-Have

- Soft shadow blur
- Light radius visualization
- Normal map debug view
- Shadow map debug view
- Sprite layer sorting
- Camera movement
- Basic asset loader

## Experimental

- Static SDF texture generation
- One SDF-raymarched light
- SDF debug visualization

## Not in MVP

- Editor
- Mobile
- Physics engine
- Networking
- Particle system
- Scripting
- Audio
- Complex ECS
- Development shader reload
- Advanced materials

## Demo Target

A PC demo room with:

- 1,000 sprites
- 4 normal-mapped materials
- 4 dynamic point lights
- 2 shadow-casting lights
- 20 rectangle/segment occluders
- Camera movement
- Debug views for color, normal, light, shadow, and SDF
- Frame timing, draw call, sprite, light, occluder, and hot-path allocation metrics

## Launch

```bash
make run-lighting
```

For a GPU-independent diagnostic run:

```bash
make run-lighting-nop
```

The demo reads `examples/lighting_room/lumago.conf`. Use `LUMAGO_DEBUG_VIEW`
with `color`, `normal`, `light`, `shadow`, or `sdf` to validate each debug view,
and `LUMAGO_SHADOW_MODE=sdf` to exercise the experimental SDF path.

## Verification Evidence

Last local diagnostic run:

```text
LUMAGO_RENDERER=nop go run ./examples/lighting_room
target=1920x1080@60fps
observed_fps=2385.2
sprites=1000
materials=4
lights=4
shadow_lights=2
occluders=20
draws=1
cpu_ms=0.419
alloc_bytes=227232
debug=final
```

The `NopRenderer` run verifies scene composition, counters, frame timing output,
and allocation measurement without requiring a live Vulkan display. Use
`make run-lighting` on a Vulkan-capable desktop to capture the final GPU FPS and
pass timing values for hardware evaluation.
