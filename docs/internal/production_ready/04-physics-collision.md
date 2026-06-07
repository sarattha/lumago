# Physics and Collision Production Plan

## Goal

Add a deterministic 2D collision layer suitable for platformers, top-down games,
tilemaps, triggers, projectiles, and simple character movement.

## Why This Matters Compared With Unity

Unity ships Rigidbody2D, Collider2D, triggers, casts, layers, and masks. LumaGo
does not need to duplicate all of Box2D immediately, but production games need a
shared collision contract instead of custom collision per demo.

## Task Checklists

- [ ] Define static, kinematic, and trigger body types.
- [ ] Add AABB colliders as the first supported shape.
- [ ] Add collision layers and masks.
- [ ] Add broadphase queries for tilemaps and world colliders.
- [ ] Add fixed-step overlap resolution for kinematic movement.
- [ ] Add trigger enter, stay, and exit events.
- [ ] Add ray casts and shape casts for ground checks, ledge checks, and line of
      sight.
- [ ] Integrate tilemap-generated static colliders.
- [ ] Add debug drawing for collider bounds, casts, normals, and trigger states.
- [ ] Add tests for deterministic collision outcomes across fixed-step updates.

## Exception Criteria

- Do not implement full rigid-body simulation, joints, or continuous collision in
  the first version.
- Do not make collision depend on rendered sprite bounds unless explicitly
  configured.
- Do not use frame-rate-dependent collision resolution.
- Do not merge lighting occluders and physics colliders into one type; allow
  generation from shared source data instead.

## Evaluation

- A character can walk, sprint, jump, land, and collide against tilemap walls.
- Trigger zones fire stable events without visual rendering.
- Ray casts can detect ground and walls with layer-mask filtering.
- Collision results are reproducible under the fixed update loop.
- Debug overlays make collider and cast behavior inspectable.
