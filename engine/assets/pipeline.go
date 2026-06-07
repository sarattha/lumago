package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DefaultPixelsPerUnit = 16
	DefaultTileSize      = 16
)

type FilterMode string

const (
	FilterNearest FilterMode = "nearest"
	FilterLinear  FilterMode = "linear"
)

type WrapMode string

const (
	WrapClampToEdge WrapMode = "clamp_to_edge"
	WrapRepeat      WrapMode = "repeat"
)

type AssetSize struct {
	W int `json:"w"`
	H int `json:"h"`
}

type AssetRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type AssetVec2 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type NormalMapMetadata struct {
	Source          string `json:"source,omitempty"`
	Required        bool   `json:"required,omitempty"`
	NeutralFallback bool   `json:"neutralFallback,omitempty"`
}

type TextureMetadata struct {
	Name          string            `json:"name"`
	Source        string            `json:"source"`
	PixelsPerUnit int               `json:"pixelsPerUnit,omitempty"`
	TileSize      AssetSize         `json:"tileSize,omitempty"`
	Filter        FilterMode        `json:"filter,omitempty"`
	Wrap          WrapMode          `json:"wrap,omitempty"`
	Normal        NormalMapMetadata `json:"normal,omitempty"`
}

type SpriteMetadata struct {
	Name          string    `json:"name"`
	Texture       string    `json:"texture"`
	Rect          AssetRect `json:"rect"`
	Pivot         AssetVec2 `json:"pivot,omitempty"`
	PixelsPerUnit int       `json:"pixelsPerUnit,omitempty"`
}

type AtlasMetadata struct {
	Name      string   `json:"name"`
	Sprites   []string `json:"sprites"`
	Padding   int      `json:"padding,omitempty"`
	Extrusion int      `json:"extrusion,omitempty"`
}

type AnimationFrameMetadata struct {
	Sprite  string  `json:"sprite"`
	Seconds float32 `json:"seconds"`
}

type AnimationMetadata struct {
	Name   string                   `json:"name"`
	Frames []AnimationFrameMetadata `json:"frames"`
	Loop   bool                     `json:"loop,omitempty"`
}

type FontMetadata struct {
	Name          string `json:"name"`
	Source        string `json:"source"`
	PixelsPerUnit int    `json:"pixelsPerUnit,omitempty"`
}

type TilemapMetadata struct {
	Name     string    `json:"name"`
	Source   string    `json:"source"`
	TileSize AssetSize `json:"tileSize,omitempty"`
	Sprites  []string  `json:"sprites,omitempty"`
}

type AssetMetadata struct {
	Version    int                 `json:"version"`
	Textures   []TextureMetadata   `json:"textures,omitempty"`
	Sprites    []SpriteMetadata    `json:"sprites,omitempty"`
	Atlases    []AtlasMetadata     `json:"atlases,omitempty"`
	Animations []AnimationMetadata `json:"animations,omitempty"`
	Fonts      []FontMetadata      `json:"fonts,omitempty"`
	Tilemaps   []TilemapMetadata   `json:"tilemaps,omitempty"`
}

type ManifestTexture struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Source        string     `json:"source"`
	Width         int        `json:"width"`
	Height        int        `json:"height"`
	PixelsPerUnit int        `json:"pixelsPerUnit"`
	TileSize      AssetSize  `json:"tileSize"`
	Filter        FilterMode `json:"filter"`
	Wrap          WrapMode   `json:"wrap"`
	NormalMapID   string     `json:"normalMapId,omitempty"`
}

type ManifestNormalMap struct {
	ID              string `json:"id"`
	Texture         string `json:"texture"`
	Source          string `json:"source,omitempty"`
	Width           int    `json:"width,omitempty"`
	Height          int    `json:"height,omitempty"`
	NeutralFallback bool   `json:"neutralFallback,omitempty"`
}

type ManifestSprite struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	TextureID     string    `json:"textureId"`
	Rect          AssetRect `json:"rect"`
	Pivot         AssetVec2 `json:"pivot"`
	PixelsPerUnit int       `json:"pixelsPerUnit"`
}

type ManifestAtlas struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Sprites   []string `json:"sprites"`
	Padding   int      `json:"padding"`
	Extrusion int      `json:"extrusion"`
}

type ManifestAnimation struct {
	ID     string                   `json:"id"`
	Name   string                   `json:"name"`
	Frames []AnimationFrameMetadata `json:"frames"`
	Loop   bool                     `json:"loop,omitempty"`
}

type ManifestFont struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	PixelsPerUnit int    `json:"pixelsPerUnit"`
}

type ManifestTilemap struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Source   string    `json:"source"`
	TileSize AssetSize `json:"tileSize"`
	Sprites  []string  `json:"sprites,omitempty"`
}

type AssetManifest struct {
	GeneratedFrom string              `json:"generatedFrom,omitempty"`
	Textures      []ManifestTexture   `json:"textures,omitempty"`
	NormalMaps    []ManifestNormalMap `json:"normalMaps,omitempty"`
	Sprites       []ManifestSprite    `json:"sprites,omitempty"`
	Atlases       []ManifestAtlas     `json:"atlases,omitempty"`
	Animations    []ManifestAnimation `json:"animations,omitempty"`
	Fonts         []ManifestFont      `json:"fonts,omitempty"`
	Tilemaps      []ManifestTilemap   `json:"tilemaps,omitempty"`
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	parts := make([]string, len(e))
	for i := range e {
		parts[i] = e[i].Error()
	}
	return strings.Join(parts, "; ")
}

func (e ValidationErrors) Unwrap() []error {
	errs := make([]error, len(e))
	for i := range e {
		errs[i] = e[i]
	}
	return errs
}

func LoadAssetMetadata(path string) (AssetMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AssetMetadata{}, err
	}
	var metadata AssetMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return AssetMetadata{}, err
	}
	return metadata, nil
}

func ImportAssetMetadataFile(path string) (AssetManifest, error) {
	metadata, err := LoadAssetMetadata(path)
	if err != nil {
		return AssetManifest{}, err
	}
	baseDir := filepath.Dir(path)
	manifest, err := ImportAssetMetadata(metadata, baseDir)
	manifest.GeneratedFrom = filepath.ToSlash(filepath.Clean(path))
	return manifest, err
}

func ImportAssetMetadata(metadata AssetMetadata, baseDir string) (AssetManifest, error) {
	importer := assetImporter{
		metadata: metadata,
		baseDir:  baseDir,
		textures: make(map[string]ManifestTexture),
		sprites:  make(map[string]ManifestSprite),
	}
	return importer.importManifest()
}

func WriteAssetManifest(path string, manifest AssetManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

type assetImporter struct {
	metadata AssetMetadata
	baseDir  string
	errs     ValidationErrors
	textures map[string]ManifestTexture
	sprites  map[string]ManifestSprite
}

func (i *assetImporter) importManifest() (AssetManifest, error) {
	manifest := AssetManifest{}
	i.importTextures(&manifest)
	i.importSprites(&manifest)
	i.importAtlases(&manifest)
	i.importAnimations(&manifest)
	i.importFonts(&manifest)
	i.importTilemaps(&manifest)
	i.sortManifest(&manifest)
	if len(i.errs) > 0 {
		return manifest, i.errs
	}
	return manifest, nil
}

func (i *assetImporter) importTextures(manifest *AssetManifest) {
	seen := map[string]bool{}
	for index, texture := range i.metadata.Textures {
		field := fmt.Sprintf("textures[%d]", index)
		if texture.Name == "" {
			i.addErr(field+".name", "is required")
			continue
		}
		if seen[texture.Name] {
			i.addErr(field+".name", "duplicate texture name "+texture.Name)
			continue
		}
		seen[texture.Name] = true
		if texture.Source == "" {
			i.addErr(field+".source", "is required")
			continue
		}
		config, format, ok := i.imageConfig(field+".source", texture.Source)
		if !ok {
			continue
		}
		if !supportedImageFormat(format) {
			i.addErr(field+".source", "unsupported image format "+format)
			continue
		}
		tileSize := texture.TileSize
		if tileSize.W == 0 && tileSize.H == 0 {
			tileSize = AssetSize{W: DefaultTileSize, H: DefaultTileSize}
		}
		if tileSize.W <= 0 || tileSize.H <= 0 {
			i.addErr(field+".tileSize", "must be positive")
		}
		ppu := texture.PixelsPerUnit
		if ppu == 0 {
			ppu = DefaultPixelsPerUnit
		}
		if ppu <= 0 {
			i.addErr(field+".pixelsPerUnit", "must be positive")
		}
		filter := texture.Filter
		if filter == "" {
			filter = FilterNearest
		}
		if filter != FilterNearest && filter != FilterLinear {
			i.addErr(field+".filter", "must be nearest or linear")
		}
		wrap := texture.Wrap
		if wrap == "" {
			wrap = WrapClampToEdge
		}
		if wrap != WrapClampToEdge && wrap != WrapRepeat {
			i.addErr(field+".wrap", "must be clamp_to_edge or repeat")
		}

		normalID := i.importNormalMap(&manifest.NormalMaps, texture, config, field)
		manifestTexture := ManifestTexture{
			ID:            stableAssetID("texture", texture.Name),
			Name:          texture.Name,
			Source:        cleanSlash(texture.Source),
			Width:         config.Width,
			Height:        config.Height,
			PixelsPerUnit: ppu,
			TileSize:      tileSize,
			Filter:        filter,
			Wrap:          wrap,
			NormalMapID:   normalID,
		}
		manifest.Textures = append(manifest.Textures, manifestTexture)
		i.textures[texture.Name] = manifestTexture
	}
}

func (i *assetImporter) importNormalMap(normals *[]ManifestNormalMap, texture TextureMetadata, albedo image.Config, field string) string {
	source := texture.Normal.Source
	if source == "" {
		source = normalConventionSource(texture.Source)
		if _, err := os.Stat(i.resolve(source)); err != nil {
			if texture.Normal.Required {
				i.addErr(field+".normal.source", "missing required normal map "+source)
			}
			if texture.Normal.NeutralFallback {
				id := stableAssetID("normal", texture.Name+":neutral")
				*normals = append(*normals, ManifestNormalMap{ID: id, Texture: texture.Name, NeutralFallback: true})
				return id
			}
			return ""
		}
	}
	config, format, ok := i.imageConfig(field+".normal.source", source)
	if !ok {
		return ""
	}
	if !supportedImageFormat(format) {
		i.addErr(field+".normal.source", "unsupported image format "+format)
		return ""
	}
	if config.Width != albedo.Width || config.Height != albedo.Height {
		i.addErr(field+".normal.source", fmt.Sprintf("size %dx%d does not match albedo %dx%d", config.Width, config.Height, albedo.Width, albedo.Height))
		return ""
	}
	id := stableAssetID("normal", texture.Name)
	*normals = append(*normals, ManifestNormalMap{
		ID:      id,
		Texture: texture.Name,
		Source:  cleanSlash(source),
		Width:   config.Width,
		Height:  config.Height,
	})
	return id
}

func (i *assetImporter) importSprites(manifest *AssetManifest) {
	seen := map[string]bool{}
	for index, sprite := range i.metadata.Sprites {
		field := fmt.Sprintf("sprites[%d]", index)
		if sprite.Name == "" {
			i.addErr(field+".name", "is required")
			continue
		}
		if seen[sprite.Name] {
			i.addErr(field+".name", "duplicate sprite name "+sprite.Name)
			continue
		}
		seen[sprite.Name] = true
		texture, ok := i.textures[sprite.Texture]
		if !ok {
			i.addErr(field+".texture", "unknown texture "+sprite.Texture)
			continue
		}
		if sprite.Rect.W <= 0 || sprite.Rect.H <= 0 {
			i.addErr(field+".rect", "width and height must be positive")
			continue
		}
		if sprite.Rect.X < 0 || sprite.Rect.Y < 0 || sprite.Rect.X+sprite.Rect.W > texture.Width || sprite.Rect.Y+sprite.Rect.H > texture.Height {
			i.addErr(field+".rect", fmt.Sprintf("rectangle %+v is outside texture %s (%dx%d)", sprite.Rect, texture.Name, texture.Width, texture.Height))
			continue
		}
		ppu := sprite.PixelsPerUnit
		if ppu == 0 {
			ppu = texture.PixelsPerUnit
		}
		if ppu <= 0 {
			i.addErr(field+".pixelsPerUnit", "must be positive")
		}
		pivot := sprite.Pivot
		if pivot.X == 0 && pivot.Y == 0 {
			pivot = AssetVec2{X: 0.5, Y: 0.5}
		}
		if pivot.X < 0 || pivot.X > 1 || pivot.Y < 0 || pivot.Y > 1 {
			i.addErr(field+".pivot", "must be normalized between 0 and 1")
		}
		manifestSprite := ManifestSprite{
			ID:            stableAssetID("sprite", sprite.Name),
			Name:          sprite.Name,
			TextureID:     texture.ID,
			Rect:          sprite.Rect,
			Pivot:         pivot,
			PixelsPerUnit: ppu,
		}
		manifest.Sprites = append(manifest.Sprites, manifestSprite)
		i.sprites[sprite.Name] = manifestSprite
	}
}

func (i *assetImporter) importAtlases(manifest *AssetManifest) {
	seen := map[string]bool{}
	for index, atlas := range i.metadata.Atlases {
		field := fmt.Sprintf("atlases[%d]", index)
		if atlas.Name == "" {
			i.addErr(field+".name", "is required")
			continue
		}
		if seen[atlas.Name] {
			i.addErr(field+".name", "duplicate atlas name "+atlas.Name)
			continue
		}
		seen[atlas.Name] = true
		if atlas.Padding < 0 {
			i.addErr(field+".padding", "must be zero or positive")
		}
		if atlas.Extrusion < 0 {
			i.addErr(field+".extrusion", "must be zero or positive")
		}
		sprites := append([]string(nil), atlas.Sprites...)
		sort.Strings(sprites)
		for _, sprite := range sprites {
			if _, ok := i.sprites[sprite]; !ok {
				i.addErr(field+".sprites", "unknown sprite "+sprite)
			}
		}
		manifest.Atlases = append(manifest.Atlases, ManifestAtlas{
			ID:        stableAssetID("atlas", atlas.Name),
			Name:      atlas.Name,
			Sprites:   sprites,
			Padding:   atlas.Padding,
			Extrusion: atlas.Extrusion,
		})
	}
}

func (i *assetImporter) importAnimations(manifest *AssetManifest) {
	seen := map[string]bool{}
	for index, animation := range i.metadata.Animations {
		field := fmt.Sprintf("animations[%d]", index)
		if animation.Name == "" {
			i.addErr(field+".name", "is required")
			continue
		}
		if seen[animation.Name] {
			i.addErr(field+".name", "duplicate animation name "+animation.Name)
			continue
		}
		seen[animation.Name] = true
		for frameIndex, frame := range animation.Frames {
			frameField := fmt.Sprintf("%s.frames[%d]", field, frameIndex)
			if _, ok := i.sprites[frame.Sprite]; !ok {
				i.addErr(frameField+".sprite", "unknown sprite "+frame.Sprite)
			}
			if frame.Seconds <= 0 {
				i.addErr(frameField+".seconds", "must be positive")
			}
		}
		manifest.Animations = append(manifest.Animations, ManifestAnimation{
			ID:     stableAssetID("animation", animation.Name),
			Name:   animation.Name,
			Frames: append([]AnimationFrameMetadata(nil), animation.Frames...),
			Loop:   animation.Loop,
		})
	}
}

func (i *assetImporter) importFonts(manifest *AssetManifest) {
	seen := map[string]bool{}
	for index, font := range i.metadata.Fonts {
		field := fmt.Sprintf("fonts[%d]", index)
		if font.Name == "" {
			i.addErr(field+".name", "is required")
			continue
		}
		if seen[font.Name] {
			i.addErr(field+".name", "duplicate font name "+font.Name)
			continue
		}
		seen[font.Name] = true
		if font.Source == "" {
			i.addErr(field+".source", "is required")
			continue
		}
		if _, err := os.Stat(i.resolve(font.Source)); err != nil {
			i.addErr(field+".source", "missing file "+font.Source)
			continue
		}
		ppu := font.PixelsPerUnit
		if ppu == 0 {
			ppu = DefaultPixelsPerUnit
		}
		if ppu <= 0 {
			i.addErr(field+".pixelsPerUnit", "must be positive")
		}
		manifest.Fonts = append(manifest.Fonts, ManifestFont{
			ID:            stableAssetID("font", font.Name),
			Name:          font.Name,
			Source:        cleanSlash(font.Source),
			PixelsPerUnit: ppu,
		})
	}
}

func (i *assetImporter) importTilemaps(manifest *AssetManifest) {
	seen := map[string]bool{}
	for index, tilemap := range i.metadata.Tilemaps {
		field := fmt.Sprintf("tilemaps[%d]", index)
		if tilemap.Name == "" {
			i.addErr(field+".name", "is required")
			continue
		}
		if seen[tilemap.Name] {
			i.addErr(field+".name", "duplicate tilemap name "+tilemap.Name)
			continue
		}
		seen[tilemap.Name] = true
		if tilemap.Source == "" {
			i.addErr(field+".source", "is required")
			continue
		}
		if _, err := os.Stat(i.resolve(tilemap.Source)); err != nil {
			i.addErr(field+".source", "missing file "+tilemap.Source)
		}
		tileSize := tilemap.TileSize
		if tileSize.W == 0 && tileSize.H == 0 {
			tileSize = AssetSize{W: DefaultTileSize, H: DefaultTileSize}
		}
		if tileSize.W <= 0 || tileSize.H <= 0 {
			i.addErr(field+".tileSize", "must be positive")
		}
		for _, sprite := range tilemap.Sprites {
			if _, ok := i.sprites[sprite]; !ok {
				i.addErr(field+".sprites", "unknown sprite "+sprite)
			}
		}
		sprites := append([]string(nil), tilemap.Sprites...)
		sort.Strings(sprites)
		manifest.Tilemaps = append(manifest.Tilemaps, ManifestTilemap{
			ID:       stableAssetID("tilemap", tilemap.Name),
			Name:     tilemap.Name,
			Source:   cleanSlash(tilemap.Source),
			TileSize: tileSize,
			Sprites:  sprites,
		})
	}
}

func (i *assetImporter) sortManifest(manifest *AssetManifest) {
	sort.Slice(manifest.Textures, func(a, b int) bool { return manifest.Textures[a].ID < manifest.Textures[b].ID })
	sort.Slice(manifest.NormalMaps, func(a, b int) bool { return manifest.NormalMaps[a].ID < manifest.NormalMaps[b].ID })
	sort.Slice(manifest.Sprites, func(a, b int) bool { return manifest.Sprites[a].ID < manifest.Sprites[b].ID })
	sort.Slice(manifest.Atlases, func(a, b int) bool { return manifest.Atlases[a].ID < manifest.Atlases[b].ID })
	sort.Slice(manifest.Animations, func(a, b int) bool { return manifest.Animations[a].ID < manifest.Animations[b].ID })
	sort.Slice(manifest.Fonts, func(a, b int) bool { return manifest.Fonts[a].ID < manifest.Fonts[b].ID })
	sort.Slice(manifest.Tilemaps, func(a, b int) bool { return manifest.Tilemaps[a].ID < manifest.Tilemaps[b].ID })
}

func (i *assetImporter) imageConfig(field, source string) (image.Config, string, bool) {
	file, err := os.Open(i.resolve(source))
	if err != nil {
		i.addErr(field, "missing file "+source)
		return image.Config{}, "", false
	}
	defer file.Close()
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		i.addErr(field, "cannot decode image metadata: "+err.Error())
		return image.Config{}, "", false
	}
	return config, format, true
}

func (i *assetImporter) resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(i.baseDir, filepath.FromSlash(path))
}

func (i *assetImporter) addErr(field, message string) {
	i.errs = append(i.errs, ValidationError{Field: field, Message: message})
}

func stableAssetID(kind, key string) string {
	sum := sha256.Sum256([]byte(kind + "\x00" + cleanSlash(strings.TrimSpace(key))))
	return kind + ":" + hex.EncodeToString(sum[:8])
}

func cleanSlash(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func normalConventionSource(source string) string {
	ext := filepath.Ext(source)
	return strings.TrimSuffix(source, ext) + "_n" + ext
}

func supportedImageFormat(format string) bool {
	return format == "png" || format == "jpeg" || format == "gif"
}

type HotReloader struct {
	metadataPath string
	manifest     AssetManifest
	modTimes     map[string]time.Time
}

func NewHotReloader(metadataPath string) (*HotReloader, AssetManifest, error) {
	manifest, err := ImportAssetMetadataFile(metadataPath)
	if err != nil {
		return nil, manifest, err
	}
	reloader := &HotReloader{metadataPath: metadataPath}
	if err := reloader.remember(manifest); err != nil {
		return nil, manifest, err
	}
	return reloader, manifest, nil
}

func (r *HotReloader) Manifest() AssetManifest {
	return r.manifest
}

func (r *HotReloader) ReloadChanged() (AssetManifest, bool, error) {
	changed, err := r.changed()
	if err != nil || !changed {
		return r.manifest, changed, err
	}
	manifest, err := ImportAssetMetadataFile(r.metadataPath)
	if err != nil {
		return r.manifest, true, err
	}
	if err := r.remember(manifest); err != nil {
		return r.manifest, true, err
	}
	return manifest, true, nil
}

func (r *HotReloader) remember(manifest AssetManifest) error {
	paths := append([]string{r.metadataPath}, watchedManifestSources(filepath.Dir(r.metadataPath), manifest)...)
	modTimes := make(map[string]time.Time, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		modTimes[path] = info.ModTime()
	}
	r.manifest = manifest
	r.modTimes = modTimes
	return nil
}

func (r *HotReloader) changed() (bool, error) {
	if r == nil {
		return false, errors.New("assets: nil hot reloader")
	}
	for path, last := range r.modTimes {
		info, err := os.Stat(path)
		if err != nil {
			return true, err
		}
		if !info.ModTime().Equal(last) {
			return true, nil
		}
	}
	return false, nil
}

func watchedManifestSources(baseDir string, manifest AssetManifest) []string {
	var paths []string
	add := func(source string) {
		if source == "" {
			return
		}
		if filepath.IsAbs(source) {
			paths = append(paths, source)
			return
		}
		paths = append(paths, filepath.Join(baseDir, filepath.FromSlash(source)))
	}
	for _, texture := range manifest.Textures {
		add(texture.Source)
	}
	for _, normal := range manifest.NormalMaps {
		add(normal.Source)
	}
	for _, font := range manifest.Fonts {
		add(font.Source)
	}
	for _, tilemap := range manifest.Tilemaps {
		add(tilemap.Source)
	}
	sort.Strings(paths)
	return paths
}
