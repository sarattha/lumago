package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sarattha/lumago/engine/graphics"
)

func TestLoadDemoConfigReadsWindowRendererDebugAndDevelopmentToggles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lumago.conf")
	text := `window_width=1920
window_height=1080
renderer=nop
debug_view=sdf
shadow_mode=sdf
development=true
shader_reload=true
debug_overlay=true
debug_labels=true
frame_limit=2
diagnostics_every=3
shader_directory=testdata/shaders
vulkan_validation=true
`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}

	config := loadDemoConfig(path)
	if config.Width != 1920 || config.Height != 1080 {
		t.Fatalf("window=%dx%d, want 1920x1080", config.Width, config.Height)
	}
	if config.Renderer != "nop" || config.DebugView != graphics.DebugViewSDF || config.ShadowMode != graphics.ShadowModeSDFExperimental {
		t.Fatalf("renderer/debug/shadow=%q/%s/%s", config.Renderer, config.DebugView, config.ShadowMode)
	}
	if !config.Development || !config.ShaderReload || !config.DebugOverlay || !config.DebugLabels || !config.VulkanValidation {
		t.Fatalf("development toggles were not enabled: %+v", config)
	}
	if config.FrameLimit != 2 || config.DiagnosticsEvery != 3 || config.ShaderDirectory != "testdata/shaders" {
		t.Fatalf("frame/shader config=%+v", config)
	}
}

func TestEnvironmentOverridesDemoConfig(t *testing.T) {
	t.Setenv("LUMAGO_RENDERER", "nop")
	t.Setenv("LUMAGO_DEBUG_VIEW", "shadow")
	t.Setenv("LUMAGO_SHADOW_MODE", "sdf")
	t.Setenv("LUMAGO_FRAME_LIMIT", "1")
	t.Setenv("LUMAGO_VULKAN_VALIDATION", "1")

	config := loadDemoConfig(filepath.Join(t.TempDir(), "missing.conf"))
	if config.Renderer != "nop" || config.DebugView != graphics.DebugViewShadowFactor || config.ShadowMode != graphics.ShadowModeSDFExperimental {
		t.Fatalf("environment overrides missing: %+v", config)
	}
	if config.FrameLimit != 1 || !config.VulkanValidation {
		t.Fatalf("environment numeric/bool overrides missing: %+v", config)
	}
}
