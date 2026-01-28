package world

// TerrainType determines the visual palette for surface tiles.
type TerrainType uint8

const (
	TerrainRocky    TerrainType = iota // brown/gray rocks
	TerrainIce                          // cyan/white ice
	TerrainVolcanic                     // red/orange lava
	TerrainInterior                     // standard ship/base interior
)

// TileKind represents the structural type of a tile.
type TileKind uint8

const (
	TileVoid  TileKind = iota // empty space outside the hull
	TileFloor                 // walkable floor
	TileWall                  // impassable wall
	TileDoor                  // door (walkable, can open/close)
	// Surface terrain tiles
	TileGround     // walkable terrain (color varies by TerrainType)
	TileRock       // impassable terrain obstacle
	TileHazard     // impassable hazard (lava, crevasse)
	TileShuttlePad // landing/exit point (walkable)
)

// EquipmentKind identifies fixed equipment placed on a floor tile.
type EquipmentKind uint8

const (
	EquipNone            EquipmentKind = iota
	EquipBed                           // sleeping berth
	EquipLocker                        // storage locker (future)
	EquipViewscreen                    // displays external view
	EquipNavConsole                    // navigation station
	EquipPilotConsole                  // pilot station
	EquipScienceConsole                // science station
	EquipCargoConsole                  // cargo management terminal
	EquipIncinerator                   // waste disposal (future)
	EquipMedical                       // medical station (future)
	EquipFoodStation                   // food replicator (clean organic → body)
	EquipDrinkStation                  // drink replicator (clean water → body)
	EquipToilet                        // waste processing
	EquipShower                        // hygiene
	EquipOrganicTank                   // organic matter storage
	EquipMatterRecycler                // combined recycler (dirty → clean, water + organic)
	EquipWaterTank                     // water storage tank
	EquipEngine                        // propulsion
	EquipPowerCell                     // energy storage (battery)
	EquipGenerator                     // power generator
	EquipCargoTile                     // designated cargo storage pad
	// Surface equipment
	EquipTerminal   // interactable terminal (objective target)
	EquipLootCrate  // searchable container
	EquipObjective  // glowing objective marker
)

// Tile represents a single map tile.
type Tile struct {
	Kind      TileKind
	Equipment EquipmentKind
}

// TileGrid is a 2D grid of tiles.
type TileGrid struct {
	Width  int
	Height int
	Tiles  []Tile
}

// NewTileGrid creates an empty tile grid filled with void.
func NewTileGrid(w, h int) *TileGrid {
	return &TileGrid{
		Width:  w,
		Height: h,
		Tiles:  make([]Tile, w*h),
	}
}

// Get returns the tile at (x, y). Out-of-bounds returns void.
func (g *TileGrid) Get(x, y int) Tile {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return Tile{Kind: TileVoid}
	}
	return g.Tiles[y*g.Width+x]
}

// Set writes a tile at (x, y). Out-of-bounds writes are ignored.
func (g *TileGrid) Set(x, y int, t Tile) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.Tiles[y*g.Width+x] = t
	}
}

// IsWalkable returns true if an entity can walk on (x, y).
func (g *TileGrid) IsWalkable(x, y int) bool {
	t := g.Get(x, y)
	switch t.Kind {
	case TileFloor, TileDoor, TileGround, TileShuttlePad:
		return true
	default:
		return false
	}
}

// Describe returns a human-readable description of a tile.
func (t Tile) Describe() string {
	if t.Equipment != EquipNone {
		return equipDescriptions[t.Equipment]
	}
	return tileDescriptions[t.Kind]
}

var tileDescriptions = map[TileKind]string{
	TileVoid:       "Empty space",
	TileFloor:      "Floor",
	TileWall:       "Hull wall",
	TileDoor:       "Door",
	TileGround:     "Ground",
	TileRock:       "Rock formation",
	TileHazard:     "Hazardous terrain",
	TileShuttlePad: "Shuttle landing pad",
}

var equipDescriptions = map[EquipmentKind]string{
	EquipBed:            "Sleeping Berth - rest to recover",
	EquipLocker:         "Storage Locker - personal storage",
	EquipViewscreen:     "Viewscreen - external view display",
	EquipNavConsole:     "Nav Station - navigation and jump control",
	EquipPilotConsole:   "Pilot Station - manual flight controls",
	EquipScienceConsole: "Science Station - sensor analysis",
	EquipCargoConsole:   "Cargo Console - manage and jettison cargo",
	EquipIncinerator:    "Incinerator - waste disposal",
	EquipMedical:        "Medical Station - treat injuries",
	EquipFoodStation:    "Food Replicator - dispenses meals from clean organics",
	EquipDrinkStation:   "Drink Replicator - dispenses clean water",
	EquipToilet:         "Toilet - waste processing",
	EquipShower:         "Shower - hygiene station",
	EquipOrganicTank:    "Organic Tank - organic matter storage",
	EquipMatterRecycler: "Matter Recycler - converts dirty matter to clean",
	EquipWaterTank:      "Water Tank - fresh water storage",
	EquipEngine:         "Engine - provides thrust",
	EquipPowerCell:      "Battery - stores energy for ship systems",
	EquipGenerator:      "Generator - produces energy",
	EquipCargoTile:      "Cargo Pad - designated cargo space",
	EquipTerminal:       "Terminal - data access point",
	EquipLootCrate:      "Crate - searchable container",
	EquipObjective:      "Objective - mission target",
}
