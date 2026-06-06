package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/sarattha/lumago/engine/graphics"
)

const defaultConfigPath = "examples/trex_runner/lumago.conf"

type runnerConfig struct {
	Width            int
	Height           int
	Renderer         string
	DebugView        graphics.DebugView2D
	ShadowMode       graphics.ShadowMode2D
	Development      bool
	ShaderReload     bool
	DebugOverlay     bool
	FrameLimit       int
	DiagnosticsEvery int
	ShaderDirectory  string
	VulkanValidation bool
}

func defaultRunnerConfig() runnerConfig {
	return runnerConfig{
		Width:            runnerTargetWidth,
		Height:           runnerTargetHeight,
		Renderer:         "vulkan",
		DebugView:        graphics.DebugViewFinalComposite,
		ShadowMode:       graphics.ShadowModeHardMaps,
		DebugOverlay:     true,
		DiagnosticsEvery: 60,
		ShaderDirectory:  "shaders/bin",
	}
}

func loadRunnerConfig(path string) runnerConfig {
	config := defaultRunnerConfig()
	applyRunnerEnvironment(&config)
	file, err := os.Open(path)
	if err != nil {
		return config
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		applyRunnerConfigValue(&config, strings.TrimSpace(key), strings.TrimSpace(value))
	}
	applyRunnerEnvironment(&config)
	return config
}

func applyRunnerEnvironment(config *runnerConfig) {
	if value := os.Getenv("LUMAGO_RENDERER"); value != "" {
		config.Renderer = value
	}
	if value := os.Getenv("LUMAGO_DEBUG_VIEW"); value != "" {
		config.DebugView = parseRunnerDebugView(value)
	}
	if value := os.Getenv("LUMAGO_SHADOW_MODE"); value != "" {
		config.ShadowMode = parseRunnerShadowMode(value)
	}
	if value := os.Getenv("LUMAGO_FRAME_LIMIT"); value != "" {
		config.FrameLimit = parseRunnerInt(value, config.FrameLimit)
	}
	if value := os.Getenv("LUMAGO_VULKAN_VALIDATION"); value != "" {
		config.VulkanValidation = parseRunnerBool(value)
	}
}

func applyRunnerConfigValue(config *runnerConfig, key string, value string) {
	switch strings.ToLower(key) {
	case "window_width":
		config.Width = parseRunnerInt(value, config.Width)
	case "window_height":
		config.Height = parseRunnerInt(value, config.Height)
	case "renderer":
		config.Renderer = value
	case "debug_view":
		config.DebugView = parseRunnerDebugView(value)
	case "shadow_mode":
		config.ShadowMode = parseRunnerShadowMode(value)
	case "development":
		config.Development = parseRunnerBool(value)
	case "shader_reload":
		config.ShaderReload = parseRunnerBool(value)
	case "debug_overlay":
		config.DebugOverlay = parseRunnerBool(value)
	case "frame_limit":
		config.FrameLimit = parseRunnerInt(value, config.FrameLimit)
	case "diagnostics_every":
		config.DiagnosticsEvery = parseRunnerInt(value, config.DiagnosticsEvery)
	case "shader_directory":
		config.ShaderDirectory = value
	case "vulkan_validation":
		config.VulkanValidation = parseRunnerBool(value)
	}
}

func parseRunnerDebugView(value string) graphics.DebugView2D {
	switch strings.ToLower(value) {
	case "color", "scene_color":
		return graphics.DebugViewSceneColor
	case "normal", "scene_normal":
		return graphics.DebugViewSceneNormal
	case "light", "light_buffer":
		return graphics.DebugViewLightBuffer
	case "shadow", "shadow_factor":
		return graphics.DebugViewShadowFactor
	case "sdf":
		return graphics.DebugViewSDF
	default:
		return graphics.DebugViewFinalComposite
	}
}

func parseRunnerShadowMode(value string) graphics.ShadowMode2D {
	switch strings.ToLower(value) {
	case "sdf", "sdf_experimental":
		return graphics.ShadowModeSDFExperimental
	default:
		return graphics.ShadowModeHardMaps
	}
}

func parseRunnerBool(value string) bool {
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseRunnerInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
