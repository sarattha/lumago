package main

import (
	"testing"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/graphics"
)

func TestLightingRoomMatchesMVPAcceptanceTarget(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	world := buildLightingRoom(game, config)

	if len(world.Sprites()) != acceptanceSprites {
		t.Fatalf("sprites=%d, want %d", len(world.Sprites()), acceptanceSprites)
	}
	if countMaterials(world.Sprites()) != acceptanceMaterials {
		t.Fatalf("materials=%d, want %d", countMaterials(world.Sprites()), acceptanceMaterials)
	}
	if len(world.Lights()) != acceptanceLights {
		t.Fatalf("lights=%d, want %d", len(world.Lights()), acceptanceLights)
	}
	if countShadowLights(world.Lights()) != acceptanceShadowLights {
		t.Fatalf("shadow lights=%d, want %d", countShadowLights(world.Lights()), acceptanceShadowLights)
	}
	if len(world.Occluders()) != acceptanceOccluders {
		t.Fatalf("occluders=%d, want %d", len(world.Occluders()), acceptanceOccluders)
	}
	if config.Width != acceptanceTargetWidth || config.Height != acceptanceTargetHeight {
		t.Fatalf("target resolution=%dx%d, want %dx%d", config.Width, config.Height, acceptanceTargetWidth, acceptanceTargetHeight)
	}
}

func TestLightingRoomCarriesConfiguredDebugViews(t *testing.T) {
	for _, debugView := range []graphics.DebugView2D{
		graphics.DebugViewSceneColor,
		graphics.DebugViewSceneNormal,
		graphics.DebugViewLightBuffer,
		graphics.DebugViewShadowFactor,
		graphics.DebugViewSDF,
	} {
		config := defaultDemoConfig()
		config.DebugView = debugView
		game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
		world := buildLightingRoom(game, config)
		if got := world.LightingConfig().DebugView; got != debugView {
			t.Fatalf("debug view=%s, want %s", got, debugView)
		}
	}
}

func countMaterials(sprites []graphics.SpriteDrawCommand) int {
	seen := map[graphics.Material2D]bool{}
	for _, sprite := range sprites {
		seen[sprite.Sprite.Material] = true
	}
	return len(seen)
}

func countShadowLights(lights []graphics.Light2D) int {
	count := 0
	for _, light := range lights {
		if light.CastShadows {
			count++
		}
	}
	return count
}
