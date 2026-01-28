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
	if t.Equipment != world.EquipNone {
		return equipVisuals(t.Equipment)
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

func equipVisuals(eq world.EquipmentKind) (glyph byte, fg, bg uint8) {
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
		return 'X', ColorLightCyan, ColorBlack // cargo transporter pad
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
		return 177, ColorLightMagenta, ColorBlack // ▒ recycler (magenta)
	case world.EquipGenerator:
		return 177, ColorBrown, ColorBlack // ▒ generator (gold)
	// --- propulsion ---
	case world.EquipEngine:
		return '%', ColorBrown, ColorBlack // engine (gold)
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

// surfaceTileVisuals returns glyph/colors for terrain-aware tiles.
func surfaceTileVisuals(t world.Tile, terrain world.TerrainType) (glyph byte, fg, bg uint8) {
	// Equipment takes priority (same as ship)
	if t.Equipment != world.EquipNone {
		return equipVisuals(t.Equipment)
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
		// Structures use standard wall visuals
		return '#', ColorLightGray, ColorDarkGray
	case world.TileFloor:
		// Structures use standard floor visuals
		return '.', ColorDarkGray, ColorBlack
	case world.TileDoor:
		return '+', ColorBrown, ColorBlack
	default:
		return ' ', ColorBlack, ColorBlack
	}
}

func terrainGround(terrain world.TerrainType) (byte, uint8, uint8) {
	switch terrain {
	case world.TerrainRocky:
		return '.', ColorBrown, ColorBlack
	case world.TerrainIce:
		return '.', ColorLightCyan, ColorBlack
	case world.TerrainVolcanic:
		return '.', ColorRed, ColorBlack
	case world.TerrainInterior:
		return '.', ColorDarkGray, ColorBlack
	default:
		return '.', ColorDarkGray, ColorBlack
	}
}

func terrainRock(terrain world.TerrainType) (byte, uint8, uint8) {
	switch terrain {
	case world.TerrainRocky:
		return '#', ColorLightGray, ColorDarkGray
	case world.TerrainIce:
		return '#', ColorWhite, ColorCyan
	case world.TerrainVolcanic:
		return '#', ColorBrown, ColorRed
	case world.TerrainInterior:
		return '#', ColorLightGray, ColorDarkGray
	default:
		return '#', ColorLightGray, ColorDarkGray
	}
}

func terrainHazard(terrain world.TerrainType) (byte, uint8, uint8) {
	switch terrain {
	case world.TerrainRocky:
		return '^', ColorRed, ColorBlack // unstable ground
	case world.TerrainIce:
		return '~', ColorBlue, ColorBlack // crevasse
	case world.TerrainVolcanic:
		return '~', ColorYellow, ColorRed // lava
	case world.TerrainInterior:
		return '!', ColorRed, ColorBlack // breach/danger
	default:
		return '!', ColorRed, ColorBlack
	}
}

func terrainShuttlePad(terrain world.TerrainType) (byte, uint8, uint8) {
	// Shuttle pad is always bright white on green to stand out
	return 'H', ColorWhite, ColorGreen
}
