# Shaders

This directory contains shader source for the Vulkan renderer.

Planned shader groups:

- Sprite color pass
- Sprite normal pass
- Light accumulation
- Shadow map
- SDF shadow
- Composite

Compile GLSL to SPIR-V in the future with a script such as:

```bash
make shaders
```

Phase 00 uses `quad.vert` and `quad.frag` to render one visible checker-textured
quad to the swapchain. Later phases will replace this with sprite batching and
material texture descriptors.
