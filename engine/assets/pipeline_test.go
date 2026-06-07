package assets

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestImportAssetMetadataBuildsDeterministicManifest(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "tiles.png"), 32, 16)
	writePNG(t, filepath.Join(dir, "tiles_n.png"), 32, 16)
	writeFile(t, filepath.Join(dir, "level01.json"), "{}")
	writeFile(t, filepath.Join(dir, "font.ttf"), "font")

	metadata := AssetMetadata{
		Version: 1,
		Textures: []TextureMetadata{{
			Name:          "tiles",
			Source:        "tiles.png",
			PixelsPerUnit: 16,
			TileSize:      AssetSize{W: 16, H: 16},
		}},
		Sprites: []SpriteMetadata{
			{Name: "grass", Texture: "tiles", Rect: AssetRect{X: 0, Y: 0, W: 16, H: 16}},
			{Name: "stone", Texture: "tiles", Rect: AssetRect{X: 16, Y: 0, W: 16, H: 16}, Pivot: &AssetVec2{X: 0.5, Y: 1}},
		},
		Atlases: []AtlasMetadata{{
			Name:      "world",
			Sprites:   []string{"stone", "grass"},
			Padding:   2,
			Extrusion: 1,
		}},
		Animations: []AnimationMetadata{{
			Name:   "sparkle",
			Frames: []AnimationFrameMetadata{{Sprite: "grass", Seconds: 0.1}},
			Loop:   true,
		}},
		Fonts:    []FontMetadata{{Name: "ui", Source: "font.ttf"}},
		Tilemaps: []TilemapMetadata{{Name: "level01", Source: "level01.json", Sprites: []string{"stone", "grass"}}},
	}

	first, err := ImportAssetMetadata(metadata, dir)
	if err != nil {
		t.Fatalf("ImportAssetMetadata returned error: %v", err)
	}
	second, err := ImportAssetMetadata(metadata, dir)
	if err != nil {
		t.Fatalf("second ImportAssetMetadata returned error: %v", err)
	}
	firstJSON := mustJSON(t, first)
	secondJSON := mustJSON(t, second)
	if firstJSON != secondJSON {
		t.Fatalf("manifest is not deterministic:\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}

	if len(first.Textures) != 1 {
		t.Fatalf("texture count=%d, want 1", len(first.Textures))
	}
	texture := first.Textures[0]
	if texture.Filter != FilterNearest || texture.Wrap != WrapClampToEdge {
		t.Fatalf("texture sampling=%s/%s, want nearest/clamp_to_edge", texture.Filter, texture.Wrap)
	}
	if texture.PixelsPerUnit != 16 || texture.TileSize != (AssetSize{W: 16, H: 16}) {
		t.Fatalf("texture profile ppu=%d tile=%+v, want 16 and 16x16", texture.PixelsPerUnit, texture.TileSize)
	}
	if texture.NormalMapID == "" || len(first.NormalMaps) != 1 || first.NormalMaps[0].Source != "tiles_n.png" {
		t.Fatalf("normal-map convention was not recorded: texture=%+v normals=%+v", texture, first.NormalMaps)
	}

	spriteByName := map[string]ManifestSprite{}
	for _, sprite := range first.Sprites {
		spriteByName[sprite.Name] = sprite
	}
	if spriteByName["grass"].Rect != (AssetRect{X: 0, Y: 0, W: 16, H: 16}) {
		t.Fatalf("grass rect=%+v", spriteByName["grass"].Rect)
	}
	if spriteByName["grass"].PixelsPerUnit != 16 || spriteByName["grass"].Pivot != (AssetVec2{X: 0.5, Y: 0.5}) {
		t.Fatalf("grass defaults=%+v", spriteByName["grass"])
	}
	if first.Atlases[0].Padding != 2 || first.Atlases[0].Extrusion != 1 {
		t.Fatalf("atlas bleed settings=%+v", first.Atlases[0])
	}
	if got := first.Atlases[0].Sprites; len(got) != 2 || got[0] != "grass" || got[1] != "stone" {
		t.Fatalf("atlas sprites were not sorted deterministically: %+v", got)
	}
	if first.Tilemaps[0].TileSize != (AssetSize{W: 16, H: 16}) {
		t.Fatalf("tilemap default tile size=%+v, want 16x16", first.Tilemaps[0].TileSize)
	}
}

func TestImportAssetMetadataPreservesExplicitZeroPivot(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "tiles.png"), 16, 16)

	manifest, err := ImportAssetMetadata(AssetMetadata{
		Textures: []TextureMetadata{{Name: "tiles", Source: "tiles.png"}},
		Sprites: []SpriteMetadata{
			{Name: "corner", Texture: "tiles", Rect: AssetRect{X: 0, Y: 0, W: 16, H: 16}, Pivot: &AssetVec2{X: 0, Y: 0}},
			{Name: "default", Texture: "tiles", Rect: AssetRect{X: 0, Y: 0, W: 16, H: 16}},
		},
	}, dir)
	if err != nil {
		t.Fatalf("ImportAssetMetadata returned error: %v", err)
	}

	sprites := map[string]ManifestSprite{}
	for _, sprite := range manifest.Sprites {
		sprites[sprite.Name] = sprite
	}
	if sprites["corner"].Pivot != (AssetVec2{X: 0, Y: 0}) {
		t.Fatalf("explicit zero pivot=%+v, want origin", sprites["corner"].Pivot)
	}
	if sprites["default"].Pivot != (AssetVec2{X: 0.5, Y: 0.5}) {
		t.Fatalf("default pivot=%+v, want center", sprites["default"].Pivot)
	}
}

func TestImportAssetMetadataValidationErrors(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "hero.png"), 16, 16)
	writePNG(t, filepath.Join(dir, "hero_n.png"), 8, 8)
	writeFile(t, filepath.Join(dir, "map.json"), "{}")

	metadata := AssetMetadata{
		Textures: []TextureMetadata{
			{Name: "hero", Source: "hero.png", Normal: NormalMapMetadata{Required: true}},
			{Name: "hero", Source: "hero.png"},
			{Name: "missing", Source: "missing.png"},
		},
		Sprites: []SpriteMetadata{
			{Name: "idle", Texture: "hero", Rect: AssetRect{X: 8, Y: 8, W: 16, H: 16}},
			{Name: "idle", Texture: "hero", Rect: AssetRect{X: 0, Y: 0, W: 16, H: 16}},
		},
		Atlases:  []AtlasMetadata{{Name: "main", Sprites: []string{"missing"}, Padding: -1}},
		Tilemaps: []TilemapMetadata{{Name: "map", Source: "map.json", TileSize: AssetSize{W: 0, H: -1}}},
	}

	_, err := ImportAssetMetadata(metadata, dir)
	if err == nil {
		t.Fatal("ImportAssetMetadata returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"duplicate texture name hero",
		"missing file missing.png",
		"size 8x8 does not match albedo 16x16",
		"rectangle {X:8 Y:8 W:16 H:16} is outside texture hero (16x16)",
		"duplicate sprite name idle",
		"unknown sprite missing",
		"padding: must be zero or positive",
		"tileSize: must be positive",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("validation error %q missing from %q", want, message)
		}
	}
}

func TestImportAssetMetadataNeutralNormalFallback(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "hero.png"), 16, 16)

	manifest, err := ImportAssetMetadata(AssetMetadata{
		Textures: []TextureMetadata{{
			Name:   "hero",
			Source: "hero.png",
			Normal: NormalMapMetadata{NeutralFallback: true},
		}},
	}, dir)
	if err != nil {
		t.Fatalf("ImportAssetMetadata returned error: %v", err)
	}
	if len(manifest.NormalMaps) != 1 || !manifest.NormalMaps[0].NeutralFallback {
		t.Fatalf("neutral fallback normal not recorded: %+v", manifest.NormalMaps)
	}
	if manifest.Textures[0].NormalMapID != manifest.NormalMaps[0].ID {
		t.Fatalf("texture normal id=%q, want %q", manifest.Textures[0].NormalMapID, manifest.NormalMaps[0].ID)
	}
}

func TestHotReloaderReloadChanged(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "tiles.png"), 16, 16)
	metadataPath := filepath.Join(dir, "assets.json")
	writeAssetMetadata(t, metadataPath, AssetMetadata{
		Textures: []TextureMetadata{{Name: "tiles", Source: "tiles.png"}},
	})

	reloader, manifest, err := NewHotReloader(metadataPath)
	if err != nil {
		t.Fatalf("NewHotReloader returned error: %v", err)
	}
	if len(manifest.Textures) != 1 {
		t.Fatalf("texture count=%d, want 1", len(manifest.Textures))
	}
	if _, changed, err := reloader.ReloadChanged(); err != nil || changed {
		t.Fatalf("ReloadChanged unchanged=(changed=%v err=%v), want false nil", changed, err)
	}

	time.Sleep(10 * time.Millisecond)
	writePNG(t, filepath.Join(dir, "tiles.png"), 32, 16)
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(filepath.Join(dir, "tiles.png"), future, future); err != nil {
		t.Fatal(err)
	}
	manifest, changed, err := reloader.ReloadChanged()
	if err != nil {
		t.Fatalf("ReloadChanged returned error: %v", err)
	}
	if !changed {
		t.Fatal("ReloadChanged changed=false, want true")
	}
	if manifest.Textures[0].Width != 32 {
		t.Fatalf("reloaded texture width=%d, want 32", manifest.Textures[0].Width)
	}
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 255, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeAssetMetadata(t *testing.T, path string, metadata AssetMetadata) {
	t.Helper()
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
