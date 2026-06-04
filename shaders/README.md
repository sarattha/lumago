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
glslc shaders/sprite_color.vert -o shaders/bin/sprite_color.vert.spv
glslc shaders/sprite_color.frag -o shaders/bin/sprite_color.frag.spv
```
