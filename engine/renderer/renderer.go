package renderer

import "github.com/sarattha/lumago/engine/graphics"

type FrameStats struct {
	Sprites   int
	Lights    int
	DrawCalls int
	Vertices  int
	Indices   int
}

type Renderer interface {
	BeginFrame(camera graphics.Camera2D) error
	SubmitSpriteBatch(batch graphics.SpriteBatch) error
	ConfigureLighting(config graphics.LightingConfig2D) error
	SubmitLights(lights []graphics.Light2D) error
	SubmitOccluders(occluders []graphics.Occluder2D) error
	Stats() FrameStats
	EndFrame() error
	Resize(width, height int) error
	Close() error
}
