package graphics

import (
	"testing"

	lmath "github.com/sarattha/lumago/engine/math"
)

func TestLightingConfigWithDefaults(t *testing.T) {
	got := (LightingConfig2D{DebugView: DebugView2D(99)}).WithDefaults()

	if got.DebugView != DebugViewFinalComposite {
		t.Fatalf("debug view=%v, want final", got.DebugView)
	}
	if got.Ambient != DefaultLightingConfig2D().Ambient {
		t.Fatalf("ambient=%+v, want default", got.Ambient)
	}

	custom := LightingConfig2D{
		Ambient:    lmath.Color{R: 0.2, G: 0.3, B: 0.4, A: 1},
		DebugView:  DebugViewShadowFactor,
		ShadowMode: ShadowModeSDFExperimental,
	}.WithDefaults()
	if custom.DebugView != DebugViewShadowFactor || custom.Ambient.R != 0.2 || custom.ShadowMode != ShadowModeSDFExperimental {
		t.Fatalf("custom config was not preserved: %+v", custom)
	}

	invalidMode := (LightingConfig2D{ShadowMode: ShadowMode2D(99)}).WithDefaults()
	if invalidMode.ShadowMode != ShadowModeHardMaps {
		t.Fatalf("invalid shadow mode=%v, want hard maps", invalidMode.ShadowMode)
	}
}

func TestDebugViewString(t *testing.T) {
	tests := map[DebugView2D]string{
		DebugViewFinalComposite: "final",
		DebugViewSceneColor:     "scene_color",
		DebugViewSceneNormal:    "scene_normal",
		DebugViewLightBuffer:    "light_buffer",
		DebugViewShadowFactor:   "shadow_factor",
		DebugViewSDF:            "sdf",
		DebugView2D(42):         "final",
	}
	for view, want := range tests {
		if got := view.String(); got != want {
			t.Fatalf("view %d string=%q, want %q", view, got, want)
		}
	}
}

func TestShadowModeString(t *testing.T) {
	tests := map[ShadowMode2D]string{
		ShadowModeHardMaps:        "hard_maps",
		ShadowModeSDFExperimental: "sdf_experimental",
		ShadowMode2D(42):          "hard_maps",
	}
	for mode, want := range tests {
		if got := mode.String(); got != want {
			t.Fatalf("mode %d string=%q, want %q", mode, got, want)
		}
	}
}
