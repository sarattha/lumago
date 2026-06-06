package vulkan

import (
	"testing"

	vk "github.com/sarattha/lumago/engine/renderer/vulkan/internal/vk"
)

func TestNeutralQuadTextureDoesNotTintSpriteColors(t *testing.T) {
	pixels := neutralQuadTexturePixels()
	if len(pixels) != 4 {
		t.Fatalf("neutral texture bytes=%d, want one RGBA pixel", len(pixels))
	}
	for i, value := range pixels {
		if value != 0xff {
			t.Fatalf("neutral texture byte %d=0x%x, want 0xff", i, value)
		}
	}
}

func TestClampToEdgeSamplerAddressModeIsAvailable(t *testing.T) {
	if vk.SamplerAddressModeClampToEdge != 2 {
		t.Fatalf("clamp-to-edge enum=%d, want Vulkan VK_SAMPLER_ADDRESS_MODE_CLAMP_TO_EDGE", vk.SamplerAddressModeClampToEdge)
	}
}
