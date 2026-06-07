package assets

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
)

func TestLoadTextureInfoDecodesImageMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "texture.png")
	img := image.NewRGBA(image.Rect(0, 0, 3, 5))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

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

	registry := NewRegistry()
	info := registry.LoadTextureInfo(path)

	if info.Width != 3 || info.Height != 5 {
		t.Fatalf("texture size=%dx%d, want 3x5", info.Width, info.Height)
	}
	if again := registry.LoadTexture(path); again != info.ID {
		t.Fatalf("LoadTexture returned id %d, want cached id %d", again, info.ID)
	}
}

func TestRegisterGeneratedTextureCachesInfoAndPixels(t *testing.T) {
	registry := NewRegistry()
	path := "generated://sprite"
	pixels := []lmath.Color{
		{R: 1, A: 1},
		{G: 1, A: 1},
		{B: 1, A: 1},
		{R: 1, G: 1, A: 1},
	}

	id := registry.RegisterGeneratedTexture(path, 2, 2, pixels)
	if id == graphics.InvalidTexture {
		t.Fatal("generated texture id is invalid")
	}
	if again := registry.RegisterGeneratedTexture(path, 2, 2, pixels); again != id {
		t.Fatalf("cached generated texture id=%d, want %d", again, id)
	}
	info, ok := registry.TextureByPath(path)
	if !ok {
		t.Fatalf("%s was not registered", path)
	}
	if info.Width != 2 || info.Height != 2 {
		t.Fatalf("generated texture size=%dx%d, want 2x2", info.Width, info.Height)
	}
	data, ok := graphics.RegisteredTextureData(id)
	if !ok {
		t.Fatalf("generated texture data for id=%d was not registered", id)
	}
	if len(data.Pixels) != len(pixels) || data.Pixels[3].R != 1 || data.Pixels[3].G != 1 {
		t.Fatalf("generated pixels were not preserved: %+v", data.Pixels)
	}
}
