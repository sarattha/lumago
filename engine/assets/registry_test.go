package assets

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
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
