package render

import "github.com/spacehole-rogue/spacehole_rogue/internal/world"

// RenderTileGrid writes a TileGrid into a CellBuffer at the given offset.
func RenderTileGrid(buf *CellBuffer, grid *world.TileGrid, offsetX, offsetY int) {
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			tile := grid.Get(x, y)
			glyph, fg, bg := tileVisuals(tile)
			buf.Set(offsetX+x, offsetY+y, glyph, fg, bg)
		}
	}
}

func tileVisuals(t world.Tile) (glyph byte, fg, bg uint8) {
	if t.Equipment != nil {
		return equipVisuals(t.Equipment.Kind, t.Equipment.On)
	}

	switch t.Kind {
	case world.TileWall:
		return '#', ColorLightGray, ColorDarkGray
	case world.TileFloor:
		return '.', ColorDarkGray, ColorBlack
	case world.TileDoor:
		return '+', ColorDarkGray, ColorLightGray // inverted hull colors — stands out
	default:
		return ' ', ColorBlack, ColorBlack
	}
}

func equipVisuals(eq world.EquipmentKind, on bool) (glyph byte, fg, bg uint8) {
	switch eq {
	// --- crew quarters ---
	case world.EquipBed:
		return 127, ColorBrown, ColorBlack // ⌂ sleeping berth (open box)
	case world.EquipLocker:
		return ':', ColorBrown, ColorBlack // locker
	// --- bridge: viewscreen ---
	case world.EquipViewscreen:
		return '-', ColorLightCyan, ColorBlack // viewscreen
	// --- consoles: all = with color variant ---
	case world.EquipNavConsole:
		return '=', ColorLightCyan, ColorBlack // nav station
	case world.EquipPilotConsole:
		return '=', ColorLightGreen, ColorBlack // pilot station
	case world.EquipScienceConsole:
		return '=', ColorLightMagenta, ColorBlack // science station
	case world.EquipCargoConsole:
		return '=', ColorBrown, ColorBlack // cargo console
	case world.EquipCargoTransporter:
		// Toggleable - show darker when OFF
		if on {
			return 'X', ColorLightCyan, ColorBlack
		}
		return 'X', ColorDarkGray, ColorBlack
	// --- replicators: all $ with color variant ---
	case world.EquipFoodStation:
		return '$', ColorLightGreen, ColorBlack // food replicator (green)
	case world.EquipDrinkStation:
		return '$', ColorLightBlue, ColorBlack // drink replicator (blue)
	case world.EquipMedical:
		return '$', ColorLightCyan, ColorBlack // medical (cyan = between green & blue)
	// --- incinerator ---
	case world.EquipIncinerator:
		return '*', ColorRed, ColorBlack // incinerator (destroy matter → energy)
	// --- hygiene ---
	case world.EquipToilet:
		return 'o', ColorLightGray, ColorBlack // toilet
	case world.EquipShower:
		return '~', ColorLightCyan, ColorBlack // shower
	// --- tanks: all ■ with color variant ---
	case world.EquipOrganicTank:
		return 254, ColorGreen, ColorBlack // ■ organic tank (green)
	case world.EquipWaterTank:
		return 254, ColorBlue, ColorBlack // ■ water tank (blue)
	case world.EquipPowerCell:
		return 254, ColorBrown, ColorBlack // ■ battery (gold)
	case world.EquipFuelTank:
		return 254, ColorLightMagenta, ColorBlack // ■ fuel tank (purple = jump juice)
	// --- processors: all ▒ with color variant ---
	case world.EquipMatterRecycler:
		// Toggleable - show darker when OFF
		if on {
			return 177, ColorLightMagenta, ColorBlack
		}
		return 177, ColorDarkGray, ColorBlack
	case world.EquipGenerator:
		// Toggleable - show darker when OFF
		if on {
			return 177, ColorBrown, ColorBlack
		}
		return 177, ColorDarkGray, ColorBlack
	// --- propulsion ---
	case world.EquipEngine:
		// Toggleable - show darker when OFF
		if on {
			return '%', ColorBrown, ColorBlack
		}
		return '%', ColorDarkGray, ColorBlack
	case world.EquipJumpDrive:
		return 'J', ColorLightMagenta, ColorBlack // jump drive (purple = FTL)
	// --- cargo ---
	case world.EquipCargoTile:
		return 176, ColorDarkGray, ColorBlack // ░ cargo pad
	// --- surface equipment ---
	case world.EquipTerminal:
		return 'T', ColorLightCyan, ColorBlack // terminal
	case world.EquipLootCrate:
		return 'L', ColorBrown, ColorBlack // loot crate
	case world.EquipObjective:
		return '*', ColorYellow, ColorBlack // objective marker
	// --- salvage supplies: all & with color variant ---
	case world.EquipFuelCell:
		return '&', ColorLightMagenta, ColorBlack // fuel cells (purple = jump juice)
	case world.EquipSpareParts:
		return '&', ColorLightGray, ColorBlack // spare parts (gray = metal)
	case world.EquipPowerPack:
		return '&', ColorLightCyan, ColorBlack // power pack (cyan = electric)
	default:
		return '?', ColorWhite, ColorBlack
	}
}

// RenderSurfaceGrid writes a surface TileGrid with terrain-specific visuals.
func RenderSurfaceGrid(buf *CellBuffer, grid *world.TileGrid, terrain world.TerrainType, offsetX, offsetY int) {
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			tile := grid.Get(x, y)
			glyph, fg, bg := surfaceTileVisuals(tile, terrain)
			buf.Set(offsetX+x, offsetY+y, glyph, fg, bg)
		}
	}
}

// RenderSurfaceGridClipped renders a surface grid clipped to a viewport.
// Only tiles within (vpX, vpY) to (vpX+vpW-1, vpY+vpH-1) are drawn.
// camX, camY is the world position of the top-left of the viewport.
func RenderSurfaceGridClipped(buf *CellBuffer, grid *world.TileGrid, terrain world.TerrainType,
	vpX, vpY, vpW, vpH int, camX, camY int) {
	// Call the fog-aware version with no fog (everything visible)
	RenderSurfaceGridWithFog(buf, grid, terrain, vpX, vpY, vpW, vpH, camX, camY, nil, nil)
}

// RenderSurfaceGridWithFog renders a surface grid with fog of war support.
// isVisible and isSeen are callbacks that check visibility at world coordinates.
// If both are nil, everything is rendered normally (no fog).
func RenderSurfaceGridWithFog(buf *CellBuffer, grid *world.TileGrid, terrain world.TerrainType,
	vpX, vpY, vpW, vpH int, camX, camY int,
	isVisible func(x, y int) bool, isSeen func(x, y int) bool) {

	for sy := 0; sy < vpH; sy++ {
		for sx := 0; sx < vpW; sx++ {
			// World coordinates
			wx := camX + sx
			wy := camY + sy
			// Screen coordinates
			screenX := vpX + sx
			screenY := vpY + sy

			// Check world bounds
			if wx < 0 || wx >= grid.Width || wy < 0 || wy >= grid.Height {
				// Outside map - fill with terrain background
				bg := terrainBG(terrain)
				buf.Set(screenX, screenY, ' ', bg, bg)
				continue
			}

			// Check visibility (if fog callbacks provided)
			if isVisible != nil && isSeen != nil {
				if !isSeen(wx, wy) {
					// Never seen - show terrain background only
					bg := terrainBG(terrain)
					buf.Set(screenX, screenY, ' ', bg, bg)
					continue
				}
				if !isVisible(wx, wy) {
					// Seen but not currently visible - show dimmed
					tile := grid.Get(wx, wy)
					glyph, _, bg := surfaceTileVisuals(tile, terrain)
					// Use dark gray for remembered tiles
					buf.Set(screenX, screenY, glyph, ColorDarkGray, bg)
					continue
				}
			}

			// Visible or no fog - render normally
			tile := grid.Get(wx, wy)
			glyph, fg, bg := surfaceTileVisuals(tile, terrain)
			buf.Set(screenX, screenY, glyph, fg, bg)
		}
	}
}

// surfaceTileVisuals returns glyph/colors for terrain-aware tiles.
func surfaceTileVisuals(t world.Tile, terrain world.TerrainType) (glyph byte, fg, bg uint8) {
	// Equipment takes priority - but use terrain background
	if t.Equipment != nil {
		glyph, fg, _ = equipVisuals(t.Equipment.Kind, t.Equipment.On)
		return glyph, fg, terrainBG(terrain)
	}

	// Terrain-specific tile appearance
	switch t.Kind {
	case world.TileGround:
		return terrainGround(terrain)
	case world.TileRock:
		return terrainRock(terrain)
	case world.TileHazard:
		return terrainHazard(terrain)
	case world.TileShuttlePad:
		return terrainShuttlePad(terrain)
	case world.TileWall:
		// Man-made structures: clearly artificial
		return terrainWall(terrain)
	case world.TileFloor:
		// Interior floor in structures
		return terrainFloor(terrain)
	case world.TileDoor:
		return terrainDoor(terrain)
	default:
		bg := terrainBG(terrain)
		return ' ', bg, bg
	}
}

// terrainBG returns the base background color for a terrain type.
func terrainBG(terrain world.TerrainType) uint8 {
	switch terrain {
	case world.TerrainRocky:
		return ColorBrown // earthy desert brown
	case world.TerrainIce:
		return ColorBlue // deep frozen blue
	case world.TerrainVolcanic:
		return ColorRed // volcanic red
	case world.TerrainInterior:
		return ColorBlack // station interior
	default:
		return ColorBlack
	}
}

func terrainGround(terrain world.TerrainType) (byte, uint8, uint8) {
	switch terrain {
	case world.TerrainRocky:
		// Sandy desert: yellow dots on brown
		return '.', ColorYellow, ColorBrown
	case world.TerrainIce:
		// Frozen tundra: cyan dots on deep blue
		return '.', ColorLightCyan, ColorBlue
	case world.TerrainVolcanic:
		// Cooled lava: brown/orange on dark red
		return '.', ColorBrown, ColorRed
	case world.TerrainInterior:
		// Station floor: gray on black
		return '.', ColorDarkGray, ColorBlack
	default:
		return '.', ColorDarkGray, ColorBlack
	}
}

func terrainRock(terrain world.TerrainType) (byte, uint8, uint8) {
	// Natural rock formations: use ▒ (177) to distinguish from man-made walls
	switch terrain {
	case world.TerrainRocky:
		// Granite boulders: gray on brown
		return 177, ColorLightGray, ColorBrown
	case world.TerrainIce:
		// Ice formations: white on cyan
		return 177, ColorWhite, ColorCyan
	case world.TerrainVolcanic:
		// Volcanic rock: dark on red
		return 177, ColorDarkGray, ColorRed
	case world.TerrainInterior:
		// Debris: gray on black
		return 177, ColorDarkGray, ColorBlack
	default:
		return 177, ColorLightGray, ColorBlack
	}
}

func terrainHazard(terrain world.TerrainType) (byte, uint8, uint8) {
	switch terrain {
	case world.TerrainRocky:
		// Unstable ground: bright warning on contrasting bg
		return '^', ColorYellow, ColorRed
	case world.TerrainIce:
		// Crevasse: dark crack in ice
		return '~', ColorBlack, ColorCyan
	case world.TerrainVolcanic:
		// Lava pool: glowing yellow on bright red
		return '~', ColorYellow, ColorLightRed
	case world.TerrainInterior:
		// Breach/danger: red warning
		return '!', ColorLightRed, ColorBlack
	default:
		return '!', ColorLightRed, ColorBlack
	}
}

func terrainShuttlePad(terrain world.TerrainType) (byte, uint8, uint8) {
	// Shuttle pad: white H on cyan - stands out on any terrain
	return 'H', ColorWhite, ColorCyan
}

func terrainWall(terrain world.TerrainType) (byte, uint8, uint8) {
	// Man-made structure walls: solid block, clearly artificial
	switch terrain {
	case world.TerrainRocky, world.TerrainVolcanic:
		// Metal walls on hostile terrain: gray block
		return 219, ColorLightGray, ColorDarkGray // █
	case world.TerrainIce:
		// Research station: white walls
		return 219, ColorWhite, ColorDarkGray
	case world.TerrainInterior:
		// Standard station wall
		return '#', ColorLightGray, ColorDarkGray
	default:
		return '#', ColorLightGray, ColorDarkGray
	}
}

func terrainFloor(terrain world.TerrainType) (byte, uint8, uint8) {
	// Interior floor in structures
	switch terrain {
	case world.TerrainRocky, world.TerrainVolcanic, world.TerrainIce:
		// Shelter interior: dark floor
		return '.', ColorDarkGray, ColorBlack
	case world.TerrainInterior:
		// Station floor
		return '.', ColorDarkGray, ColorBlack
	default:
		return '.', ColorDarkGray, ColorBlack
	}
}

func terrainDoor(terrain world.TerrainType) (byte, uint8, uint8) {
	// Doors in structures
	return '+', ColorBrown, ColorBlack
}
