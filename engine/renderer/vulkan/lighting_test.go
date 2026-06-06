package vulkan

import (
	"testing"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

func TestDefaultLightingRenderTargetsMatchSwapchainExtent(t *testing.T) {
	targets := defaultLightingRenderTargets(vk.Extent2D{Width: 1280, Height: 720}, vk.FormatB8g8r8a8Srgb)

	for _, target := range []lightingTarget{targets.SceneColor, targets.SceneNormal, targets.LightBuffer} {
		if target.Width != 1280 || target.Height != 720 {
			t.Fatalf("%s extent=%dx%d, want 1280x720", target.Name, target.Width, target.Height)
		}
	}
	if targets.SceneEmissive.Width != 1280 || targets.SceneEmissive.Height != 720 {
		t.Fatalf("scene emissive extent=%dx%d, want 1280x720", targets.SceneEmissive.Width, targets.SceneEmissive.Height)
	}
	if targets.SceneNormal.Format != vk.FormatR8g8b8a8Unorm {
		t.Fatalf("normal format=%d, want R8G8B8A8", targets.SceneNormal.Format)
	}
}

func TestDefaultLightingPasses(t *testing.T) {
	passes := defaultLightingPasses(graphics.DebugViewFinalComposite)
	want := []lightingPassKind{
		lightingPassSpriteColor,
		lightingPassSpriteNormal,
		lightingPassLightAccumulation,
		lightingPassComposite,
	}
	if len(passes) != len(want) {
		t.Fatalf("pass count=%d, want %d", len(passes), len(want))
	}
	if len(passes[0].Outputs) != 2 || passes[0].Outputs[1] != lightingTargetSceneEmissive {
		t.Fatalf("sprite color outputs=%v, want color and emissive", passes[0].Outputs)
	}
	for i, kind := range want {
		if passes[i].Kind != kind {
			t.Fatalf("pass %d kind=%d, want %d", i, passes[i].Kind, kind)
		}
	}

	debug := defaultLightingPasses(graphics.DebugViewSceneNormal)
	composite := debug[len(debug)-1]
	if len(composite.Inputs) != 1 || composite.Inputs[0] != lightingTargetSceneNormal {
		t.Fatalf("debug composite inputs=%v, want scene normal only", composite.Inputs)
	}
}

func TestPackLights(t *testing.T) {
	lights := []graphics.Light2D{
		{
			Position:    lmath.Vec2{X: 10, Y: 20},
			Radius:      200,
			Color:       lmath.Color{R: 1, G: 0.5, B: 0.25, A: 0.75},
			Intensity:   1.5,
			Falloff:     2,
			CastShadows: true,
		},
		{
			Position:  lmath.Vec2{X: -5, Y: 8},
			Radius:    -1,
			Intensity: -2,
			Falloff:   -3,
		},
	}

	data := packLights(nil, lights)
	if len(data) != len(lights)*packedLightStride {
		t.Fatalf("packed bytes=%d, want %d", len(data), len(lights)*packedLightStride)
	}

	first := unpackLight(data, 0)
	if first.Position != lights[0].Position || first.Radius != 200 || first.Intensity != 1.5 || !first.CastShadows {
		t.Fatalf("first light unpacked as %+v", first)
	}
	if first.Color.A != 0.75 {
		t.Fatalf("first alpha=%f, want 0.75", first.Color.A)
	}

	second := unpackLight(data, 1)
	if second.Radius != 0 || second.Intensity != 0 || second.Falloff != 0 {
		t.Fatalf("negative light fields were not clamped: %+v", second)
	}
	if second.Color.A != 1 {
		t.Fatalf("default alpha=%f, want 1", second.Color.A)
	}
}

func TestPrepareLightsForFrameTransformsWorldLightsToFramebufferSpace(t *testing.T) {
	lights := []graphics.Light2D{
		{
			Position:  lmath.Vec2{X: 12, Y: 23},
			Radius:    50,
			Intensity: 1,
		},
	}
	camera := graphics.Camera2D{
		Position: lmath.Vec2{X: 10, Y: 20},
		Zoom:     2,
	}

	got := prepareLightsForFrame(nil, lights, camera)
	if len(got) != 1 {
		t.Fatalf("lights=%d, want 1", len(got))
	}
	if got[0].Position.X != 4 || got[0].Position.Y != 6 {
		t.Fatalf("position=%+v, want framebuffer position (4, 6)", got[0].Position)
	}
	if got[0].Radius != 100 {
		t.Fatalf("radius=%f, want 100", got[0].Radius)
	}
	if lights[0].Position.X != 12 || lights[0].Radius != 50 {
		t.Fatalf("source light was mutated: %+v", lights[0])
	}
}

func TestLitSpriteBatchForLightingRespondsToInteriorLights(t *testing.T) {
	batch := singleSpriteBatch(graphics.Material2D{})
	config := graphics.LightingConfig2D{
		Ambient: lmath.Color{R: 0, G: 0, B: 0, A: 1},
	}

	dark, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})
	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, []graphics.Light2D{
		{
			Position:  lmath.Vec2{X: 50, Y: 50},
			Radius:    8,
			Color:     lmath.White(),
			Intensity: 1,
			Falloff:   1,
		},
	}, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	center := litSpriteVertexCount() / 2
	if !(dark.Vertices[center].Color.R < lit.Vertices[center].Color.R) {
		t.Fatalf("expected interior light to increase center red channel: dark=%+v lit=%+v", dark.Vertices[center].Color, lit.Vertices[center].Color)
	}
	if lit.Vertices[0].Color.R != 0 {
		t.Fatalf("corner was lit by an interior-only light: %+v", lit.Vertices[0].Color)
	}
	if batch.Vertices[0].Color != lmath.White() {
		t.Fatalf("source batch vertex was mutated: %+v", batch.Vertices[0].Color)
	}
}

func TestLitSpriteBatchForLightingSamplesRegisteredNormalMaps(t *testing.T) {
	flatNormal := graphics.TextureID(9001)
	tiltedNormal := graphics.TextureID(9002)
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     flatNormal,
		Width:  1,
		Height: 1,
		Pixels: []lmath.Color{{R: 0.5, G: 0.5, B: 1, A: 1}},
	})
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     tiltedNormal,
		Width:  1,
		Height: 1,
		Pixels: []lmath.Color{{R: 1, G: 0.5, B: 1, A: 1}},
	})
	flatBatch := singleSpriteBatch(graphics.Material2D{Normal: flatNormal})
	tiltedBatch := singleSpriteBatch(graphics.Material2D{Normal: tiltedNormal})
	config := graphics.LightingConfig2D{
		Ambient:   lmath.White(),
		DebugView: graphics.DebugViewSceneNormal,
	}

	flat, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, flatBatch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})
	tilted, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, tiltedBatch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	if flat.Vertices[0].Color == tilted.Vertices[0].Color {
		t.Fatalf("different normal maps produced same debug output: flat=%+v tilted=%+v", flat.Vertices[0].Color, tilted.Vertices[0].Color)
	}
	if tilted.Vertices[0].Color.R != 1 {
		t.Fatalf("tilted normal red=%f, want 1", tilted.Vertices[0].Color.R)
	}
}

func TestLitSpriteBatchForLightingSamplesRegisteredAlbedoTextures(t *testing.T) {
	albedo := graphics.TextureID(9010)
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     albedo,
		Width:  1,
		Height: 1,
		Pixels: []lmath.Color{{R: 0.25, G: 0.5, B: 0.75, A: 1}},
	})
	batch := normalizedTexturedSpriteBatch(graphics.Material2D{Albedo: albedo})
	config := graphics.LightingConfig2D{
		Ambient:   lmath.White(),
		DebugView: graphics.DebugViewSceneColor,
	}

	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	if got := lit.Vertices[0].Color; got.R != 0.25 || got.G != 0.5 || got.B != 0.75 || got.A != 1 {
		t.Fatalf("albedo vertex color=%+v, want registered texture color", got)
	}
}

func TestLitSpriteBatchForLightingKeepsTexturedSpritesTexelSharp(t *testing.T) {
	albedo := graphics.TextureID(9020)
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     albedo,
		Width:  2,
		Height: 2,
		Pixels: []lmath.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
			{R: 0, G: 0, B: 1, A: 1},
			{R: 1, G: 1, B: 1, A: 1},
		},
	})
	batch := normalizedTexturedSpriteBatch(graphics.Material2D{Albedo: albedo})
	config := graphics.LightingConfig2D{
		Ambient:   lmath.White(),
		DebugView: graphics.DebugViewSceneColor,
	}

	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	if lit.Stats.VertexCount != 16 || lit.Stats.IndexCount != 24 {
		t.Fatalf("texel-sharp geometry vertices=%d indices=%d, want 16/24 for four texels", lit.Stats.VertexCount, lit.Stats.IndexCount)
	}
	for cell := 0; cell < 4; cell++ {
		first := lit.Vertices[cell*4].Color
		for i := 1; i < 4; i++ {
			if got := lit.Vertices[cell*4+i].Color; got != first {
				t.Fatalf("cell %d vertex %d color=%+v, want constant texel color %+v", cell, i, got, first)
			}
		}
	}
	seen := map[lmath.Color]bool{}
	for cell := 0; cell < 4; cell++ {
		seen[lit.Vertices[cell*4].Color] = true
	}
	for _, color := range []lmath.Color{
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
	} {
		if !seen[color] {
			t.Fatalf("missing preserved texel color %+v from first vertices=%+v %+v %+v %+v", color, lit.Vertices[0].Color, lit.Vertices[4].Color, lit.Vertices[8].Color, lit.Vertices[12].Color)
		}
	}
	if len(seen) != 4 {
		t.Fatalf("texel colors were not preserved: first vertices=%+v %+v %+v %+v", lit.Vertices[0].Color, lit.Vertices[4].Color, lit.Vertices[8].Color, lit.Vertices[12].Color)
	}
}

func TestLitSpriteBatchForLightingSkipsTransparentTexels(t *testing.T) {
	albedo := graphics.TextureID(9021)
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     albedo,
		Width:  2,
		Height: 2,
		Pixels: []lmath.Color{
			{R: 1, A: 1},
			{G: 1, A: 0},
			{B: 1, A: 1},
			{R: 1, G: 1, B: 1, A: 1},
		},
	})
	batch := normalizedTexturedSpriteBatch(graphics.Material2D{Albedo: albedo})
	config := graphics.LightingConfig2D{
		Ambient:   lmath.White(),
		DebugView: graphics.DebugViewSceneColor,
	}

	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	if lit.Stats.VertexCount != 12 || lit.Stats.IndexCount != 18 {
		t.Fatalf("transparent texel geometry vertices=%d indices=%d, want 12/18 for three visible texels", lit.Stats.VertexCount, lit.Stats.IndexCount)
	}
	for _, vertex := range lit.Vertices {
		if vertex.Color.A <= transparentTexelCutoff {
			t.Fatalf("transparent texel was emitted: %+v", vertex.Color)
		}
	}
}

func TestLitSpriteBatchForLightingFallsBackForLargeTextures(t *testing.T) {
	albedo := graphics.TextureID(9022)
	pixels := make([]lmath.Color, maxLitSpriteTexelCells+1)
	for i := range pixels {
		pixels[i] = lmath.White()
	}
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     albedo,
		Width:  maxLitSpriteTexelCells + 1,
		Height: 1,
		Pixels: pixels,
	})
	batch := normalizedTexturedSpriteBatch(graphics.Material2D{Albedo: albedo})
	config := graphics.LightingConfig2D{
		Ambient:   lmath.White(),
		DebugView: graphics.DebugViewSceneColor,
	}

	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, nil, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	if lit.Stats.VertexCount != litSpriteVertexCount() || lit.Stats.IndexCount != litSpriteIndexCount() {
		t.Fatalf("large texture geometry vertices=%d indices=%d, want fixed grid %d/%d", lit.Stats.VertexCount, lit.Stats.IndexCount, litSpriteVertexCount(), litSpriteIndexCount())
	}
}

func TestLitSpriteBatchForLightingAppliesShadowMaps(t *testing.T) {
	batch := singleSpriteBatch(graphics.Material2D{})
	lights := []graphics.Light2D{
		{
			Position:    lmath.Vec2{X: 40, Y: 50},
			Radius:      80,
			Color:       lmath.White(),
			Intensity:   1,
			Falloff:     1,
			CastShadows: true,
		},
	}
	segments := []shadowSegment{{A: lmath.Vec2{X: 48, Y: 35}, B: lmath.Vec2{X: 48, Y: 65}}}
	shadows := buildLightShadowMaps(nil, lights, segments, 256)
	config := graphics.LightingConfig2D{Ambient: lmath.Color{A: 1}}

	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, lights, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})
	shadowed, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, lights, shadows, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})

	center := litSpriteVertexCount() / 2
	if !(shadowed.Vertices[center].Color.R < lit.Vertices[center].Color.R) {
		t.Fatalf("shadowed center was not darker: lit=%+v shadowed=%+v", lit.Vertices[center].Color, shadowed.Vertices[center].Color)
	}
}

func TestShadowFactorDebugViewOutputsGrayscaleShadow(t *testing.T) {
	batch := singleSpriteBatch(graphics.Material2D{})
	lights := []graphics.Light2D{
		{
			Position:    lmath.Vec2{X: 40, Y: 50},
			Radius:      80,
			Color:       lmath.White(),
			Intensity:   1,
			Falloff:     1,
			CastShadows: true,
		},
	}
	segments := []shadowSegment{{A: lmath.Vec2{X: 48, Y: 35}, B: lmath.Vec2{X: 48, Y: 65}}}
	shadows := buildLightShadowMaps(nil, lights, segments, 256)
	config := graphics.LightingConfig2D{Ambient: lmath.Color{A: 1}, DebugView: graphics.DebugViewShadowFactor}

	got, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, lights, shadows, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})
	center := got.Vertices[litSpriteVertexCount()/2].Color
	if center.R != 0 || center.G != 0 || center.B != 0 {
		t.Fatalf("center debug color=%+v, want shadow black", center)
	}
}

func TestLitSpriteBatchForLightingAppliesSDFMode(t *testing.T) {
	batch := singleSpriteBatch(graphics.Material2D{})
	lights := []graphics.Light2D{
		{
			Position:    lmath.Vec2{X: 40, Y: 50},
			Radius:      80,
			Color:       lmath.White(),
			Intensity:   1,
			Falloff:     1,
			CastShadows: true,
		},
	}
	occluder := graphics.SegmentOccluder2D(lmath.Vec2{X: 48, Y: 35}, lmath.Vec2{X: 48, Y: 65}, 1)
	sdf := buildStaticSDFTextureFromOccluders(sdfTexture{}, []graphics.Occluder2D{occluder}, graphics.DefaultCamera2D(), 100, 100, 2)
	config := graphics.LightingConfig2D{Ambient: lmath.Color{A: 1}, ShadowMode: graphics.ShadowModeSDFExperimental}

	lit, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, lights, nil, sdfTexture{}, config, vk.Extent2D{Width: 100, Height: 100})
	shadowed, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, lights, nil, sdf, config, vk.Extent2D{Width: 100, Height: 100})

	center := litSpriteVertexCount() / 2
	if !(shadowed.Vertices[center].Color.R < lit.Vertices[center].Color.R) {
		t.Fatalf("sdf shadowed center was not darker: lit=%+v shadowed=%+v", lit.Vertices[center].Color, shadowed.Vertices[center].Color)
	}
}

func TestSDFDebugViewOutputsDistanceVisualization(t *testing.T) {
	batch := singleSpriteBatch(graphics.Material2D{})
	occluder := graphics.RectOccluder2D(lmath.Rect{X: 45, Y: 45, W: 10, H: 10}, 1)
	sdf := buildStaticSDFTextureFromOccluders(sdfTexture{}, []graphics.Occluder2D{occluder}, graphics.DefaultCamera2D(), 100, 100, 2)
	config := graphics.LightingConfig2D{Ambient: lmath.Color{A: 1}, DebugView: graphics.DebugViewSDF, ShadowMode: graphics.ShadowModeSDFExperimental}

	got, _, _ := litSpriteBatchForLighting(graphics.SpriteBatch{}, nil, nil, batch, nil, nil, sdf, config, vk.Extent2D{Width: 100, Height: 100})
	center := got.Vertices[litSpriteVertexCount()/2].Color
	if !(center.B > center.R) {
		t.Fatalf("sdf debug center color=%+v, want inside occluder tint", center)
	}
}

func singleSpriteBatch(material graphics.Material2D) graphics.SpriteBatch {
	var batch graphics.SpriteBatch
	batch.Build([]graphics.SpriteDrawCommand{
		{
			Sprite: graphics.Sprite{
				Material: material,
				Src:      lmath.Rect{W: 20, H: 20},
				Color:    lmath.White(),
			},
			Transform: graphics.Transform2D{
				Position: lmath.Vec2{X: 50, Y: 50},
				Scale:    lmath.Vec2{X: 1, Y: 1},
			},
		},
	}, graphics.DefaultCamera2D(), 100, 100)
	return batch
}

func normalizedTexturedSpriteBatch(material graphics.Material2D) graphics.SpriteBatch {
	var batch graphics.SpriteBatch
	batch.Build([]graphics.SpriteDrawCommand{
		{
			Sprite: graphics.Sprite{
				Material: material,
				Src:      lmath.Rect{W: 1, H: 1},
				Color:    lmath.White(),
			},
			Transform: graphics.Transform2D{
				Position: lmath.Vec2{X: 50, Y: 50},
				Scale:    lmath.Vec2{X: 20, Y: 20},
			},
		},
	}, graphics.DefaultCamera2D(), 100, 100)
	return batch
}
