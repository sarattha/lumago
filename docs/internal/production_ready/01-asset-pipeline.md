# Asset Pipeline Production Plan

## Goal

Build a deterministic 2D asset pipeline that turns source artwork and metadata
into stable runtime assets for sprites, atlases, normal maps, animations, fonts,
and tilemaps.

## Why This Matters Compared With Unity

Unity's production advantage starts at import time: sprites have Pixels Per Unit,
slicing, pivots, filtering, atlas workflows, and stable metadata. LumaGo needs the
same kind of repeatable import contract without requiring a full editor first.

## Task Checklists

- [ ] Define asset metadata files for textures, sprites, atlases, normal maps,
      animation frames, fonts, and tilemap sources.
- [ ] Add first-class `PixelsPerUnit` and `TileSize` settings, with `16x16` as a
      documented supported tile profile.
- [ ] Support sprite sheet slicing with explicit rectangles, pivots, names, and
      source texture references.
- [ ] Add atlas packing metadata with padding and extrusion to prevent texture
      bleeding.
- [ ] Add filter and wrap settings, including nearest-neighbor defaults for pixel
      art.
- [ ] Add normal-map pairing conventions, such as `hero.png` plus `hero_n.png`,
      with explicit override metadata.
- [ ] Introduce deterministic asset IDs that do not depend on runtime load order.
- [ ] Add generated asset manifests for runtime loading.
- [ ] Add development hot reload for changed metadata and source images.
- [ ] Add validation errors for missing files, duplicate sprite names, invalid
      rectangles, mismatched normal-map sizes, and unsupported formats.

## Exception Criteria

- Do not build a full Unity-style importer UI in this phase.
- Do not add platform-specific texture compression until the base manifest format
  is stable.
- Do not break direct `LoadTexture` usage; keep it as a simple fallback path for
  demos and tests.
- Do not expose Vulkan texture handles or backend details in asset metadata.

## Evaluation

- A 16x16 tile sheet can be sliced into named sprites and rendered at predictable
  world size.
- Re-running import produces the same asset IDs and manifest ordering.
- Atlas padding prevents visible seams in nearest-neighbor pixel-art scenes.
- Missing or invalid normal maps fail with clear diagnostics or use a documented
  neutral-normal fallback.
- Existing lighting room and T-Rex demos still load their assets.
