package assets

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
)

type Registry struct {
	nextTextureID graphics.TextureID
	textures      map[string]graphics.TextureID
	textureInfo   map[graphics.TextureID]graphics.TextureInfo
}

func NewRegistry() *Registry {
	return &Registry{
		nextTextureID: 1,
		textures:      make(map[string]graphics.TextureID),
		textureInfo:   make(map[graphics.TextureID]graphics.TextureInfo),
	}
}

func (r *Registry) LoadTexture(path string) graphics.TextureID {
	if id, ok := r.textures[path]; ok {
		return id
	}

	id := r.nextTextureID
	r.nextTextureID++
	r.textures[path] = id
	config, pixels, _ := imageData(path)
	r.textureInfo[id] = graphics.TextureInfo{
		ID:     id,
		Path:   path,
		Width:  config.Width,
		Height: config.Height,
	}
	graphics.RegisterTextureData(graphics.TextureData{
		ID:     id,
		Width:  config.Width,
		Height: config.Height,
		Pixels: pixels,
	})
	return id
}

func (r *Registry) Texture(id graphics.TextureID) (graphics.TextureInfo, bool) {
	info, ok := r.textureInfo[id]
	return info, ok
}

func (r *Registry) TextureByPath(path string) (graphics.TextureInfo, bool) {
	id, ok := r.textures[path]
	if !ok {
		return graphics.TextureInfo{}, false
	}
	return r.Texture(id)
}

func (r *Registry) LoadTextureInfo(path string) graphics.TextureInfo {
	id := r.LoadTexture(path)
	info, _ := r.Texture(id)
	return info
}

func imageData(path string) (image.Config, []lmath.Color, bool) {
	file, err := os.Open(path)
	if err != nil {
		return image.Config{}, nil, false
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return image.Config{}, nil, false
	}
	bounds := img.Bounds()
	config := image.Config{Width: bounds.Dx(), Height: bounds.Dy()}
	pixels := make([]lmath.Color, 0, config.Width*config.Height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels = append(pixels, lmath.Color{
				R: float32(r) / 65535,
				G: float32(g) / 65535,
				B: float32(b) / 65535,
				A: float32(a) / 65535,
			})
		}
	}
	return config, pixels, true
}
