# Tilemap System Production Plan

## Goal

Add a native tilemap system that supports grid-based 2D worlds, efficient
rendering, collision generation, occluder generation, and external map import.

## Why This Matters Compared With Unity

Unity's Tilemap workflow is a major production feature for 2D games. LumaGo
currently composes tile-like scenes from sprites. Production projects need maps
as structured data, not hand-authored sprite loops.

## Task Checklists

- [ ] Define a tilemap data model with map size, tile size, layers, tilesets, and
      cell coordinates.
- [ ] Support orthogonal tilemaps first.
- [ ] Add tile chunks for rendering, culling, and dirty-region rebuilds.
- [ ] Resolve tile IDs to atlas frames from the asset pipeline.
- [ ] Support multiple tile layers with stable draw ordering.
- [ ] Generate sprite batches from visible tile chunks.
- [ ] Generate static collision shapes from solid tiles.
- [ ] Generate rectangle and segment occluders from wall tiles for lighting.
- [ ] Add animated tile definitions that resolve to frame sequences.
- [ ] Add Tiled `.tmx` or `.json` import after the internal tilemap model is
      stable.

## Exception Criteria

- Do not implement isometric, hex, or staggered maps in the first production
  version.
- Do not require the renderer to know about tilemaps; tilemaps should produce
  normal sprite, collider, and occluder data.
- Do not rebuild the full map every frame when only a chunk changes.
- Do not support arbitrary per-tile scripts in this phase.

## Evaluation

- A 16x16 orthogonal map renders with correct layer ordering and camera culling.
- Static tile collision matches visible wall tiles.
- Tilemap wall occluders cast shadows in the existing lighting path.
- Editing one chunk rebuilds only that chunk's render data.
- Imported Tiled maps match their source dimensions, layers, and tileset frames.
