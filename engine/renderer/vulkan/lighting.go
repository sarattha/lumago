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
	lightingTargetSceneEmissive
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
	SceneColor    lightingTarget
	SceneNormal   lightingTarget
	LightBuffer   lightingTarget
	SceneEmissive lightingTarget
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
		SceneEmissive: lightingTarget{
			Kind:   lightingTargetSceneEmissive,
			Name:   "scene_emissive",
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
			Outputs: []lightingTargetKind{lightingTargetSceneColor, lightingTargetSceneEmissive},
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
				lightingTargetSceneEmissive,
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

func shadeSpriteVerticesForLighting(dst []graphics.SpriteVertex, batch graphics.SpriteBatch, lights []graphics.Light2D, config graphics.LightingConfig2D, extent vk.Extent2D) []graphics.SpriteVertex {
	if len(batch.Vertices) == 0 {
		return dst[:0]
	}
	if cap(dst) < len(batch.Vertices) {
		dst = make([]graphics.SpriteVertex, len(batch.Vertices))
	} else {
		dst = dst[:len(batch.Vertices)]
	}
	copy(dst, batch.Vertices)

	config = config.WithDefaults()
	for spriteIndex, command := range batch.Commands {
		start := spriteIndex * 4
		if start+4 > len(dst) {
			break
		}
		for i := 0; i < 4; i++ {
			vertex := &dst[start+i]
			base := vertex.Color
			normal := materialNormal(command.Sprite.Material, vertex.UV)
			light := accumulatedLight(clipToFramebuffer(vertex.Position, extent), normal, lights, config.Ambient)
			emissive := max0(command.Sprite.Material.Emissive)

			switch config.DebugView {
			case graphics.DebugViewSceneColor:
				vertex.Color = base
			case graphics.DebugViewSceneNormal:
				vertex.Color = lmath.Color{R: normal.X*0.5 + 0.5, G: normal.Y*0.5 + 0.5, B: normalZ(normal)*0.5 + 0.5, A: base.A}
			case graphics.DebugViewLightBuffer:
				vertex.Color = lmath.Color{R: light.R, G: light.G, B: light.B, A: base.A}
			default:
				vertex.Color = lmath.Color{
					R: base.R*light.R + base.R*emissive,
					G: base.G*light.G + base.G*emissive,
					B: base.B*light.B + base.B*emissive,
					A: base.A,
				}
			}
		}
	}
	return dst
}

func clipToFramebuffer(position lmath.Vec2, extent vk.Extent2D) lmath.Vec2 {
	return lmath.Vec2{
		X: (position.X + 1) * 0.5 * float32(extent.Width),
		Y: (1 - position.Y) * 0.5 * float32(extent.Height),
	}
}

func materialNormal(material graphics.Material2D, uv lmath.Vec2) lmath.Vec2 {
	if material.Normal == graphics.InvalidTexture {
		return lmath.Vec2{}
	}
	x := float32(math.Sin(float64((uv.X + uv.Y) * 18.8495559215)))
	y := float32(math.Cos(float64((uv.X - uv.Y) * 18.8495559215)))
	return lmath.Vec2{X: x * 0.35, Y: y * 0.35}
}

func normalZ(normal lmath.Vec2) float32 {
	xy := normal.X*normal.X + normal.Y*normal.Y
	if xy >= 1 {
		return 0
	}
	return float32(math.Sqrt(float64(1 - xy)))
}

func accumulatedLight(pixel lmath.Vec2, normal lmath.Vec2, lights []graphics.Light2D, ambient lmath.Color) lmath.Color {
	result := ambient
	for _, light := range lights {
		radius := max0(light.Radius)
		if radius == 0 {
			continue
		}
		delta := light.Position.Sub(pixel)
		distance := float32(math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y)))
		attenuation := 1 - distance/radius
		if attenuation <= 0 {
			continue
		}
		falloff := light.Falloff
		if falloff <= 0 {
			falloff = 1
		}
		attenuation = float32(math.Pow(float64(attenuation), float64(falloff)))
		dir := lmath.Vec2{X: delta.X / radius, Y: delta.Y / radius}
		nz := normalZ(normal)
		dz := float32(1)
		len := float32(math.Sqrt(float64(dir.X*dir.X + dir.Y*dir.Y + dz*dz)))
		ndotl := (normal.X*dir.X + normal.Y*dir.Y + nz*dz) / len
		if ndotl < 0 {
			ndotl = 0
		}
		intensity := max0(light.Intensity) * attenuation * ndotl
		result.R += light.Color.R * intensity
		result.G += light.Color.G * intensity
		result.B += light.Color.B * intensity
	}
	return result
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
