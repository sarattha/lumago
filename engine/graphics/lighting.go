package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type DebugView2D uint8

const (
	DebugViewFinalComposite DebugView2D = iota
	DebugViewSceneColor
	DebugViewSceneNormal
	DebugViewLightBuffer
	DebugViewShadowFactor
	DebugViewSDF
)

type ShadowMode2D uint8

const (
	ShadowModeHardMaps ShadowMode2D = iota
	ShadowModeSDFExperimental
)

type LightingConfig2D struct {
	Ambient    lmath.Color
	DebugView  DebugView2D
	ShadowMode ShadowMode2D
}

func DefaultLightingConfig2D() LightingConfig2D {
	return LightingConfig2D{
		Ambient:   lmath.Color{R: 0.12, G: 0.12, B: 0.14, A: 1},
		DebugView: DebugViewFinalComposite,
	}
}

func (c LightingConfig2D) WithDefaults() LightingConfig2D {
	if c.Ambient.A == 0 && c.Ambient.R == 0 && c.Ambient.G == 0 && c.Ambient.B == 0 {
		c.Ambient = DefaultLightingConfig2D().Ambient
	}
	if c.DebugView > DebugViewSDF {
		c.DebugView = DebugViewFinalComposite
	}
	if c.ShadowMode > ShadowModeSDFExperimental {
		c.ShadowMode = ShadowModeHardMaps
	}
	return c
}

func (v DebugView2D) String() string {
	switch v {
	case DebugViewFinalComposite:
		return "final"
	case DebugViewSceneColor:
		return "scene_color"
	case DebugViewSceneNormal:
		return "scene_normal"
	case DebugViewLightBuffer:
		return "light_buffer"
	case DebugViewShadowFactor:
		return "shadow_factor"
	case DebugViewSDF:
		return "sdf"
	default:
		return "final"
	}
}

func (m ShadowMode2D) String() string {
	switch m {
	case ShadowModeHardMaps:
		return "hard_maps"
	case ShadowModeSDFExperimental:
		return "sdf_experimental"
	default:
		return "hard_maps"
	}
}
