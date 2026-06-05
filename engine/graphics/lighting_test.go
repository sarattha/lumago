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
		Ambient:   lmath.Color{R: 0.2, G: 0.3, B: 0.4, A: 1},
		DebugView: DebugViewShadowFactor,
	}.WithDefaults()
	if custom.DebugView != DebugViewShadowFactor || custom.Ambient.R != 0.2 {
		t.Fatalf("custom config was not preserved: %+v", custom)
	}
}

func TestDebugViewString(t *testing.T) {
	tests := map[DebugView2D]string{
		DebugViewFinalComposite: "final",
		DebugViewSceneColor:     "scene_color",
		DebugViewSceneNormal:    "scene_normal",
		DebugViewLightBuffer:    "light_buffer",
		DebugViewShadowFactor:   "shadow_factor",
		DebugView2D(42):         "final",
	}
	for view, want := range tests {
		if got := view.String(); got != want {
			t.Fatalf("view %d string=%q, want %q", view, got, want)
		}
	}
}
