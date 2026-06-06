package app

import (
	"testing"
	"time"

	"github.com/sarattha/lumago/engine/graphics"
	"github.com/sarattha/lumago/engine/input"
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

func TestRunFrameCPUTimeIncludesEndFrameWork(t *testing.T) {
	game := NewGame(Config{Width: 64, Height: 64})
	world := scene.New()
	world.AddSprite(graphics.SpriteDrawCommand{Sprite: graphics.Sprite{Color: lmath.White()}})
	game.SetScene(world)
	fake := &statsRenderer{endFrameDelay: 20 * time.Millisecond}
	game.SetRenderer(fake)

	if err := game.Run(); err != nil {
		t.Fatal(err)
	}
	if got := game.Stats().CPUFrameTime; got < fake.endFrameDelay {
		t.Fatalf("cpu frame time=%s, want at least EndFrame delay %s", got, fake.endFrameDelay)
	}
}

func TestRunFrameUsesConfiguredViewportForSpriteBatch(t *testing.T) {
	game := NewGame(Config{Width: 1920, Height: 1080})
	world := scene.New()
	world.AddSprite(graphics.SpriteDrawCommand{Sprite: graphics.Sprite{Color: lmath.White()}})
	game.SetScene(world)
	fake := &statsRenderer{}
	game.SetRenderer(fake)
	game.SetWindow(fixedFramebufferWindow{width: 3024, height: 1964})

	if err := game.runFrame(); err != nil {
		t.Fatal(err)
	}
	if fake.batch.Stats.ViewportWidth != 1920 || fake.batch.Stats.ViewportHeight != 1080 {
		t.Fatalf("batch viewport=%dx%d, want configured 1920x1080", fake.batch.Stats.ViewportWidth, fake.batch.Stats.ViewportHeight)
	}
}

type statsRenderer struct {
	stats           renderer.FrameStats
	batch           graphics.SpriteBatch
	hotPathAllocSet bool
	endFrameDelay   time.Duration
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
	r.batch = batch
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
func (r *statsRenderer) EndFrame() error {
	if r.endFrameDelay > 0 {
		time.Sleep(r.endFrameDelay)
	}
	return nil
}
func (r *statsRenderer) Resize(width, height int) error {
	return nil
}
func (r *statsRenderer) Close() error { return nil }

type fixedFramebufferWindow struct {
	width  int
	height int
}

func (w fixedFramebufferWindow) ShouldClose() bool           { return true }
func (w fixedFramebufferWindow) PollEvents()                 {}
func (w fixedFramebufferWindow) FramebufferSize() (int, int) { return w.width, w.height }
func (w fixedFramebufferWindow) WaitForFramebuffer()         {}
func (w fixedFramebufferWindow) KeyDown(input.Key) bool      { return false }
func (w fixedFramebufferWindow) Close()                      {}
