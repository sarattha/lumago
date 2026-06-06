package graphics

import (
	"math"
	"sort"
)

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

type SpriteVertex struct {
	Position lmath.Vec2
	UV       lmath.Vec2
	Color    lmath.Color
}

type SpriteBatchStats struct {
	SpriteCount    int
	DrawCalls      int
	VertexCount    int
	IndexCount     int
	ViewportWidth  int
	ViewportHeight int
}

type SpriteBatch struct {
	Commands []SpriteDrawCommand
	Vertices []SpriteVertex
	Indices  []uint32
	Stats    SpriteBatchStats
}

func (b *SpriteBatch) Reset() {
	b.Commands = b.Commands[:0]
	b.Vertices = b.Vertices[:0]
	b.Indices = b.Indices[:0]
	b.Stats = SpriteBatchStats{}
}

func (b *SpriteBatch) Build(commands []SpriteDrawCommand, camera Camera2D, viewportWidth, viewportHeight int) {
	b.Reset()
	if len(commands) == 0 {
		return
	}

	b.Commands = append(b.Commands, commands...)
	SortSpriteCommands(b.Commands)

	b.Vertices = ensureSpriteVertices(b.Vertices, len(b.Commands)*4)
	b.Indices = ensureSpriteIndices(b.Indices, len(b.Commands)*6)

	view := camera.ViewMatrix()
	for i, command := range b.Commands {
		writeSpriteGeometry(b.Vertices[i*4:], b.Indices[i*6:], uint32(i*4), command, view, viewportWidth, viewportHeight)
	}

	b.Stats = SpriteBatchStats{
		SpriteCount:    len(b.Commands),
		DrawCalls:      1,
		VertexCount:    len(b.Vertices),
		IndexCount:     len(b.Indices),
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
	}
}

func SortSpriteCommands(commands []SpriteDrawCommand) {
	sort.SliceStable(commands, func(i, j int) bool {
		a, b := commands[i], commands[j]
		if a.Layer != b.Layer {
			return a.Layer < b.Layer
		}
		if a.Transform.Z != b.Transform.Z {
			return a.Transform.Z < b.Transform.Z
		}
		if a.Sprite.Material.Albedo != b.Sprite.Material.Albedo {
			return a.Sprite.Material.Albedo < b.Sprite.Material.Albedo
		}
		if a.Sprite.Material.Normal != b.Sprite.Material.Normal {
			return a.Sprite.Material.Normal < b.Sprite.Material.Normal
		}
		if a.Sprite.Material.Roughness != b.Sprite.Material.Roughness {
			return a.Sprite.Material.Roughness < b.Sprite.Material.Roughness
		}
		return a.Sprite.Material.Emissive < b.Sprite.Material.Emissive
	})
}

func ensureSpriteVertices(vertices []SpriteVertex, count int) []SpriteVertex {
	if cap(vertices) < count {
		return make([]SpriteVertex, count)
	}
	return vertices[:count]
}

func ensureSpriteIndices(indices []uint32, count int) []uint32 {
	if cap(indices) < count {
		return make([]uint32, count)
	}
	return indices[:count]
}

func writeSpriteGeometry(vertices []SpriteVertex, indices []uint32, base uint32, command SpriteDrawCommand, view Mat3, viewportWidth, viewportHeight int) {
	src := command.Sprite.Src
	if src.W == 0 {
		src.W = 1
	}
	if src.H == 0 {
		src.H = 1
	}

	scale := command.Transform.Scale
	if scale.X == 0 {
		scale.X = 1
	}
	if scale.Y == 0 {
		scale.Y = 1
	}

	w := src.W * scale.X
	h := src.H * scale.Y
	halfW := w * 0.5
	halfH := h * 0.5
	local := [4]lmath.Vec2{
		{X: -halfW, Y: -halfH},
		{X: halfW, Y: -halfH},
		{X: halfW, Y: halfH},
		{X: -halfW, Y: halfH},
	}
	uv := [4]lmath.Vec2{
		{X: src.X, Y: src.Y + src.H},
		{X: src.X + src.W, Y: src.Y + src.H},
		{X: src.X + src.W, Y: src.Y},
		{X: src.X, Y: src.Y},
	}

	sin, cos := sincos(command.Transform.Rotation)
	for i := range local {
		rotated := lmath.Vec2{
			X: local[i].X*cos - local[i].Y*sin,
			Y: local[i].X*sin + local[i].Y*cos,
		}
		world := rotated.Add(command.Transform.Position)
		position := view.TransformPoint(world)
		vertices[i] = SpriteVertex{
			Position: normalizeToClip(position, viewportWidth, viewportHeight),
			UV:       uv[i],
			Color:    command.Sprite.Color,
		}
	}

	indices[0] = base
	indices[1] = base + 1
	indices[2] = base + 2
	indices[3] = base + 2
	indices[4] = base + 3
	indices[5] = base
}

func normalizeToClip(point lmath.Vec2, viewportWidth, viewportHeight int) lmath.Vec2 {
	if viewportWidth <= 0 || viewportHeight <= 0 {
		return point
	}
	return lmath.Vec2{
		X: (point.X / float32(viewportWidth) * 2) - 1,
		Y: 1 - (point.Y / float32(viewportHeight) * 2),
	}
}

func sincos(radians float32) (float32, float32) {
	if radians == 0 {
		return 0, 1
	}
	return float32(math.Sin(float64(radians))), float32(math.Cos(float64(radians)))
}
