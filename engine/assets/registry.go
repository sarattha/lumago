package assets

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/sarattha/lumago/engine/graphics"
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
	config, _ := imageConfig(path)
	r.textureInfo[id] = graphics.TextureInfo{
		ID:     id,
		Path:   path,
		Width:  config.Width,
		Height: config.Height,
	}
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

func imageConfig(path string) (image.Config, bool) {
	file, err := os.Open(path)
	if err != nil {
		return image.Config{}, false
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return image.Config{}, false
	}
	return config, true
}
