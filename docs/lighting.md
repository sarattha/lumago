# Lighting and Shadow Design

## Option B — Normal-Mapped 2D Lighting

Each sprite may have:

- Albedo texture
- Normal map texture
- Optional material parameters

Pipeline:

```text
Pass 1: Albedo pass
  Render sprite color into sceneColor.

Pass 2: Normal pass
  Render sprite normals into sceneNormal.

Pass 3: Light accumulation pass
  For each visible light, calculate lighting using sceneNormal.

Pass 4: Composite pass
  final = sceneColor * lighting
```

## Option C — Per-Light 2D Shadow Maps

The Go engine owns occluders. The Vulkan backend renders the shadow result.

Per light:

```text
1. Gather nearby occluder segments.
2. Convert occluder edges into light-relative space.
3. Render or compute nearest blocking distance per angle.
4. Store result in a 1D shadow texture or small 2D polar texture.
5. During light rendering, compare pixel distance against blocker distance.
```

MVP scope:

- Line segment occluders
- Rectangle occluders
- Tilemap wall occluders
- Hard shadows first
- Optional blur later

## Option D — SDF-Based Shadows

SDF shadows are experimental.

First implementation:

- Build a static SDF texture from level geometry.
- Upload the SDF texture to GPU.
- Raymarch from light to pixel in the light shader.
- Compare with shadow-map mode.

Concept:

```text
For each pixel affected by light:
  ray = pixelPosition - lightPosition
  march along ray
  sample SDF
  if distance <= threshold:
      shadowed
  else:
      lit
```

Current limitations:

- Enable with `LightingConfig2D.ShadowMode = ShadowModeSDFExperimental`, or run the lighting room demo with `LUMAGO_SHADOW_MODE=sdf`.
- Inspect the generated SDF with `DebugViewSDF`, or run the demo with `LUMAGO_DEBUG_VIEW=sdf`.
- Only the first shadow-casting light is raymarched in SDF mode; other shadow-casting lights remain unshadowed in this experimental path.
- The SDF is generated from static occluders only. Occluders marked with `ShadowCaster2D.Dynamic` are excluded so moving geometry does not poison the static field.
- The first implementation uses a coarse framebuffer-space SDF and CPU-side packing in the Vulkan backend. It is for comparison against hard shadow maps, not the default shipping shadow path.
