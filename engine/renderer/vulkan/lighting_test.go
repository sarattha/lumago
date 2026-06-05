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
