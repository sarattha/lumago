# HUD System Roadmap

## Current State

LumaGo does not have a dedicated HUD or UI system yet.

The current demos draw HUD-like elements as normal scene sprites. For example,
the T-Rex score display is built from rectangle sprites in the same scene as
the player, obstacles, lights, and occluders.

That means HUD elements currently:

- Use regular `graphics.SpriteDrawCommand` values.
- Share the world/camera coordinate space.
- Depend on manual positioning.
- Participate in normal layer sorting.
- Can be affected by lighting unless made emissive.
- Do not support anchors, layout, fonts, or text shaping.

This is acceptable for demos, but it is not an engine-level HUD system.

## Goal

Add a dedicated screen-space HUD path for 2D games that lets gameplay code draw
readable UI without depending on world camera placement or renderer internals.

HUD rendering should be:

- Screen-space by default.
- Independent from the world camera.
- Rendered after world sprites.
- Unlit by default.
- Easy to anchor to viewport edges and corners.
- Compatible with sprite batching.
- Explicit enough to keep the renderer abstraction clean.

## Proposed API Direction

Candidate scene-level API:

```go
world.AddHUDSprite(graphics.HUDSpriteDrawCommand{
	Sprite: sprite,
	Anchor: graphics.AnchorTopRight,
	Offset: lmath.Vec2{X: -24, Y: -24},
	Layer: 10,
})
```

Candidate text API:

```go
world.AddHUDText(graphics.HUDTextDrawCommand{
	Text: "00042",
	Font: font,
	Anchor: graphics.AnchorTopRight,
	Offset: lmath.Vec2{X: -32, Y: -24},
	Color: lmath.White(),
})
```

Initial anchor set:

- `AnchorTopLeft`
- `AnchorTop`
- `AnchorTopRight`
- `AnchorLeft`
- `AnchorCenter`
- `AnchorRight`
- `AnchorBottomLeft`
- `AnchorBottom`
- `AnchorBottomRight`

## Render Model

The renderer should receive world and HUD data separately:

```text
BeginFrame(worldCamera)
SubmitSpriteBatch(worldSprites)
SubmitHUDBatch(hudSprites, viewportSize)
ConfigureLighting(...)
SubmitLights(...)
EndFrame()
```

The first implementation can convert HUD commands into a normal sprite batch
using an identity screen-space camera. The important contract is that gameplay
code no longer manually maps HUD elements into world coordinates.

## Text Strategy

Text should not start as ad hoc rectangle digits in examples.

Recommended staged path:

- Phase 1: Bitmap font atlas with fixed-width glyphs.
- Phase 2: Proportional glyph metrics and alignment.
- Phase 3: Optional richer shaping if needed later.

Minimum useful text features:

- Left, center, and right alignment.
- Fixed pixel size.
- Color and alpha.
- Layer ordering.
- Optional shadow/outline for readability.
- No lighting by default.

## Milestones

1. Add HUD command types and storage on `scene.Scene`.
2. Add anchor resolution against the active viewport.
3. Build HUD sprite batches separately from world sprite batches.
4. Render HUD after the final world composite.
5. Add bitmap font atlas support for HUD text.
6. Convert T-Rex score display from rectangle digits to HUD text.
7. Add tests for anchor resolution, viewport resize behavior, and text layout.

## Non-Goals For First Pass

- Full retained-mode UI widgets.
- Flexbox or complex layout.
- Rich text.
- Input focus and forms.
- Editor tooling.
- Localization or advanced text shaping.

Those can be layered later after the screen-space render contract is stable.
