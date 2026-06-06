package main

import (
	"fmt"
	"time"

	"github.com/sarattha/lumago/engine/graphics"
	"github.com/sarattha/lumago/engine/renderer"
)

type diagnosticsRenderer struct {
	next          renderer.Renderer
	every         int
	frame         int
	lastFrameTime time.Duration
	frameReady    bool
}

func newDiagnosticsRenderer(next renderer.Renderer, every int) renderer.Renderer {
	if every <= 0 {
		every = 60
	}
	return &diagnosticsRenderer{next: next, every: every}
}

func (r *diagnosticsRenderer) BeginFrame(camera graphics.Camera2D) error {
	return r.next.BeginFrame(camera)
}

func (r *diagnosticsRenderer) SetCPUFrameTime(duration time.Duration) {
	r.lastFrameTime = duration
	r.next.SetCPUFrameTime(duration)
	if r.frameReady {
		r.printOverlay()
		r.frameReady = false
	}
}

func (r *diagnosticsRenderer) SetHotPathAllocBytes(bytes uint64) {
	r.next.SetHotPathAllocBytes(bytes)
}

func (r *diagnosticsRenderer) SubmitSpriteBatch(batch graphics.SpriteBatch) error {
	return r.next.SubmitSpriteBatch(batch)
}

func (r *diagnosticsRenderer) ConfigureLighting(config graphics.LightingConfig2D) error {
	return r.next.ConfigureLighting(config)
}

func (r *diagnosticsRenderer) SubmitLights(lights []graphics.Light2D) error {
	return r.next.SubmitLights(lights)
}

func (r *diagnosticsRenderer) SubmitOccluders(occluders []graphics.Occluder2D) error {
	return r.next.SubmitOccluders(occluders)
}

func (r *diagnosticsRenderer) Stats() renderer.FrameStats {
	return r.next.Stats()
}

func (r *diagnosticsRenderer) EndFrame() error {
	if err := r.next.EndFrame(); err != nil {
		return err
	}
	r.frame++
	r.frameReady = true
	return nil
}

func (r *diagnosticsRenderer) printOverlay() {
	if r.frame%r.every == 0 || r.frame == 1 {
		stats := r.next.Stats()
		fmt.Printf("overlay frame=%d cpu_ms=%.3f gpu_ms=%.3f alloc_bytes=%d draws=%d sprites=%d lights=%d occluders=%d debug=%s\n",
			r.frame,
			float64(r.lastFrameTime.Microseconds())/1000,
			float64(stats.GPUFrameTime.Microseconds())/1000,
			stats.HotPathAllocBytes,
			stats.DrawCalls,
			stats.Sprites,
			stats.Lights,
			stats.Occluders,
			stats.DebugView,
		)
		for _, pass := range stats.Passes {
			fmt.Printf("overlay_pass frame=%d name=%s cpu_ms=%.3f gpu_ms=%.3f\n",
				r.frame,
				pass.Name,
				float64(pass.CPUTime.Microseconds())/1000,
				float64(pass.GPUTime.Microseconds())/1000,
			)
		}
	}
}

func (r *diagnosticsRenderer) Resize(width, height int) error {
	return r.next.Resize(width, height)
}

func (r *diagnosticsRenderer) Close() error {
	return r.next.Close()
}
