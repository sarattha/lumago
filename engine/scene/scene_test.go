package scene

import (
	"testing"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
)

func TestBuildSpriteBatchReusesStorage(t *testing.T) {
	s := New()
	s.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Src:   lmath.Rect{W: 8, H: 8},
			Color: lmath.White(),
		},
	})

	first := s.BuildSpriteBatch(100, 100)
	commandCap := cap(first.Commands)
	vertexCap := cap(first.Vertices)
	indexCap := cap(first.Indices)

	second := s.BuildSpriteBatch(100, 100)
	if cap(second.Commands) != commandCap || cap(second.Vertices) != vertexCap || cap(second.Indices) != indexCap {
		t.Fatalf("batch storage was not reused")
	}
}
