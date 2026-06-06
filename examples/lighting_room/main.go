package main

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/platform/desktop"
	"github.com/sarattha/lumago/engine/renderer"
	vulkanrenderer "github.com/sarattha/lumago/engine/renderer/vulkan"
)

var errFrameLimit = errors.New("lighting room frame limit reached")

func main() {
	runtime.LockOSThread()

	config := loadDemoConfig(defaultConfigPath)
	game := app.NewGame(app.Config{
		Width:     config.Width,
		Height:    config.Height,
		Title:     "LumaGo Lighting Room Example",
		FixedStep: time.Second / acceptanceTargetFPS,
	})

	world := buildLightingRoom(game, config)
	game.SetScene(world)
	lightTime := float32(0)
	frame := 0
	game.SetUpdateFunc(func(dt time.Duration) error {
		lightTime += float32(dt.Seconds())
		updateLights(world, lightTime)
		frame++
		if config.FrameLimit > 0 && frame >= config.FrameLimit {
			return errFrameLimit
		}
		return nil
	})

	var selectedRenderer renderer.Renderer
	if config.Renderer == "nop" {
		selectedRenderer = renderer.NewNopRenderer()
	} else {
		window, err := desktop.NewWindow(game.Config.Width, game.Config.Height, game.Config.Title)
		if err != nil {
			panic(err)
		}
		vulkanRenderer, err := vulkanrenderer.NewRenderer(vulkanrenderer.Config{
			Window:          window,
			ShaderDirectory: config.ShaderDirectory,
			Validation:      config.VulkanValidation,
			Development:     config.Development || config.ShaderReload,
			DebugLabels:     config.DebugLabels,
		})
		if err != nil {
			window.Close()
			panic(err)
		}
		game.SetWindow(window)
		selectedRenderer = vulkanRenderer
	}
	if config.DebugOverlay {
		selectedRenderer = newDiagnosticsRenderer(selectedRenderer, config.DiagnosticsEvery)
	}
	game.SetRenderer(selectedRenderer)

	err := game.Run()
	if err != nil && err != errFrameLimit {
		panic(err)
	}

	stats := game.Stats()
	observedFPS := 0.0
	if stats.CPUFrameTime > 0 {
		observedFPS = float64(time.Second) / float64(stats.CPUFrameTime)
	}
	fmt.Printf("LumaGo lighting room finished: target=%dx%d@%dfps observed_fps=%.1f sprites=%d materials=%d lights=%d shadow_lights=%d occluders=%d cpu_ms=%.3f alloc_bytes=%d debug=%s\n",
		acceptanceTargetWidth,
		acceptanceTargetHeight,
		acceptanceTargetFPS,
		observedFPS,
		stats.Sprites,
		acceptanceMaterials,
		stats.Lights,
		acceptanceShadowLights,
		stats.Occluders,
		float64(stats.CPUFrameTime.Microseconds())/1000,
		stats.HotPathAllocBytes,
		stats.DebugView,
	)
}
