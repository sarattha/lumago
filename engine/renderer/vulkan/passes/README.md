# Vulkan Render Passes

Planned passes:

1. `sprite_color` — writes albedo sprites into scene color.
2. `sprite_normal` — writes normal maps into scene normal.
3. `shadow_map` — generates per-light shadow data from occluders.
4. `light_accum` — accumulates point lights using normals and shadows.
5. `sdf_shadow` — experimental SDF raymarched shadow mode.
6. `composite` — combines scene color, light buffer, and emissive.
