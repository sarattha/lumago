package app

import (
	"errors"
	"runtime"
	"time"

	"github.com/sarattha/lumago/engine/assets"
	"github.com/sarattha/lumago/engine/input"
	"github.com/sarattha/lumago/engine/renderer"
	"github.com/sarattha/lumago/engine/scene"
)

type Config struct {
	Width     int
	Height    int
	Title     string
	FixedStep time.Duration
}

type Window interface {
	ShouldClose() bool
	PollEvents()
	FramebufferSize() (int, int)
	WaitForFramebuffer()
	KeyDown(input.Key) bool
	Close()
}

type Game struct {
	Config     Config
	Assets     *assets.Registry
	scene      *scene.Scene
	renderer   renderer.Renderer
	window     Window
	updateFunc func(time.Duration) error
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

func (g *Game) SetWindow(window Window) {
	g.window = window
}

func (g *Game) SetUpdateFunc(update func(time.Duration) error) {
	g.updateFunc = update
}

func (g *Game) Stats() renderer.FrameStats {
	return g.renderer.Stats()
}

func (g *Game) Run() error {
	if g.scene == nil {
		return errors.New("lumago: no scene set")
	}

	defer g.renderer.Close()

	if g.window == nil {
		return g.runFrame()
	}
	defer g.window.Close()

	step := g.Config.FixedStep
	if step <= 0 {
		step = time.Second / 60
	}

	last := time.Now()
	accumulator := time.Duration(0)
	lastWidth, lastHeight := g.window.FramebufferSize()
	if err := g.renderer.Resize(lastWidth, lastHeight); err != nil {
		return err
	}

	for !g.window.ShouldClose() {
		g.window.PollEvents()
		width, height := g.window.FramebufferSize()
		if width == 0 || height == 0 {
			g.window.WaitForFramebuffer()
			last = time.Now()
			continue
		}
		if width != lastWidth || height != lastHeight {
			if err := g.renderer.Resize(width, height); err != nil {
				return err
			}
			lastWidth, lastHeight = width, height
		}

		now := time.Now()
		accumulator += now.Sub(last)
		last = now
		for accumulator >= step {
			if g.updateFunc != nil {
				if err := g.updateFunc(step); err != nil {
					return err
				}
			}
			accumulator -= step
		}

		if err := g.runFrame(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) runFrame() error {
	frameStart := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)
	if err := g.renderer.BeginFrame(g.scene.Camera()); err != nil {
		return err
	}

	width, height := g.renderViewportSize()
	if err := g.renderer.SubmitSpriteBatch(g.scene.BuildSpriteBatch(width, height)); err != nil {
		return err
	}

	if err := g.renderer.ConfigureLighting(g.scene.LightingConfig()); err != nil {
		return err
	}

	if err := g.renderer.SubmitLights(g.scene.Lights()); err != nil {
		return err
	}

	if err := g.renderer.SubmitOccluders(g.scene.Occluders()); err != nil {
		return err
	}

	err := g.renderer.EndFrame()
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	if memEnd.TotalAlloc >= memStart.TotalAlloc {
		g.renderer.SetHotPathAllocBytes(memEnd.TotalAlloc - memStart.TotalAlloc)
	}
	g.renderer.SetCPUFrameTime(time.Since(frameStart))
	return err
}

func (g *Game) renderViewportSize() (int, int) {
	width, height := g.Config.Width, g.Config.Height
	if (width <= 0 || height <= 0) && g.window != nil {
		width, height = g.window.FramebufferSize()
	}
	return width, height
}
