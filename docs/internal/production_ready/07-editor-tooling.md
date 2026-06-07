# Editor and Tooling Production Plan

## Goal

Build lightweight production tooling around LumaGo so artists, designers, and
engineers can inspect assets, author scenes, debug rendering, and iterate quickly
without a full Unity-scale editor.

## Why This Matters Compared With Unity

Unity wins production workflows because of authoring and inspection, not only
runtime features. LumaGo needs practical tools for asset preview, tilemaps,
scene inspection, collision, lighting, profiling, and hot reload.

## Task Checklists

- [ ] Add an asset inspector for imported textures, sprite frames, pivots,
      slicing, PPU, tile size, normal-map pairing, and atlas placement.
- [ ] Add sprite and atlas preview with nearest-neighbor zoom and transparent
      background modes.
- [ ] Add tilemap inspection for layers, chunks, tilesets, collision flags, and
      occluder generation.
- [ ] Prefer Tiled integration first for map authoring, then add native editing
      only where external tools are insufficient.
- [ ] Add a scene viewer that can load serialized scenes and display entities,
      sprites, lights, colliders, and occluders.
- [ ] Add debug overlays for sprite bounds, atlas frame IDs, draw order, collider
      bounds, tile chunks, occluder segments, light radius, shadow factors, and
      SDF state.
- [ ] Add live controls for debug views: final, scene color, normal, light
      buffer, shadow factor, and SDF.
- [ ] Add hot reload for changed textures, metadata, tilemaps, animation clips,
      shaders, and scene files.
- [ ] Add frame capture and deterministic replay for reproducing rendering and
      gameplay bugs.
- [ ] Add performance panels for CPU time, GPU time, pass timings, draw calls,
      sprite count, light count, occluder count, and hot-path allocations.
- [ ] Add validation reports for assets, tilemaps, scenes, animations, and
      renderer compatibility.
- [ ] Add command-line tooling for import, validate, preview, and package tasks
      so CI can run the same checks as local development.

## Exception Criteria

- Do not block production readiness on a full drag-and-drop Unity clone.
- Do not make the runtime depend on editor-only packages.
- Do not hide validation failures behind visual tooling only; every validation
  path needs CLI output.
- Do not make hot reload mutate source files unless explicitly requested by a
  user action.
- Do not require Vulkan to inspect asset metadata; non-render validation should
  run headless.

## Evaluation

- A developer can inspect a sprite sheet, verify `16x16` slicing, and see atlas
  padding before running the game.
- A designer can open a tilemap, inspect collision and occluder generation, and
  compare it against rendered output.
- A rendering issue can be debugged with buffer views, light overlays, pass
  timings, and replayable frame data.
- CI can run import and validation checks without opening a window.
- Existing demos remain runnable without editor tooling installed.
