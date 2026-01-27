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
		return '+', ColorBrown, ColorBlack
	default:
		return ' ', ColorBlack, ColorBlack
	}
}

func equipVisuals(eq world.EquipmentKind) (glyph byte, fg, bg uint8) {
	switch eq {
	case world.EquipConsole:
		return '=', ColorGreen, ColorBlack
	case world.EquipBed:
		return 'b', ColorBrown, ColorBlack
	case world.EquipPowerCell:
		return 'P', ColorYellow, ColorBlack
	case world.EquipEngine:
		return 'E', ColorLightRed, ColorBlack
	case world.EquipToilet:
		return 't', ColorLightGray, ColorBlack
	case world.EquipShower:
		return 's', ColorLightCyan, ColorBlack
	case world.EquipReplicator:
		return 'R', ColorLightGreen, ColorBlack
	case world.EquipWaterRecycler:
		return 'w', ColorLightBlue, ColorBlack
	case world.EquipFoodStore:
		return 'f', ColorGreen, ColorBlack
	case world.EquipWaterTank:
		return 'W', ColorBlue, ColorBlack
	default:
		return '?', ColorWhite, ColorBlack
	}
}
