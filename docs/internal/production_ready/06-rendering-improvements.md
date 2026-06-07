# Rendering Improvements Production Plan

## Goal

Harden LumaGo's 2D Vulkan renderer for production scenes with batching, culling,
layer masks, richer materials, text, UI, particles, and post-processing.

## Why This Matters Compared With Unity

LumaGo already has a focused rendering direction: sprite batching,
normal-mapped lighting, hard shadows, and experimental SDF shadows. Unity's
advantage is breadth and tooling around that renderer. The production renderer
should keep LumaGo's narrow backend boundary while filling common 2D needs.

## Task Checklists

- [ ] Support batching across multiple atlas pages or texture arrays without
      per-sprite renderer calls.
- [ ] Add camera culling for sprites, tile chunks, lights, and occluders.
- [ ] Add render layers and layer masks for cameras, sprites, lights, and
      occluders.
- [ ] Add per-light layer masks so lights affect only selected scene layers.
- [ ] Add material blend modes for alpha, additive, multiply, and unlit sprites.
- [ ] Move production lighting work toward GPU-native passes and avoid CPU-side
      lit sprite subdivision in shipping paths.
- [ ] Add render-to-texture targets for mini-maps, portals, previews, and
      offscreen effects.
- [ ] Add a separate UI/HUD render pass after world rendering.
- [ ] Add bitmap or MSDF font rendering with atlas support.
- [ ] Add nine-slice sprites for UI panels and scalable 2D elements.
- [ ] Add a basic particle system that batches textured quads.
- [ ] Add post-processing basics: bloom, color grading, vignette, and screen
      fade.

## Exception Criteria

- Do not leak Vulkan handles or descriptor details into `engine/graphics` or game
  code.
- Do not optimize by breaking deterministic render ordering.
- Do not make post-processing mandatory for simple demos.
- Do not expand the Vulkan shim broadly; add only backend calls needed by each
  feature.

## Evaluation

- Large tilemap scenes render with camera culling and stable draw counts.
- Multiple atlas pages do not force one draw call per sprite.
- Lights and cameras respect layer masks.
- UI renders after world content and remains screen-space stable.
- Text, particles, and post effects can be toggled independently for diagnostics.
