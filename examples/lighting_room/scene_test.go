package main

import (
	"math"
	"testing"

	"github.com/sarattha/lumago/engine/app"
	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
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

func TestLightingRoomLoadsExternalAssetTextures(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	buildLightingRoom(game, config)

	for _, name := range []string{
		"floor.png",
		"wall.png",
		"character.png",
		"prop.png",
		"floor_n.png",
		"wall_n.png",
		"character_n.png",
		"prop_n.png",
	} {
		path := lightingRoomAssetPath(name)
		info, ok := game.Assets.TextureByPath(path)
		if !ok {
			t.Fatalf("%s was not loaded", path)
		}
		if info.Width != 16 || info.Height != 16 {
			t.Fatalf("%s size=%dx%d, want 16x16 external asset texture", path, info.Width, info.Height)
		}
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

func TestLightingRoomSpriteGridIsCenteredInTargetViewport(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	world := buildLightingRoom(game, config)

	min, max := spriteBounds(world.Sprites())
	leftMargin := min.X - acceptanceTileSize/2
	rightMargin := float32(acceptanceTargetWidth) - (max.X + acceptanceTileSize/2)
	topMargin := min.Y - acceptanceTileSize/2
	bottomMargin := float32(acceptanceTargetHeight) - (max.Y + acceptanceTileSize/2)

	if math.Abs(float64(leftMargin-rightMargin)) > 4 {
		t.Fatalf("horizontal margins left=%.1f right=%.1f, want centered", leftMargin, rightMargin)
	}
	if math.Abs(float64(topMargin-bottomMargin)) > 40 {
		t.Fatalf("vertical margins top=%.1f bottom=%.1f, want room centered", topMargin, bottomMargin)
	}
}

func TestLightingRoomMaterialDistributionIsSpatialNotColumnStriped(t *testing.T) {
	for _, x := range []int{8, 16, 24, 32} {
		seen := map[int]bool{}
		for y := 2; y < acceptanceTileRows-2; y++ {
			seen[roomMaterialIndex(x, y)] = true
		}
		if len(seen) < 2 {
			t.Fatalf("column %d uses %d material(s), want spatially mixed room detail", x, len(seen))
		}
	}

	for _, y := range []int{7, 13, 18} {
		seen := map[int]bool{}
		for x := 2; x < acceptanceTileColumns-2; x++ {
			seen[roomMaterialIndex(x, y)] = true
		}
		if len(seen) < 3 {
			t.Fatalf("row %d uses %d material(s), want floor, wall, and prop variation", y, len(seen))
		}
	}
}

func TestLightingRoomVisualCommunicationRolesAreObvious(t *testing.T) {
	counts := map[spriteRole]int{}
	for y := 0; y < acceptanceTileRows; y++ {
		for x := 0; x < acceptanceTileColumns; x++ {
			counts[roomSpriteSpecAt(x, y).Role]++
		}
	}

	if counts[spriteRoleWall] < 180 {
		t.Fatalf("wall sprites=%d, want broad wall bands", counts[spriteRoleWall])
	}
	if counts[spriteRoleProp] < 95 {
		t.Fatalf("prop sprites=%d, want readable prop blocks", counts[spriteRoleProp])
	}
	if counts[spriteRoleLightMarker] != 36 {
		t.Fatalf("light marker sprites=%d, want four 3x3 marker glyphs", counts[spriteRoleLightMarker])
	}
	if counts[spriteRoleLegend] < 20 {
		t.Fatalf("legend sprites=%d, want visible debug legend/labels", counts[spriteRoleLegend])
	}

	floorA := roomSpriteSpecAt(6, 11)
	floorB := roomSpriteSpecAt(7, 11)
	if floorA.Role != spriteRoleFloor || floorB.Role != spriteRoleFloor {
		t.Fatalf("floor samples roles=%v,%v, want floor", floorA.Role, floorB.Role)
	}
	if sameColor(floorA.Color, floorB.Color) {
		t.Fatalf("floor samples have same color %+v, want visible tile contrast", floorA.Color)
	}

	leftWall := roomSpriteSpecAt(0, 12)
	nextWall := roomSpriteSpecAt(1, 12)
	if leftWall.Role != spriteRoleWall || nextWall.Role != spriteRoleWall {
		t.Fatalf("left band roles=%v,%v, want two-tile wall band", leftWall.Role, nextWall.Role)
	}
	if leftWall.Scale.X <= 1 || nextWall.Scale.X <= 1 {
		t.Fatalf("wall scales=%+v,%+v, want expanded band tiles", leftWall.Scale, nextWall.Scale)
	}

	propEdge := roomSpriteSpecAt(24, 7)
	propFace := roomSpriteSpecAt(25, 8)
	if propEdge.Role != spriteRoleProp || propFace.Role != spriteRoleProp {
		t.Fatalf("prop samples roles=%v,%v, want prop block", propEdge.Role, propFace.Role)
	}
	if propEdge.Color.R >= propFace.Color.R {
		t.Fatalf("prop edge=%+v face=%+v, want darker edge and brighter face", propEdge.Color, propFace.Color)
	}

	lightCenter := roomSpriteSpecAt(9, 9)
	lightArm := roomSpriteSpecAt(10, 9)
	if lightCenter.Role != spriteRoleLightMarker || lightArm.Role != spriteRoleLightMarker {
		t.Fatalf("light marker roles=%v,%v, want center and arm glyph", lightCenter.Role, lightArm.Role)
	}
	if lightCenter.Scale.X <= lightArm.Scale.X {
		t.Fatalf("light marker scales center=%+v arm=%+v, want emphasized center", lightCenter.Scale, lightArm.Scale)
	}

	for _, point := range []struct{ x, y int }{{22, 20}, {26, 20}, {30, 20}, {34, 20}} {
		if got := roomSpriteSpecAt(point.x, point.y).Role; got != spriteRoleLegend {
			t.Fatalf("legend swatch (%d,%d) role=%v, want legend", point.x, point.y, got)
		}
	}
}

func TestLightingRoomLightsAreInsideRoomAndOccludersAlignToRoom(t *testing.T) {
	config := defaultDemoConfig()
	game := app.NewGame(app.Config{Width: config.Width, Height: config.Height})
	world := buildLightingRoom(game, config)

	for i, light := range world.Lights() {
		if light.Position.X < 120 || light.Position.X > 1810 || light.Position.Y < 90 || light.Position.Y > 960 {
			t.Fatalf("light %d position=%+v, want inside room bounds", i, light.Position)
		}
	}

	for i, occluder := range world.Occluders() {
		for _, point := range occluder.Points {
			if point.X < 90 || point.X > 1840 || point.Y < 70 || point.Y > 1010 {
				t.Fatalf("occluder %d point=%+v outside visible room bounds", i, point)
			}
		}
		for _, segment := range occluder.Segments {
			if segment.A.X < 90 || segment.B.X > 1840 || segment.A.Y < 70 || segment.B.Y > 1010 {
				t.Fatalf("occluder %d segment=%+v outside visible room bounds", i, segment)
			}
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

func spriteBounds(sprites []graphics.SpriteDrawCommand) (lmath.Vec2, lmath.Vec2) {
	min := lmath.Vec2{X: float32(acceptanceTargetWidth), Y: float32(acceptanceTargetHeight)}
	max := lmath.Vec2{}
	for _, sprite := range sprites {
		position := sprite.Transform.Position
		if position.X < min.X {
			min.X = position.X
		}
		if position.Y < min.Y {
			min.Y = position.Y
		}
		if position.X > max.X {
			max.X = position.X
		}
		if position.Y > max.Y {
			max.Y = position.Y
		}
	}
	return min, max
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

func sameColor(a, b lmath.Color) bool {
	const epsilon = 0.0001
	return math.Abs(float64(a.R-b.R)) < epsilon &&
		math.Abs(float64(a.G-b.G)) < epsilon &&
		math.Abs(float64(a.B-b.B)) < epsilon &&
		math.Abs(float64(a.A-b.A)) < epsilon
}
