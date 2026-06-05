package graphics

import "math"

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

type Mat3 [9]float32

func (c Camera2D) ViewMatrix() Mat3 {
	zoom := c.Zoom
	if zoom == 0 {
		zoom = 1
	}

	sin, cos := float32(math.Sin(float64(-c.Rotation))), float32(math.Cos(float64(-c.Rotation)))
	tx := -c.Position.X
	ty := -c.Position.Y

	return Mat3{
		cos * zoom, -sin * zoom, (cos*tx - sin*ty) * zoom,
		sin * zoom, cos * zoom, (sin*tx + cos*ty) * zoom,
		0, 0, 1,
	}
}

func (m Mat3) TransformPoint(point lmath.Vec2) lmath.Vec2 {
	return lmath.Vec2{
		X: m[0]*point.X + m[1]*point.Y + m[2],
		Y: m[3]*point.X + m[4]*point.Y + m[5],
	}
}
