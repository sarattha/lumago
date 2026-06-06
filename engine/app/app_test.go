package app

import (
	"testing"
	"time"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	"github.com/sarattha/lumago/engine/renderer"
	"github.com/sarattha/lumago/engine/scene"
)

func TestRunFrameRecordsFrameTimingAndAllocationStats(t *testing.T) {
	game := NewGame(Config{Width: 64, Height: 64})
	world := scene.New()
	world.AddSprite(graphics.SpriteDrawCommand{Sprite: graphics.Sprite{Color: lmath.White()}})
	game.SetScene(world)
	fake := &statsRenderer{}
	game.SetRenderer(fake)

	if err := game.Run(); err != nil {
		t.Fatal(err)
	}
	stats := game.Stats()
	if stats.CPUFrameTime <= 0 {
		t.Fatalf("cpu frame time=%s, want measured duration", stats.CPUFrameTime)
	}
	if !fake.hotPathAllocSet {
		t.Fatalf("hot-path allocation metric was not set")
	}
}

type statsRenderer struct {
	stats           renderer.FrameStats
	hotPathAllocSet bool
}

func (r *statsRenderer) BeginFrame(camera graphics.Camera2D) error { return nil }

func (r *statsRenderer) SetCPUFrameTime(duration time.Duration) {
	r.stats.CPUFrameTime = duration
}

func (r *statsRenderer) SetHotPathAllocBytes(bytes uint64) {
	r.stats.HotPathAllocBytes = bytes
	r.hotPathAllocSet = true
}

func (r *statsRenderer) SubmitSpriteBatch(batch graphics.SpriteBatch) error {
	r.stats.Sprites = batch.Stats.SpriteCount
	return nil
}

func (r *statsRenderer) ConfigureLighting(config graphics.LightingConfig2D) error {
	r.stats.DebugView = config.WithDefaults().DebugView
	return nil
}

func (r *statsRenderer) SubmitLights(lights []graphics.Light2D) error {
	r.stats.Lights = len(lights)
	return nil
}

func (r *statsRenderer) SubmitOccluders(occluders []graphics.Occluder2D) error {
	r.stats.Occluders = len(occluders)
	return nil
}

func (r *statsRenderer) Stats() renderer.FrameStats { return r.stats }
func (r *statsRenderer) EndFrame() error            { return nil }
func (r *statsRenderer) Resize(width, height int) error {
	return nil
}
func (r *statsRenderer) Close() error { return nil }
