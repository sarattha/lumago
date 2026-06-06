package vulkan

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	erenderer "github.com/sarattha/lumago/engine/renderer"
	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

func TestRecordPassTimingCapturesNamedCPUTiming(t *testing.T) {
	r := &Renderer{}
	r.recordPassTiming("color", func() {})
	r.recordPassTiming("normal", func() {})
	r.recordPassTiming("shadow", func() {})
	r.recordPassTiming("light", func() {})
	r.recordPassTiming("sdf", func() {})
	r.recordPassTiming("composite", func() {})

	got := make([]string, 0, len(r.passTimings))
	for _, timing := range r.passTimings {
		got = append(got, timing.Name)
	}
	want := []string{"color", "normal", "shadow", "light", "sdf", "composite"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pass %d=%q, want %q", i, got[i], want[i])
		}
	}
}

func TestStatsClonesPassTimingSlice(t *testing.T) {
	r := &Renderer{
		stats:       erenderer.FrameStats{Sprites: 1},
		passTimings: []erenderer.PassTiming{{Name: "composite", CPUTime: time.Microsecond}},
	}

	stats := r.Stats()
	stats.Passes[0].Name = "mutated"
	if r.passTimings[0].Name != "composite" {
		t.Fatalf("stats pass timings alias renderer storage")
	}
}

func TestViewportExtentForBatchPrefersLogicalBatchViewport(t *testing.T) {
	batch := graphics.SpriteBatch{
		Stats: graphics.SpriteBatchStats{
			ViewportWidth:  1920,
			ViewportHeight: 1080,
		},
	}
	got := viewportExtentForBatch(batch, vk.Extent2D{Width: 3024, Height: 1964})
	if got.Width != 1920 || got.Height != 1080 {
		t.Fatalf("viewport extent=%dx%d, want logical 1920x1080", got.Width, got.Height)
	}
}

func TestViewportExtentForBatchFallsBackToSwapchainExtent(t *testing.T) {
	got := viewportExtentForBatch(graphics.SpriteBatch{}, vk.Extent2D{Width: 3024, Height: 1964})
	if got.Width != 3024 || got.Height != 1964 {
		t.Fatalf("fallback extent=%dx%d, want swapchain extent", got.Width, got.Height)
	}
}

func TestShaderFilesChangedReportsDevelopmentShaderEdits(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"quad.vert.spv", "quad.frag.spv"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte{0, 0, 0, 0}, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	r := &Renderer{shaderDirectory: dir, shaderModTimes: make(map[string]time.Time)}

	changed, err := r.shaderFilesChanged("quad.vert.spv", "quad.frag.spv")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatalf("first shader scan should establish baseline")
	}

	time.Sleep(time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, "quad.frag.spv"), []byte{1, 0, 0, 0}, 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err = r.shaderFilesChanged("quad.vert.spv", "quad.frag.spv")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatalf("shader edit was not detected")
	}
}

func TestLitSpriteBatchHotPathReusesStorageWithoutHeapAllocation(t *testing.T) {
	batch := hotPathBatch(64)
	lights := []graphics.Light2D{{
		Position:  lmath.Vec2{X: 50, Y: 50},
		Radius:    100,
		Color:     lmath.White(),
		Intensity: 1,
		Falloff:   1,
	}}
	config := graphics.LightingConfig2D{Ambient: lmath.Color{A: 1}}
	dst, vertices, indices := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, lights, nil, sdfTexture{}, config, vk.Extent2D{Width: 256, Height: 256})

	allocs := testing.AllocsPerRun(50, func() {
		dst, vertices, indices = litSpriteBatchForLighting(dst, vertices, indices, batch, lights, nil, sdfTexture{}, config, vk.Extent2D{Width: 256, Height: 256})
	})
	if allocs != 0 {
		t.Fatalf("lit sprite hot path allocs/run=%f, want 0", allocs)
	}
}

func hotPathBatch(count int) graphics.SpriteBatch {
	commands := make([]graphics.SpriteDrawCommand, count)
	for i := range commands {
		commands[i] = graphics.SpriteDrawCommand{
			Sprite: graphics.Sprite{
				Src:   lmath.Rect{W: 8, H: 8},
				Color: lmath.White(),
			},
			Transform: graphics.Transform2D{
				Position: lmath.Vec2{X: float32(i % 8 * 12), Y: float32(i / 8 * 12)},
				Scale:    lmath.Vec2{X: 1, Y: 1},
			},
		}
	}
	var batch graphics.SpriteBatch
	batch.Build(commands, graphics.DefaultCamera2D(), 256, 256)
	return batch
}
