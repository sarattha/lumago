package graphics

import (
	"math"
	"testing"

	lmath "github.com/sarattha/lumago/engine/math"
)

func TestSortSpriteCommands(t *testing.T) {
	commands := []SpriteDrawCommand{
		command(1, 0, 4),
		command(0, 9, 3),
		command(0, 2, 2),
		command(0, 2, 1),
	}

	SortSpriteCommands(commands)

	want := []TextureID{1, 2, 3, 4}
	for i, command := range commands {
		if command.Sprite.Material.Albedo != want[i] {
			t.Fatalf("command %d albedo=%d, want %d", i, command.Sprite.Material.Albedo, want[i])
		}
	}
}

func TestSpriteBatchBuildsOneDraw(t *testing.T) {
	var batch SpriteBatch
	batch.Build([]SpriteDrawCommand{
		{
			Sprite: Sprite{
				Src:   lmath.Rect{W: 20, H: 10},
				Color: lmath.White(),
			},
			Transform: Transform2D{
				Position: lmath.Vec2{X: 100, Y: 50},
				Scale:    lmath.Vec2{X: 1, Y: 1},
			},
		},
		{
			Sprite: Sprite{
				Src:   lmath.Rect{W: 10, H: 10},
				Color: lmath.White(),
			},
			Transform: Transform2D{
				Position: lmath.Vec2{X: 200, Y: 50},
				Scale:    lmath.Vec2{X: 1, Y: 1},
			},
		},
	}, DefaultCamera2D(), 400, 200)

	if batch.Stats.SpriteCount != 2 || batch.Stats.DrawCalls != 1 {
		t.Fatalf("stats=%+v, want 2 sprites and 1 draw", batch.Stats)
	}
	if len(batch.Vertices) != 8 || len(batch.Indices) != 12 {
		t.Fatalf("geometry vertices=%d indices=%d", len(batch.Vertices), len(batch.Indices))
	}
	if batch.Indices[6] != 4 || batch.Indices[11] != 4 {
		t.Fatalf("second sprite indices=%v, want base vertex 4", batch.Indices[6:12])
	}
}

func TestCameraViewMatrix(t *testing.T) {
	camera := Camera2D{
		Position: lmath.Vec2{X: 10, Y: 20},
		Zoom:     2,
	}

	got := camera.ViewMatrix().TransformPoint(lmath.Vec2{X: 12, Y: 23})
	if !near(got.X, 4) || !near(got.Y, 6) {
		t.Fatalf("transformed point=%+v, want (4, 6)", got)
	}
}

func TestTextureAtlasUVRect(t *testing.T) {
	atlas := NewTextureAtlas()
	atlas.AddPage(TextureInfo{ID: 7, Width: 256, Height: 128})
	atlas.AddFrame("hero", 7, lmath.Rect{X: 64, Y: 32, W: 32, H: 16})

	frame, ok := atlas.Frame("hero")
	if !ok {
		t.Fatal("missing atlas frame")
	}
	uv, ok := atlas.UVRect(frame)
	if !ok {
		t.Fatal("missing atlas uv rect")
	}
	if !near(uv.X, 0.25) || !near(uv.Y, 0.25) || !near(uv.W, 0.125) || !near(uv.H, 0.125) {
		t.Fatalf("uv=%+v", uv)
	}
}

func command(layer int, z float32, texture TextureID) SpriteDrawCommand {
	return SpriteDrawCommand{
		Sprite: Sprite{
			Material: Material2D{Albedo: texture},
			Src:      lmath.Rect{W: 1, H: 1},
			Color:    lmath.White(),
		},
		Transform: Transform2D{Z: z},
		Layer:     layer,
	}
}

func near(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.0001
}
