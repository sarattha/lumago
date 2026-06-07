package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sarattha/lumago/engine/app"
	engineassets "github.com/sarattha/lumago/engine/assets"
	"github.com/sarattha/lumago/engine/input"
	"github.com/sarattha/lumago/engine/platform/desktop"
	"github.com/sarattha/lumago/engine/renderer"
	vulkanrenderer "github.com/sarattha/lumago/engine/renderer/vulkan"
)

var (
	errRunnerFrameLimit = errors.New("t-rex runner frame limit reached")
	errRunnerQuit       = errors.New("t-rex runner quit requested")
)

type keyReader interface {
	KeyDown(input.Key) bool
}

type runnerControls struct {
	previousJump    bool
	previousRestart bool
}

func main() {
	runtime.LockOSThread()

	config := loadRunnerConfig(defaultConfigPath)
	catalog, hotReloader, err := prepareRunnerAssets(config)
	if err != nil {
		panic(err)
	}
	config.AssetCatalog = catalog
	game := app.NewGame(app.Config{
		Width:     config.Width,
		Height:    config.Height,
		Title:     "LumaGo T-Rex Runner",
		FixedStep: time.Second / runnerTargetFPS,
	})

	state := newRunnerState()
	game.SetScene(buildRunnerScene(game, state, config))

	var reader keyReader
	var selectedRenderer renderer.Renderer
	if config.Renderer == "nop" {
		selectedRenderer = renderer.NewNopRenderer()
	} else {
		window, err := desktop.NewWindow(game.Config.Width, game.Config.Height, game.Config.Title)
		if err != nil {
			panic(err)
		}
		reader = window
		vulkanRenderer, err := vulkanrenderer.NewRenderer(vulkanrenderer.Config{
			Window:          window,
			ShaderDirectory: config.ShaderDirectory,
			Validation:      config.VulkanValidation,
			Development:     config.Development || config.ShaderReload,
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

	controls := runnerControls{}
	frame := 0
	game.SetUpdateFunc(func(dt time.Duration) error {
		if hotReloader != nil {
			manifest, changed, err := hotReloader.ReloadChanged()
			if err != nil {
				return err
			}
			if changed {
				config.AssetCatalog = runnerAssetCatalogFromManifestRevision(manifest, hotReloader.BaseDir(), config.AssetCatalog.Revision+1)
			}
		}
		if reader != nil && reader.KeyDown(input.KeyEscape) {
			return errRunnerQuit
		}
		state.Step(float32(dt.Seconds()), controls.read(reader))
		game.SetScene(buildRunnerScene(game, state, config))
		frame++
		if config.FrameLimit > 0 && frame >= config.FrameLimit {
			return errRunnerFrameLimit
		}
		return nil
	})

	err = game.Run()
	if err != nil && err != errRunnerFrameLimit && err != errRunnerQuit {
		panic(err)
	}

	stats := game.Stats()
	observedFPS := 0.0
	if stats.CPUFrameTime > 0 {
		observedFPS = float64(time.Second) / float64(stats.CPUFrameTime)
	}
	fmt.Printf("LumaGo T-Rex runner finished: target=%dx%d@%dfps observed_fps=%.1f score=%d started=%t game_over=%t sprites=%d lights=%d occluders=%d cpu_ms=%.3f alloc_bytes=%d debug=%s\n",
		runnerTargetWidth,
		runnerTargetHeight,
		runnerTargetFPS,
		observedFPS,
		state.Score,
		state.Started,
		state.GameOver,
		stats.Sprites,
		stats.Lights,
		stats.Occluders,
		float64(stats.CPUFrameTime.Microseconds())/1000,
		stats.HotPathAllocBytes,
		stats.DebugView,
	)
}

func prepareRunnerAssets(config runnerConfig) (runnerAssetCatalog, *runnerAssetHotReloader, error) {
	metadataPath := resolveRunnerAssetMetadataPath(config.AssetMetadata)
	catalog, err := loadRunnerAssetCatalog(metadataPath)
	if err != nil {
		return runnerAssetCatalog{}, nil, err
	}
	if !config.Development {
		return catalog, nil, nil
	}
	reloader, manifest, err := engineassets.NewHotReloader(metadataPath)
	if err != nil {
		return runnerAssetCatalog{}, nil, err
	}
	return runnerAssetCatalogFromManifest(manifest, filepath.Dir(metadataPath)), &runnerAssetHotReloader{
		HotReloader: reloader,
		baseDir:     filepath.Dir(metadataPath),
	}, nil
}

type runnerAssetHotReloader struct {
	*engineassets.HotReloader
	baseDir string
}

func (r *runnerAssetHotReloader) BaseDir() string {
	return r.baseDir
}

func (c *runnerControls) read(reader keyReader) runnerInput {
	if reader == nil {
		return runnerInput{}
	}
	jumpHeld := reader.KeyDown(input.KeySpace) || reader.KeyDown(input.KeyUp) || reader.KeyDown(input.KeyW)
	restartHeld := reader.KeyDown(input.KeyR)
	duckHeld := reader.KeyDown(input.KeyDown) || reader.KeyDown(input.KeyS)
	result := runnerInput{
		Start:   jumpHeld && !c.previousJump,
		Jump:    jumpHeld && !c.previousJump,
		Duck:    duckHeld,
		Restart: restartHeld && !c.previousRestart,
	}
	c.previousJump = jumpHeld
	c.previousRestart = restartHeld
	return result
}
