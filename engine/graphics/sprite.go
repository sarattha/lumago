package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type Transform2D struct {
	Position lmath.Vec2
	Scale    lmath.Vec2
	Rotation float32
	Z        float32
}

type Sprite struct {
	Material Material2D
	Src      lmath.Rect
	Color    lmath.Color
}

type SpriteDrawCommand struct {
	Sprite    Sprite
	Transform Transform2D
	Layer     int
}
