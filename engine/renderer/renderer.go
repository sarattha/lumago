package renderer

import "github.com/sarattha/lumago/engine/graphics"

type Renderer interface {
	BeginFrame(camera graphics.Camera2D) error
	SubmitSprites(commands []graphics.SpriteDrawCommand) error
	SubmitLights(lights []graphics.Light2D) error
	SubmitOccluders(occluders []graphics.Occluder2D) error
	EndFrame() error
	Resize(width, height int) error
	Close() error
}
