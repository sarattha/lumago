package scene

import "github.com/sarattha/lumago/engine/graphics"

type Scene struct {
	sprites   []graphics.SpriteDrawCommand
	batch     graphics.SpriteBatch
	lights    []graphics.Light2D
	occluders []graphics.Occluder2D
	camera    graphics.Camera2D
}

func New() *Scene {
	return &Scene{
		camera: graphics.DefaultCamera2D(),
	}
}

func (s *Scene) AddSprite(sprite graphics.SpriteDrawCommand) {
	s.sprites = append(s.sprites, sprite)
}

func (s *Scene) AddLight(light graphics.Light2D) {
	s.lights = append(s.lights, light)
}

func (s *Scene) AddOccluder(occluder graphics.Occluder2D) {
	s.occluders = append(s.occluders, occluder)
}

func (s *Scene) Camera() graphics.Camera2D {
	return s.camera
}

func (s *Scene) SetCamera(camera graphics.Camera2D) {
	s.camera = camera
}

func (s *Scene) Sprites() []graphics.SpriteDrawCommand {
	return s.sprites
}

func (s *Scene) BuildSpriteBatch(width, height int) graphics.SpriteBatch {
	s.batch.Build(s.sprites, s.camera, width, height)
	return s.batch
}

func (s *Scene) Lights() []graphics.Light2D {
	return s.lights
}

func (s *Scene) Occluders() []graphics.Occluder2D {
	return s.occluders
}
