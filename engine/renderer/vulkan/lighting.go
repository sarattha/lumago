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
	litSpriteGrid     = 8
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

func litSpriteBatchForLighting(dst graphics.SpriteBatch, vertices []graphics.SpriteVertex, indices []uint16, batch graphics.SpriteBatch, lights []graphics.Light2D, shadows []lightShadowMap, config graphics.LightingConfig2D, extent vk.Extent2D) (graphics.SpriteBatch, []graphics.SpriteVertex, []uint16) {
	if len(batch.Vertices) == 0 {
		dst.Reset()
		return dst, vertices[:0], indices[:0]
	}

	maxSprites := int(^uint16(0)) / litSpriteVertexCount()
	commands := batch.Commands
	if len(commands) > maxSprites {
		commands = commands[:maxSprites]
	}
	vertexCount := len(commands) * litSpriteVertexCount()
	indexCount := len(commands) * litSpriteIndexCount()
	vertices = ensureLitVertices(vertices, vertexCount)
	indices = ensureLitIndices(indices, indexCount)

	dst.Commands = append(dst.Commands[:0], commands...)
	dst.Vertices = vertices
	dst.Indices = indices
	config = config.WithDefaults()
	for spriteIndex, command := range commands {
		sourceStart := spriteIndex * 4
		if sourceStart+4 > len(batch.Vertices) {
			break
		}
		writeLitSprite(
			vertices[spriteIndex*litSpriteVertexCount():(spriteIndex+1)*litSpriteVertexCount()],
			indices[spriteIndex*litSpriteIndexCount():(spriteIndex+1)*litSpriteIndexCount()],
			uint16(spriteIndex*litSpriteVertexCount()),
			batch.Vertices[sourceStart:sourceStart+4],
			command.Sprite.Material,
			lights,
			shadows,
			config,
			extent,
		)
	}
	dst.Stats = graphics.SpriteBatchStats{
		SpriteCount: len(commands),
		DrawCalls:   drawCallsForIndexCount(len(indices)),
		VertexCount: len(vertices),
		IndexCount:  len(indices),
	}
	return dst, vertices, indices
}

func writeLitSprite(dst []graphics.SpriteVertex, indices []uint16, base uint16, corners []graphics.SpriteVertex, material graphics.Material2D, lights []graphics.Light2D, shadows []lightShadowMap, config graphics.LightingConfig2D, extent vk.Extent2D) {
	vertexIndex := 0
	for y := 0; y <= litSpriteGrid; y++ {
		ty := float32(y) / litSpriteGrid
		for x := 0; x <= litSpriteGrid; x++ {
			tx := float32(x) / litSpriteGrid
			vertex := interpolateSpriteVertex(corners, tx, ty)
			vertex.Color = litVertexColor(vertex, material, lights, shadows, config, extent)
			dst[vertexIndex] = vertex
			vertexIndex++
		}
	}

	index := 0
	stride := litSpriteGrid + 1
	for y := 0; y < litSpriteGrid; y++ {
		for x := 0; x < litSpriteGrid; x++ {
			topLeft := base + uint16(y*stride+x)
			topRight := topLeft + 1
			bottomLeft := topLeft + uint16(stride)
			bottomRight := bottomLeft + 1
			indices[index+0] = topLeft
			indices[index+1] = topRight
			indices[index+2] = bottomRight
			indices[index+3] = bottomRight
			indices[index+4] = bottomLeft
			indices[index+5] = topLeft
			index += 6
		}
	}
}

func interpolateSpriteVertex(corners []graphics.SpriteVertex, tx, ty float32) graphics.SpriteVertex {
	top := interpolateVertex(corners[3], corners[2], tx)
	bottom := interpolateVertex(corners[0], corners[1], tx)
	return interpolateVertex(bottom, top, ty)
}

func interpolateVertex(a, b graphics.SpriteVertex, t float32) graphics.SpriteVertex {
	return graphics.SpriteVertex{
		Position: lerpVec2(a.Position, b.Position, t),
		UV:       lerpVec2(a.UV, b.UV, t),
		Color:    lerpColor(a.Color, b.Color, t),
	}
}

func litVertexColor(vertex graphics.SpriteVertex, material graphics.Material2D, lights []graphics.Light2D, shadows []lightShadowMap, config graphics.LightingConfig2D, extent vk.Extent2D) lmath.Color {
	base := vertex.Color
	normal := materialNormal(material, vertex.UV)
	pixel := clipToFramebuffer(vertex.Position, extent)
	light := accumulatedLight(pixel, normal, lights, shadows, config.Ambient)
	emissive := max0(material.Emissive)

	switch config.DebugView {
	case graphics.DebugViewSceneColor:
		return base
	case graphics.DebugViewSceneNormal:
		return lmath.Color{R: normal.X*0.5 + 0.5, G: normal.Y*0.5 + 0.5, B: normalZ(normal)*0.5 + 0.5, A: base.A}
	case graphics.DebugViewLightBuffer:
		return lmath.Color{R: light.R, G: light.G, B: light.B, A: base.A}
	case graphics.DebugViewShadowFactor:
		shadow := combinedShadowFactor(pixel, lights, shadows)
		return lmath.Color{R: shadow, G: shadow, B: shadow, A: base.A}
	default:
		return lmath.Color{
			R: base.R*light.R + base.R*emissive,
			G: base.G*light.G + base.G*emissive,
			B: base.B*light.B + base.B*emissive,
			A: base.A,
		}
	}
}

func clipToFramebuffer(position lmath.Vec2, extent vk.Extent2D) lmath.Vec2 {
	return lmath.Vec2{
		X: (position.X + 1) * 0.5 * float32(extent.Width),
		Y: (1 - position.Y) * 0.5 * float32(extent.Height),
	}
}

func materialNormal(material graphics.Material2D, uv lmath.Vec2) lmath.Vec2 {
	data, ok := graphics.RegisteredTextureData(material.Normal)
	if !ok {
		return lmath.Vec2{}
	}
	color := sampleTextureData(data, uv)
	return lmath.Vec2{X: color.R*2 - 1, Y: color.G*2 - 1}
}

func normalZ(normal lmath.Vec2) float32 {
	xy := normal.X*normal.X + normal.Y*normal.Y
	if xy >= 1 {
		return 0
	}
	return float32(math.Sqrt(float64(1 - xy)))
}

func accumulatedLight(pixel lmath.Vec2, normal lmath.Vec2, lights []graphics.Light2D, shadows []lightShadowMap, ambient lmath.Color) lmath.Color {
	result := ambient
	for i, light := range lights {
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
		intensity := max0(light.Intensity) * attenuation * ndotl * shadowFactorForLight(pixel, i, lights, shadows)
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

func sampleTextureData(data graphics.TextureData, uv lmath.Vec2) lmath.Color {
	if data.Width <= 0 || data.Height <= 0 || len(data.Pixels) == 0 {
		return lmath.Color{R: 0.5, G: 0.5, B: 1, A: 1}
	}
	u := uv.X - float32(math.Floor(float64(uv.X)))
	v := uv.Y - float32(math.Floor(float64(uv.Y)))
	x := int(u * float32(data.Width))
	y := int(v * float32(data.Height))
	if x >= data.Width {
		x = data.Width - 1
	}
	if y >= data.Height {
		y = data.Height - 1
	}
	index := y*data.Width + x
	if index < 0 || index >= len(data.Pixels) {
		return lmath.Color{R: 0.5, G: 0.5, B: 1, A: 1}
	}
	return data.Pixels[index]
}

func ensureLitVertices(vertices []graphics.SpriteVertex, count int) []graphics.SpriteVertex {
	if cap(vertices) < count {
		return make([]graphics.SpriteVertex, count)
	}
	return vertices[:count]
}

func ensureLitIndices(indices []uint16, count int) []uint16 {
	if cap(indices) < count {
		return make([]uint16, count)
	}
	return indices[:count]
}

func litSpriteVertexCount() int {
	return (litSpriteGrid + 1) * (litSpriteGrid + 1)
}

func litSpriteIndexCount() int {
	return litSpriteGrid * litSpriteGrid * 6
}

func drawCallsForIndexCount(indexCount int) int {
	if indexCount == 0 {
		return 0
	}
	return 1
}

func lerpVec2(a, b lmath.Vec2, t float32) lmath.Vec2 {
	return lmath.Vec2{X: a.X + (b.X-a.X)*t, Y: a.Y + (b.Y-a.Y)*t}
}

func lerpColor(a, b lmath.Color, t float32) lmath.Color {
	return lmath.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}
