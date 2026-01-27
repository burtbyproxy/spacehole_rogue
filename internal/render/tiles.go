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
	default:
		return '?', ColorWhite, ColorBlack
	}
}
