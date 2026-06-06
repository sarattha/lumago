package renderer

import (
	"time"

	"github.com/sarattha/lumago/engine/graphics"
)

type FrameStats struct {
	Sprites           int
	Lights            int
	Occluders         int
	DrawCalls         int
	Vertices          int
	Indices           int
	CPUFrameTime      time.Duration
	GPUFrameTime      time.Duration
	HotPathAllocBytes uint64
	DebugView         graphics.DebugView2D
	Passes            []PassTiming
}

type PassTiming struct {
	Name    string
	CPUTime time.Duration
	GPUTime time.Duration
}

func ClonePassTimings(passTimings []PassTiming) []PassTiming {
	if len(passTimings) == 0 {
		return nil
	}
	cloned := make([]PassTiming, len(passTimings))
	copy(cloned, passTimings)
	return cloned
}

type Renderer interface {
	BeginFrame(camera graphics.Camera2D) error
	SetCPUFrameTime(duration time.Duration)
	SetHotPathAllocBytes(bytes uint64)
	SubmitSpriteBatch(batch graphics.SpriteBatch) error
	ConfigureLighting(config graphics.LightingConfig2D) error
	SubmitLights(lights []graphics.Light2D) error
	SubmitOccluders(occluders []graphics.Occluder2D) error
	Stats() FrameStats
	EndFrame() error
	Resize(width, height int) error
	Close() error
}
