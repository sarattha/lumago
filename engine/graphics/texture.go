package graphics

import lmath "github.com/sarattha/lumago/engine/math"

type TextureID uint32

const InvalidTexture TextureID = 0

type TextureInfo struct {
	ID     TextureID
	Path   string
	Width  int
	Height int
}

func (t TextureInfo) Valid() bool {
	return t.ID != InvalidTexture
}

type AtlasFrame struct {
	Name string
	Page TextureID
	Src  lmath.Rect
}

type TextureAtlas struct {
	pages  map[TextureID]TextureInfo
	frames map[string]AtlasFrame
}

func NewTextureAtlas() *TextureAtlas {
	return &TextureAtlas{
		pages:  make(map[TextureID]TextureInfo),
		frames: make(map[string]AtlasFrame),
	}
}

func (a *TextureAtlas) AddPage(info TextureInfo) {
	if a.pages == nil {
		a.pages = make(map[TextureID]TextureInfo)
	}
	a.pages[info.ID] = info
}

func (a *TextureAtlas) AddFrame(name string, page TextureID, src lmath.Rect) {
	if a.frames == nil {
		a.frames = make(map[string]AtlasFrame)
	}
	a.frames[name] = AtlasFrame{Name: name, Page: page, Src: src}
}

func (a *TextureAtlas) Frame(name string) (AtlasFrame, bool) {
	if a == nil {
		return AtlasFrame{}, false
	}
	frame, ok := a.frames[name]
	return frame, ok
}

func (a *TextureAtlas) UVRect(frame AtlasFrame) (lmath.Rect, bool) {
	if a == nil {
		return lmath.Rect{}, false
	}
	page, ok := a.pages[frame.Page]
	if !ok || page.Width <= 0 || page.Height <= 0 {
		return lmath.Rect{}, false
	}
	return lmath.Rect{
		X: frame.Src.X / float32(page.Width),
		Y: frame.Src.Y / float32(page.Height),
		W: frame.Src.W / float32(page.Width),
		H: frame.Src.H / float32(page.Height),
	}, true
}
