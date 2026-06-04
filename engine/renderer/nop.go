package renderer

import (
	"fmt"

	"github.com/sarattha/lumago/engine/graphics"
)

type NopRenderer struct {
	frame int
}

func NewNopRenderer() *NopRenderer {
	return &NopRenderer{}
}

func (r *NopRenderer) BeginFrame(camera graphics.Camera2D) error {
	r.frame++
	fmt.Printf("frame=%d camera=(%.2f, %.2f) zoom=%.2f\n", r.frame, camera.Position.X, camera.Position.Y, camera.Zoom)
	return nil
}

func (r *NopRenderer) SubmitSprites(commands []graphics.SpriteDrawCommand) error {
	fmt.Printf("sprites=%d\n", len(commands))
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
