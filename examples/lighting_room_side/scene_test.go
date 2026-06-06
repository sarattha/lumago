package main

import (
	"testing"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/graphics"
)

func TestSideLightingRoomBuildsSideViewShowcase(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	world := buildSideLightingRoom(game, config)

	if len(world.Sprites()) < 500 {
		t.Fatalf("sprites=%d, want dense side-view showcase", len(world.Sprites()))
	}
	if countSideMaterials(world.Sprites()) != sideMaterialCount {
		t.Fatalf("materials=%d, want %d", countSideMaterials(world.Sprites()), sideMaterialCount)
	}
	if len(world.Lights()) != sideLightCount {
		t.Fatalf("lights=%d, want %d", len(world.Lights()), sideLightCount)
	}
	if countSideShadowLights(world.Lights()) != sideShadowLightCount {
		t.Fatalf("shadow lights=%d, want %d", countSideShadowLights(world.Lights()), sideShadowLightCount)
	}
	if len(world.Occluders()) != sideOccluderCount {
		t.Fatalf("occluders=%d, want %d", len(world.Occluders()), sideOccluderCount)
	}
}

func TestSideLightingRoomLoadsDownloadedAssets(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	buildSideLightingRoom(game, config)

	for _, asset := range []struct {
		Name string
		Size int
	}{
		{Name: "background.png", Size: 24},
		{Name: "background_n.png", Size: 24},
		{Name: "character.png", Size: 24},
		{Name: "character_n.png", Size: 24},
		{Name: "wall.png", Size: 18},
		{Name: "wall_n.png", Size: 18},
		{Name: "floor.png", Size: 18},
		{Name: "floor_n.png", Size: 18},
		{Name: "platform.png", Size: 18},
		{Name: "platform_n.png", Size: 18},
		{Name: "crate.png", Size: 18},
		{Name: "crate_n.png", Size: 18},
		{Name: "lamp.png", Size: 18},
		{Name: "lamp_n.png", Size: 18},
		{Name: "overlay.png", Size: 18},
		{Name: "overlay_n.png", Size: 18},
	} {
		path := sideAssetPath(asset.Name)
		info, ok := game.Assets.TextureByPath(path)
		if !ok {
			t.Fatalf("%s was not loaded", path)
		}
		if info.Width != asset.Size || info.Height != asset.Size {
			t.Fatalf("%s size=%dx%d, want %dx%d", path, info.Width, info.Height, asset.Size, asset.Size)
		}
	}
}

func TestSideLightingRoomRolesReadAsSideView(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	world := buildSideLightingRoom(game, config)
	counts := countSideRoles(world.Sprites())

	if counts[sideRoleBackground] < 180 {
		t.Fatalf("background sprites=%d, want visible side-view back wall", counts[sideRoleBackground])
	}
	if counts[sideRoleFloor] < 100 {
		t.Fatalf("floor sprites=%d, want thick bottom floor", counts[sideRoleFloor])
	}
	if counts[sideRolePlatform] < 35 {
		t.Fatalf("platform sprites=%d, want side-view platform ledges", counts[sideRolePlatform])
	}
	if counts[sideRoleLamp] != sideLightCount {
		t.Fatalf("lamp sprites=%d, want one marker per light", counts[sideRoleLamp])
	}
	if counts[sideRoleCharacter] != 2 {
		t.Fatalf("characters=%d, want scale references in side view", counts[sideRoleCharacter])
	}
}

func TestSideLightingRoomLightsAndOccludersStayInViewport(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	world := buildSideLightingRoom(game, config)

	for i, light := range world.Lights() {
		if light.Position.X < 100 || light.Position.X > sideTargetWidth-100 || light.Position.Y < 70 || light.Position.Y > sideTargetHeight-70 {
			t.Fatalf("light %d position=%+v outside side room viewport", i, light.Position)
		}
	}
	for i, occluder := range world.Occluders() {
		for _, point := range occluder.Points {
			if point.X < 120 || point.X > sideTargetWidth-120 || point.Y < 60 || point.Y > sideTargetHeight-60 {
				t.Fatalf("occluder %d point=%+v outside viewport", i, point)
			}
		}
		for _, segment := range occluder.Segments {
			if segment.A.X < 120 || segment.B.X > sideTargetWidth-120 || segment.A.Y < 60 || segment.B.Y > sideTargetHeight-60 {
				t.Fatalf("occluder %d segment=%+v outside viewport", i, segment)
			}
		}
	}
}

func countSideMaterials(sprites []graphics.SpriteDrawCommand) int {
	seen := map[graphics.Material2D]bool{}
	for _, sprite := range sprites {
		seen[sprite.Sprite.Material] = true
	}
	return len(seen)
}

func countSideRoles(sprites []graphics.SpriteDrawCommand) map[sideSpriteRole]int {
	counts := map[sideSpriteRole]int{}
	for _, sprite := range sprites {
		counts[sideRoleForLayer(sprite.Layer)]++
	}
	return counts
}

func sideRoleForLayer(layer int) sideSpriteRole {
	switch layer {
	case 0:
		return sideRoleBackground
	case 2:
		return sideRoleWall
	case 3:
		return sideRoleFloor
	case 4:
		return sideRolePlatform
	case 5:
		return sideRoleProp
	case 7:
		return sideRoleLamp
	case 9:
		return sideRoleCharacter
	default:
		return sideRoleOverlay
	}
}

func countSideShadowLights(lights []graphics.Light2D) int {
	count := 0
	for _, light := range lights {
		if light.CastShadows {
			count++
		}
	}
	return count
}
