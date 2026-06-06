package renderer

import (
	"fmt"
	"time"

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
	r.stats = FrameStats{}
	fmt.Printf("frame=%d camera=(%.2f, %.2f) zoom=%.2f\n", r.frame, camera.Position.X, camera.Position.Y, camera.Zoom)
	return nil
}

func (r *NopRenderer) SetCPUFrameTime(duration time.Duration) {
	r.stats.CPUFrameTime = duration
}

func (r *NopRenderer) SetHotPathAllocBytes(bytes uint64) {
	r.stats.HotPathAllocBytes = bytes
}

func (r *NopRenderer) SubmitSpriteBatch(batch graphics.SpriteBatch) error {
	r.stats.Sprites = batch.Stats.SpriteCount
	r.stats.DrawCalls = batch.Stats.DrawCalls
	r.stats.Vertices = batch.Stats.VertexCount
	r.stats.Indices = batch.Stats.IndexCount
	fmt.Printf("sprites=%d draws=%d vertices=%d indices=%d\n", r.stats.Sprites, r.stats.DrawCalls, r.stats.Vertices, r.stats.Indices)
	return nil
}

func (r *NopRenderer) ConfigureLighting(config graphics.LightingConfig2D) error {
	config = config.WithDefaults()
	r.stats.DebugView = config.DebugView
	fmt.Printf("ambient=(%.2f, %.2f, %.2f) debug=%s shadow_mode=%s\n", config.Ambient.R, config.Ambient.G, config.Ambient.B, config.DebugView, config.ShadowMode)
	return nil
}

func (r *NopRenderer) SubmitLights(lights []graphics.Light2D) error {
	r.stats.Lights = len(lights)
	fmt.Printf("lights=%d\n", len(lights))
	return nil
}

func (r *NopRenderer) SubmitOccluders(occluders []graphics.Occluder2D) error {
	r.stats.Occluders = len(occluders)
	fmt.Printf("occluders=%d\n", len(occluders))
	return nil
}

func (r *NopRenderer) EndFrame() error {
	fmt.Printf("present=nop cpu_ms=%.3f alloc_bytes=%d sprites=%d draws=%d lights=%d occluders=%d debug=%s\n",
		float64(r.stats.CPUFrameTime.Microseconds())/1000,
		r.stats.HotPathAllocBytes,
		r.stats.Sprites,
		r.stats.DrawCalls,
		r.stats.Lights,
		r.stats.Occluders,
		r.stats.DebugView,
	)
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
