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
- Hot reload
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
