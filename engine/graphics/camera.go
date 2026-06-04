package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type Camera2D struct {
	Position lmath.Vec2
	Zoom     float32
	Rotation float32
}

func DefaultCamera2D() Camera2D {
	return Camera2D{
		Zoom: 1.0,
	}
}
