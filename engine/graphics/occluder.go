package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type Segment2D struct {
	A lmath.Vec2
	B lmath.Vec2
}

type ShadowCaster2D struct {
	ID      string
	Dynamic bool
}

type Occluder2D struct {
	Points   []lmath.Vec2
	Segments []Segment2D
	Layer    int
	Caster   ShadowCaster2D
}

func RectOccluder2D(rect lmath.Rect, layer int) Occluder2D {
	return Occluder2D{
		Points: []lmath.Vec2{
			{X: rect.X, Y: rect.Y},
			{X: rect.X + rect.W, Y: rect.Y},
			{X: rect.X + rect.W, Y: rect.Y + rect.H},
			{X: rect.X, Y: rect.Y + rect.H},
		},
		Layer: layer,
	}
}

func SegmentOccluder2D(a, b lmath.Vec2, layer int) Occluder2D {
	return Occluder2D{
		Segments: []Segment2D{{A: a, B: b}},
		Layer:    layer,
	}
}
