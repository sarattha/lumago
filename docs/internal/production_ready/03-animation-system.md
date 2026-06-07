# Animation System Production Plan

## Goal

Add a production-ready 2D sprite animation layer for frame clips, timing,
state-driven playback, event markers, and integration with game movement state.

## Why This Matters Compared With Unity

Unity provides Sprite Renderer, Animator, clips, transitions, events, and editor
preview. LumaGo needs a smaller Go-native equivalent so games are not forced to
hand-roll frame selection in every scene.

## Task Checklists

- [ ] Define animation clips as named frame sequences referencing asset-pipeline
      sprite frames.
- [ ] Support per-frame duration and clip-level playback speed.
- [ ] Support looping, one-shot, hold-last-frame, and ping-pong playback modes.
- [ ] Add sprite flip X/Y flags for common character animation use cases.
- [ ] Add an animator component or runtime type with current state, time, and
      active clip.
- [ ] Add simple state transitions keyed by game-provided state names, such as
      `idle`, `run`, `sprint`, `jump`, `fall`, and `attack`.
- [ ] Add event markers for gameplay callbacks, such as footstep, hit frame, or
      projectile spawn.
- [ ] Add deterministic fixed-step animation updates.
- [ ] Add debug inspection for current clip, frame index, playback time, and
      fired events.

## Exception Criteria

- Do not implement Unity's full Animator graph feature set in v1.
- Do not require animation state transitions to own gameplay logic.
- Do not make animation depend on Vulkan or renderer internals.
- Do not add skeletal animation until sprite-frame animation is stable.

## Evaluation

- A character can switch between idle, run, sprint, jump, and attack clips from
  game state.
- Animation frame timing is deterministic under the fixed update loop.
- Event markers fire once at the correct frame boundary.
- Sprite flip does not require duplicate atlas frames.
- Rendering still receives ordinary sprite draw commands.
