package vulkan

import (
	"encoding/binary"
	"math"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

const (
	packedLightStride = 48
)

type lightingTargetKind uint8

const (
	lightingTargetSceneColor lightingTargetKind = iota
	lightingTargetSceneNormal
	lightingTargetLightBuffer
)

type lightingPassKind uint8

const (
	lightingPassSpriteColor lightingPassKind = iota
	lightingPassSpriteNormal
	lightingPassLightAccumulation
	lightingPassComposite
)

type lightingTarget struct {
	Kind   lightingTargetKind
	Name   string
	Width  uint32
	Height uint32
	Format vk.Format
}

type lightingRenderTargets struct {
	SceneColor  lightingTarget
	SceneNormal lightingTarget
	LightBuffer lightingTarget
}

type lightingPass struct {
	Kind    lightingPassKind
	Name    string
	Inputs  []lightingTargetKind
	Outputs []lightingTargetKind
}

func defaultLightingRenderTargets(extent vk.Extent2D, colorFormat vk.Format) lightingRenderTargets {
	return lightingRenderTargets{
		SceneColor: lightingTarget{
			Kind:   lightingTargetSceneColor,
			Name:   "scene_color",
			Width:  extent.Width,
			Height: extent.Height,
			Format: colorFormat,
		},
		SceneNormal: lightingTarget{
			Kind:   lightingTargetSceneNormal,
			Name:   "scene_normal",
			Width:  extent.Width,
			Height: extent.Height,
			Format: vk.FormatR8g8b8a8Unorm,
		},
		LightBuffer: lightingTarget{
			Kind:   lightingTargetLightBuffer,
			Name:   "light_buffer",
			Width:  extent.Width,
			Height: extent.Height,
			Format: colorFormat,
		},
	}
}

func defaultLightingPasses(debug graphics.DebugView2D) []lightingPass {
	passes := []lightingPass{
		{
			Kind:    lightingPassSpriteColor,
			Name:    "sprite_color",
			Outputs: []lightingTargetKind{lightingTargetSceneColor},
		},
		{
			Kind:    lightingPassSpriteNormal,
			Name:    "sprite_normal",
			Outputs: []lightingTargetKind{lightingTargetSceneNormal},
		},
		{
			Kind:    lightingPassLightAccumulation,
			Name:    "light_accumulation",
			Inputs:  []lightingTargetKind{lightingTargetSceneNormal},
			Outputs: []lightingTargetKind{lightingTargetLightBuffer},
		},
		{
			Kind: lightingPassComposite,
			Name: "composite",
			Inputs: []lightingTargetKind{
				lightingTargetSceneColor,
				lightingTargetSceneNormal,
				lightingTargetLightBuffer,
			},
		},
	}
	switch debug {
	case graphics.DebugViewSceneColor:
		passes[len(passes)-1].Inputs = []lightingTargetKind{lightingTargetSceneColor}
	case graphics.DebugViewSceneNormal:
		passes[len(passes)-1].Inputs = []lightingTargetKind{lightingTargetSceneNormal}
	case graphics.DebugViewLightBuffer:
		passes[len(passes)-1].Inputs = []lightingTargetKind{lightingTargetLightBuffer}
	}
	return passes
}

func prepareLightsForFrame(dst []graphics.Light2D, lights []graphics.Light2D, camera graphics.Camera2D) []graphics.Light2D {
	view := camera.ViewMatrix()
	zoom := camera.Zoom
	if zoom == 0 {
		zoom = 1
	}
	if zoom < 0 {
		zoom = -zoom
	}
	for _, light := range lights {
		frameLight := light
		frameLight.Position = view.TransformPoint(light.Position)
		frameLight.Radius = light.Radius * zoom
		dst = append(dst, frameLight)
	}
	return dst
}

func packLights(data []byte, lights []graphics.Light2D) []byte {
	size := len(lights) * packedLightStride
	if cap(data) < size {
		data = make([]byte, size)
	} else {
		data = data[:size]
	}
	for i, light := range lights {
		offset := i * packedLightStride
		putFloat32(data[offset:], light.Position.X)
		putFloat32(data[offset+4:], light.Position.Y)
		putFloat32(data[offset+8:], max0(light.Radius))
		putFloat32(data[offset+12:], 0)

		color := light.Color
		if color.A == 0 {
			color.A = 1
		}
		putFloat32(data[offset+16:], color.R)
		putFloat32(data[offset+20:], color.G)
		putFloat32(data[offset+24:], color.B)
		putFloat32(data[offset+28:], max0(light.Intensity))

		putFloat32(data[offset+32:], max0(light.Falloff))
		putFloat32(data[offset+36:], shadowFlag(light.CastShadows))
		putFloat32(data[offset+40:], color.A)
		putFloat32(data[offset+44:], 0)
	}
	return data
}

func unpackLight(data []byte, index int) graphics.Light2D {
	offset := index * packedLightStride
	return graphics.Light2D{
		Position: lmath.Vec2{
			X: float32FromBits(data[offset:]),
			Y: float32FromBits(data[offset+4:]),
		},
		Radius: float32FromBits(data[offset+8:]),
		Color: lmath.Color{
			R: float32FromBits(data[offset+16:]),
			G: float32FromBits(data[offset+20:]),
			B: float32FromBits(data[offset+24:]),
			A: float32FromBits(data[offset+40:]),
		},
		Intensity:   float32FromBits(data[offset+28:]),
		Falloff:     float32FromBits(data[offset+32:]),
		CastShadows: float32FromBits(data[offset+36:]) == 1,
	}
}

func putFloat32(data []byte, value float32) {
	binary.LittleEndian.PutUint32(data, math.Float32bits(value))
}

func float32FromBits(data []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(data))
}

func max0(value float32) float32 {
	if value < 0 {
		return 0
	}
	return value
}

func shadowFlag(enabled bool) float32 {
	if enabled {
		return 1
	}
	return 0
}
