# LumaGo Design Manifesto

## 1. 2D First

LumaGo is not a general-purpose 3D engine. Every system should be optimized around 2D games:

- Sprites
- Tilemaps
- 2D cameras
- 2D lighting
- 2D shadows
- 2D occluders
- 2D materials

## 2. Go-Native API

Game developers should write normal Go code. The engine should avoid exposing low-level renderer details to gameplay code.

## 3. Renderer Abstraction

Vulkan is the first backend, but the engine should be organized through a renderer interface.

This keeps the door open for future backends without rewriting game code.

## 4. Batching by Default

The renderer should receive large batches of data rather than many small per-sprite calls.

## 5. Lighting Is a First-Class Feature

LumaGo’s visual identity should come from dynamic 2D lighting:

- Normal maps
- Point lights
- Ambient light
- Occluders
- Shadow maps
- SDF soft shadows

## 6. Debuggability Matters

The MVP should include debug views:

- Scene color
- Scene normal
- Light buffer
- Shadow map
- SDF texture
- Draw call count
- Sprite count
- Light count
