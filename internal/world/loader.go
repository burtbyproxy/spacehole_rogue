package world

import (
	"encoding/json"
	"fmt"
)

// ShipLayout is the JSON-serializable definition of a ship layout.
type ShipLayout struct {
	Name   string       `json:"name"`
	Class  string       `json:"class"`
	Width  int          `json:"width"`
	Height int          `json:"height"`
	Tiles  []string     `json:"tiles"`
	Rooms  []RoomDef    `json:"rooms"`
	Spawn  [2]int       `json:"spawn"`
}

// RoomDef defines a named room in a ship layout.
type RoomDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LoadShipLayout parses a ShipLayout from JSON bytes.
func LoadShipLayout(data []byte) (*ShipLayout, error) {
	var layout ShipLayout
	if err := json.Unmarshal(data, &layout); err != nil {
		return nil, fmt.Errorf("parse ship layout: %w", err)
	}
	if len(layout.Tiles) != layout.Height {
		return nil, fmt.Errorf("tile rows (%d) != declared height (%d)", len(layout.Tiles), layout.Height)
	}
	return &layout, nil
}

// ToTileGrid converts a ShipLayout into a TileGrid.
func (l *ShipLayout) ToTileGrid() *TileGrid {
	grid := NewTileGrid(l.Width, l.Height)
	for y, row := range l.Tiles {
		for x, ch := range row {
			if x >= l.Width {
				break
			}
			grid.Set(x, y, charToTile(ch))
		}
	}
	return grid
}

// SpawnX returns the player spawn X coordinate.
func (l *ShipLayout) SpawnX() int { return l.Spawn[0] }

// SpawnY returns the player spawn Y coordinate.
func (l *ShipLayout) SpawnY() int { return l.Spawn[1] }

func charToTile(ch rune) Tile {
	switch ch {
	case '#':
		return Tile{Kind: TileWall}
	case '.':
		return Tile{Kind: TileFloor}
	case '+':
		return Tile{Kind: TileDoor}
	// --- crew quarters ---
	case 'b':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipBed)}
	case 'L':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipLocker)}
	// --- bridge stations ---
	case 'V':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipViewscreen)}
	case 'N':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipNavConsole)}
	case 'P':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipPilotConsole)}
	case 'S':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipScienceConsole)}
	// --- main deck ---
	case 'C':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipCargoConsole)}
	case 'I':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipIncinerator)}
	case 'M':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipMedical)}
	case 'F':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipFoodStation)}
	case 'D':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipDrinkStation)}
	case 't':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipToilet)}
	case 's':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipShower)}
	// --- engineering ---
	case 'G':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipOrganicTank)}
	case 'r':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipMatterRecycler)}
	case 'W':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipWaterTank)}
	case 'E':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipEngine)}
	case 'p':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipPowerCell)}
	case 'g':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipGenerator)}
	case 'f':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipFuelTank)}
	case 'J':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipJumpDrive)}
	// --- cargo ---
	case 'c':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipCargoTile)}
	case 'x':
		return Tile{Kind: TileFloor, Equipment: NewEquipment(EquipCargoTransporter)}
	default:
		return Tile{Kind: TileVoid}
	}
}

// TileWithEquipment creates a tile with the given kind and equipment instance.
func TileWithEquipment(kind TileKind, equipKind EquipmentKind) Tile {
	if equipKind == EquipNone {
		return Tile{Kind: kind}
	}
	return Tile{Kind: kind, Equipment: NewEquipment(equipKind)}
}
