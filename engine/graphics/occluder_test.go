package graphics

import (
	"testing"

	lmath "github.com/sarattha/lumago/engine/math"
)

func TestOccluderConstructors(t *testing.T) {
	rect := RectOccluder2D(lmath.Rect{X: 10, Y: 20, W: 30, H: 40}, 2)
	if len(rect.Points) != 4 || rect.Layer != 2 {
		t.Fatalf("rect occluder=%+v, want 4 points on layer 2", rect)
	}
	if rect.Points[2] != (lmath.Vec2{X: 40, Y: 60}) {
		t.Fatalf("rect point 2=%+v, want (40,60)", rect.Points[2])
	}

	segment := SegmentOccluder2D(lmath.Vec2{X: 1, Y: 2}, lmath.Vec2{X: 3, Y: 4}, 5)
	if len(segment.Segments) != 1 || segment.Layer != 5 {
		t.Fatalf("segment occluder=%+v, want one segment on layer 5", segment)
	}
}
