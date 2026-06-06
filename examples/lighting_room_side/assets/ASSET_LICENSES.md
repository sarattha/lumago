# Side-View Lighting Room Assets

External sprites:

- Source: Kenney Pixel Platformer, https://kenney.nl/assets/pixel-platformer
- Author: Kenney, https://www.kenney.nl
- License: Creative Commons CC0 1.0 Universal, https://creativecommons.org/publicdomain/zero/1.0/
- Pack version: 1.2

Files derived from the pack:

- `background.png` from `Tiles/Backgrounds/tile_0016.png`
- `wall.png` from `Tiles/tile_0040.png`
- `floor.png` from `Tiles/tile_0000.png`
- `platform.png` from `Tiles/tile_0090.png`
- `crate.png` from `Tiles/tile_0047.png`
- `lamp.png` from `Tiles/tile_0152.png`
- `character.png` from `Tiles/Characters/tile_0000.png`
- `overlay.png` from `Tiles/tile_0000.png`

The tile/backdrop render PNGs are kept as role-colored tile textures so the
downloaded pixel art remains legible in LumaGo's current CPU-lit sprite path,
which samples material textures into subdivided sprite vertices. The crate,
lamp, and character sprites preserve their original transparent silhouettes.

Generated neutral normal maps:

- `background_n.png`
- `wall_n.png`
- `floor_n.png`
- `platform_n.png`
- `crate_n.png`
- `lamp_n.png`
- `character_n.png`
- `overlay_n.png`

The normal maps are local neutral maps generated for the lighting demo because
the Pixel Platformer pack ships sprite albedo artwork, not normal maps.
