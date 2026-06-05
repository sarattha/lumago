package renderer

import (
	"fmt"

	"github.com/sarattha/lumago/engine/graphics"
)

type NopRenderer struct {
	frame int
	stats FrameStats
}

func NewNopRenderer() *NopRenderer {
	return &NopRenderer{}
}

func (r *NopRenderer) BeginFrame(camera graphics.Camera2D) error {
	r.frame++
	fmt.Printf("frame=%d camera=(%.2f, %.2f) zoom=%.2f\n", r.frame, camera.Position.X, camera.Position.Y, camera.Zoom)
	return nil
}

func (r *NopRenderer) SubmitSpriteBatch(batch graphics.SpriteBatch) error {
	r.stats = FrameStats{
		Sprites:   batch.Stats.SpriteCount,
		DrawCalls: batch.Stats.DrawCalls,
		Vertices:  batch.Stats.VertexCount,
		Indices:   batch.Stats.IndexCount,
	}
	fmt.Printf("sprites=%d draws=%d vertices=%d indices=%d\n", r.stats.Sprites, r.stats.DrawCalls, r.stats.Vertices, r.stats.Indices)
	return nil
}

func (r *NopRenderer) SubmitLights(lights []graphics.Light2D) error {
	fmt.Printf("lights=%d\n", len(lights))
	return nil
}

func (r *NopRenderer) SubmitOccluders(occluders []graphics.Occluder2D) error {
	fmt.Printf("occluders=%d\n", len(occluders))
	return nil
}

func (r *NopRenderer) EndFrame() error {
	fmt.Println("present=nop")
	return nil
}

func (r *NopRenderer) Stats() FrameStats {
	return r.stats
}

func (r *NopRenderer) Resize(width, height int) error {
	fmt.Printf("resize=%dx%d renderer=nop\n", width, height)
	return nil
}

func (r *NopRenderer) Close() error {
	return nil
}
