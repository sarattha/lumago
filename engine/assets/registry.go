package assets

import "github.com/sarattha/lumago/engine/graphics"

type Registry struct {
	nextTextureID graphics.TextureID
	textures      map[string]graphics.TextureID
}

func NewRegistry() *Registry {
	return &Registry{
		nextTextureID: 1,
		textures:      make(map[string]graphics.TextureID),
	}
}

func (r *Registry) LoadTexture(path string) graphics.TextureID {
	if id, ok := r.textures[path]; ok {
		return id
	}

	id := r.nextTextureID
	r.nextTextureID++
	r.textures[path] = id
	return id
}
