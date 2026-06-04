package app

import (
	"errors"

	"github.com/sarattha/lumago/engine/assets"
	"github.com/sarattha/lumago/engine/renderer"
	"github.com/sarattha/lumago/engine/scene"
)

type Config struct {
	Width  int
	Height int
	Title  string
}

type Game struct {
	Config   Config
	Assets   *assets.Registry
	scene    *scene.Scene
	renderer renderer.Renderer
}

func NewGame(config Config) *Game {
	return &Game{
		Config:   config,
		Assets:   assets.NewRegistry(),
		renderer: renderer.NewNopRenderer(),
	}
}

func (g *Game) SetScene(scene *scene.Scene) {
	g.scene = scene
}

func (g *Game) SetRenderer(renderer renderer.Renderer) {
	g.renderer = renderer
}

func (g *Game) Run() error {
	if g.scene == nil {
		return errors.New("lumago: no scene set")
	}

	if err := g.renderer.BeginFrame(g.scene.Camera()); err != nil {
		return err
	}

	if err := g.renderer.SubmitSprites(g.scene.Sprites()); err != nil {
		return err
	}

	if err := g.renderer.SubmitLights(g.scene.Lights()); err != nil {
		return err
	}

	if err := g.renderer.SubmitOccluders(g.scene.Occluders()); err != nil {
		return err
	}

	return g.renderer.EndFrame()
}
