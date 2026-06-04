package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type Light2D struct {
	Position    lmath.Vec2
	Radius      float32
	Color       lmath.Color
	Intensity   float32
	Falloff     float32
	CastShadows bool
}
