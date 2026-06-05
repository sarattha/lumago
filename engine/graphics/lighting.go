package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type DebugView2D uint8

const (
	DebugViewFinalComposite DebugView2D = iota
	DebugViewSceneColor
	DebugViewSceneNormal
	DebugViewLightBuffer
	DebugViewShadowFactor
)

type LightingConfig2D struct {
	Ambient   lmath.Color
	DebugView DebugView2D
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
	if c.DebugView > DebugViewShadowFactor {
		c.DebugView = DebugViewFinalComposite
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
	default:
		return "final"
	}
}
