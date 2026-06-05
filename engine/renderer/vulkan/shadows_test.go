package vulkan

import (
	"testing"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
)

func TestPrepareOccluderSegmentsForFrameExtractsPolygonAndExplicitSegments(t *testing.T) {
	occluders := []graphics.Occluder2D{
		graphics.RectOccluder2D(lmath.Rect{X: 10, Y: 20, W: 30, H: 40}, 1),
		graphics.SegmentOccluder2D(lmath.Vec2{X: 100, Y: 110}, lmath.Vec2{X: 120, Y: 130}, 2),
	}

	segments := prepareOccluderSegmentsForFrame(nil, occluders, graphics.DefaultCamera2D())
	if len(segments) != 5 {
		t.Fatalf("segments=%d, want 5", len(segments))
	}
	if segments[0].A != (lmath.Vec2{X: 10, Y: 20}) || segments[0].B != (lmath.Vec2{X: 40, Y: 20}) {
		t.Fatalf("first rect segment=%+v", segments[0])
	}
	if segments[4].Layer != 2 {
		t.Fatalf("explicit segment layer=%d, want 2", segments[4].Layer)
	}
}

func TestCullSegmentsForLightRadius(t *testing.T) {
	segments := []shadowSegment{
		{A: lmath.Vec2{X: 10, Y: -5}, B: lmath.Vec2{X: 10, Y: 5}},
		{A: lmath.Vec2{X: 100, Y: -5}, B: lmath.Vec2{X: 100, Y: 5}},
	}
	light := graphics.Light2D{Position: lmath.Vec2{}, Radius: 50}

	got := cullSegmentsForLight(nil, segments, light)
	if len(got) != 1 || got[0].A.X != 10 {
		t.Fatalf("culled=%+v, want only near segment", got)
	}
}

func TestBuildLightShadowMapsCreatesOnlyShadowCastingLights(t *testing.T) {
	lights := []graphics.Light2D{
		{Position: lmath.Vec2{}, Radius: 100, CastShadows: true},
		{Position: lmath.Vec2{}, Radius: 100},
	}
	segments := []shadowSegment{{A: lmath.Vec2{X: 20, Y: -10}, B: lmath.Vec2{X: 20, Y: 10}}}

	got := buildLightShadowMaps(nil, lights, segments, 64)
	if len(got) != 1 {
		t.Fatalf("shadow maps=%d, want 1", len(got))
	}
	if got[0].LightIndex != 0 || got[0].Resolution != 64 || len(got[0].Segments) != 1 {
		t.Fatalf("shadow map metadata=%+v", got[0])
	}
}

func TestBuildLightShadowMapsSupportsEightShadowCastingLights(t *testing.T) {
	lights := make([]graphics.Light2D, 8)
	for i := range lights {
		lights[i] = graphics.Light2D{
			Position:    lmath.Vec2{X: float32(i * 10)},
			Radius:      100,
			CastShadows: true,
		}
	}
	segments := []shadowSegment{{A: lmath.Vec2{X: 20, Y: -10}, B: lmath.Vec2{X: 20, Y: 10}}}

	got := buildLightShadowMaps(nil, lights, segments, 32)
	if len(got) != 8 {
		t.Fatalf("shadow maps=%d, want 8", len(got))
	}
	for i := range got {
		if got[i].LightIndex != i || got[i].Resolution != 32 {
			t.Fatalf("shadow map %d metadata=%+v", i, got[i])
		}
	}
}

func TestShadowFactorBlocksPixelsBehindOccluder(t *testing.T) {
	light := graphics.Light2D{
		Position:    lmath.Vec2{},
		Radius:      100,
		CastShadows: true,
	}
	segments := []shadowSegment{{A: lmath.Vec2{X: 20, Y: -10}, B: lmath.Vec2{X: 20, Y: 10}}}
	shadows := buildLightShadowMaps(nil, []graphics.Light2D{light}, segments, 256)

	if got := shadowFactorForLight(lmath.Vec2{X: 10, Y: 0}, 0, []graphics.Light2D{light}, shadows); got != 1 {
		t.Fatalf("before occluder shadow=%f, want 1", got)
	}
	if got := shadowFactorForLight(lmath.Vec2{X: 40, Y: 0}, 0, []graphics.Light2D{light}, shadows); got != 0 {
		t.Fatalf("behind occluder shadow=%f, want 0", got)
	}
	if got := shadowFactorForLight(lmath.Vec2{X: 40, Y: 40}, 0, []graphics.Light2D{light}, shadows); got != 1 {
		t.Fatalf("outside occluder angle shadow=%f, want 1", got)
	}
}

func TestMovingLightRegeneratesShadowLookup(t *testing.T) {
	segments := []shadowSegment{{A: lmath.Vec2{X: 20, Y: -10}, B: lmath.Vec2{X: 20, Y: 10}}}
	leftLight := graphics.Light2D{
		Position:    lmath.Vec2{},
		Radius:      100,
		CastShadows: true,
	}
	rightLight := graphics.Light2D{
		Position:    lmath.Vec2{X: 60, Y: 0},
		Radius:      100,
		CastShadows: true,
	}

	leftShadows := buildLightShadowMaps(nil, []graphics.Light2D{leftLight}, segments, 256)
	rightShadows := buildLightShadowMaps(nil, []graphics.Light2D{rightLight}, segments, 256)
	pixel := lmath.Vec2{X: 40, Y: 0}

	if got := shadowFactorForLight(pixel, 0, []graphics.Light2D{leftLight}, leftShadows); got != 0 {
		t.Fatalf("pixel with left light shadow=%f, want 0", got)
	}
	if got := shadowFactorForLight(pixel, 0, []graphics.Light2D{rightLight}, rightShadows); got != 1 {
		t.Fatalf("pixel with moved right light shadow=%f, want 1", got)
	}
}
