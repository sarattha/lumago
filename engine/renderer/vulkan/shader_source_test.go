package vulkan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLightingShaderSourcesAreNotPlaceholders(t *testing.T) {
	root := filepath.Join("..", "..", "..", "shaders")
	tests := map[string][]string{
		"sprite_color.frag":  {"albedoTexture", "materialPass.emissive", "outEmissive"},
		"sprite_normal.frag": {"normalTexture", "hasNormalMap", "vec4(0.5, 0.5, 1.0, 1.0)"},
		"light_accum.frag":   {"sceneNormal", "PointLight", "uniforms.ambient", "dot(normal, lightDir)", "shadowMaps", "shadowFactor"},
		"shadow_map.frag":    {"ShadowMapPush", "lightPosition", "lightRadius", "outShadow"},
		"composite.frag":     {"sceneColor", "lightBuffer", "sceneEmissive", "debugView", "color.rgb * light.rgb + emissive.rgb"},
	}

	for name, snippets := range tests {
		source, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(source)
		for _, snippet := range snippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s missing %q", name, snippet)
			}
		}
		if strings.Contains(text, "outColor = vec4(1.0);") || strings.Contains(text, "outLight = vec4(1.0);") {
			t.Fatalf("%s still contains placeholder output", name)
		}
		if name == "composite.frag" && strings.Contains(text, "color.rgb * color.a") {
			t.Fatalf("%s uses sprite alpha as an emissive term", name)
		}
	}
}
