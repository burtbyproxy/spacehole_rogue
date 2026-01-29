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
	EquipDoor                          // door (On=auto-open, Open=open/closed)
	EquipAirlock                       // airlock door (exit/entry point, not auto-open)
	EquipBed                           // sleeping berth
	EquipLocker                        // storage locker (future)
	EquipViewscreen                    // displays external view
	EquipNavConsole                    // navigation station
	EquipPilotConsole                  // pilot station
	EquipScienceConsole                // science station
	EquipCargoConsole                  // cargo management terminal
	EquipCargoTransporter              // beams cargo to/from surface
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
	EquipFuelTank                      // jump fuel storage
	EquipJumpDrive                     // FTL jump drive
	EquipCargoTile                     // designated cargo storage pad
	// Surface equipment
	EquipTerminal   // interactable terminal (objective target)
	EquipLootCrate  // searchable container
	EquipObjective  // glowing objective marker
	// Prologue objectives
	EquipFuelCell   // fuel cells for shuttle
	EquipSpareParts // engine parts for repair
	EquipPowerPack  // power cell for charging
)

// Tile represents a single map tile.
type Tile struct {
	Kind      TileKind
	Equipment *Equipment // nil if no equipment, otherwise an instance
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

// SetEquipmentOn sets the on/off state for equipment at (x, y).
func (g *TileGrid) SetEquipmentOn(x, y int, on bool) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		idx := y*g.Width + x
		if g.Tiles[idx].Equipment != nil {
			g.Tiles[idx].Equipment.On = on
		}
	}
}

// ToggleEquipment toggles the on/off state at (x, y) and returns the new state.
func (g *TileGrid) ToggleEquipment(x, y int) bool {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		idx := y*g.Width + x
		if eq := g.Tiles[idx].Equipment; eq != nil {
			eq.On = !eq.On
			return eq.On
		}
	}
	return false
}

// IsEquipmentOn returns true if the equipment at (x, y) is on.
func (g *TileGrid) IsEquipmentOn(x, y int) bool {
	t := g.Get(x, y)
	return t.Equipment != nil && t.Equipment.On
}

// GetEquipment returns the equipment at (x, y), or nil if none.
func (g *TileGrid) GetEquipment(x, y int) *Equipment {
	return g.Get(x, y).Equipment
}

// AnyEquipmentOn returns true if ANY equipment of the given kind is on.
func (g *TileGrid) AnyEquipmentOn(kind EquipmentKind) bool {
	for _, t := range g.Tiles {
		if t.Equipment != nil && t.Equipment.Kind == kind && t.Equipment.On {
			return true
		}
	}
	return false
}

// SetAllEquipmentOn sets on/off state for ALL equipment of a given kind.
func (g *TileGrid) SetAllEquipmentOn(kind EquipmentKind, on bool) {
	for i := range g.Tiles {
		if eq := g.Tiles[i].Equipment; eq != nil && eq.Kind == kind {
			eq.On = on
		}
	}
}

// SetAllEquipmentState sets on/off for all toggleable equipment.
func (g *TileGrid) SetAllEquipmentState(on bool) {
	toggleable := map[EquipmentKind]bool{
		EquipEngine:           true,
		EquipGenerator:        true,
		EquipMatterRecycler:   true,
		EquipCargoTransporter: true,
		EquipNavConsole:       true,
		EquipPilotConsole:     true,
		EquipScienceConsole:   true,
		EquipCargoConsole:     true,
	}
	for i := range g.Tiles {
		if eq := g.Tiles[i].Equipment; eq != nil && toggleable[eq.Kind] {
			eq.On = on
		}
	}
}

// EquipmentKindAt returns the equipment kind at (x, y), or EquipNone if none.
func (g *TileGrid) EquipmentKindAt(x, y int) EquipmentKind {
	t := g.Get(x, y)
	if t.Equipment != nil {
		return t.Equipment.Kind
	}
	return EquipNone
}

// CountEquipment returns the number of tiles with the given equipment kind.
func (g *TileGrid) CountEquipment(kind EquipmentKind) int {
	count := 0
	for _, t := range g.Tiles {
		if t.Equipment != nil && t.Equipment.Kind == kind {
			count++
		}
	}
	return count
}

// ReservedPower returns the total power reserved by ON equipment with constant draw.
func (g *TileGrid) ReservedPower() int {
	total := 0
	for _, t := range g.Tiles {
		if eq := t.Equipment; eq != nil && eq.On && eq.PowerMode == PowerConstant {
			total += eq.PowerCost
		}
	}
	return total
}

// Describe returns a human-readable description of a tile.
func (t Tile) Describe() string {
	if t.Equipment != nil {
		return equipDescriptions[t.Equipment.Kind]
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
	EquipDoor:           "Door - E: open/close, T: toggle auto",
	EquipAirlock:        "Airlock - E: exit ship",
	EquipBed:            "Sleeping Berth - rest to recover",
	EquipLocker:         "Storage Locker - personal storage",
	EquipViewscreen:     "Viewscreen - external view display",
	EquipNavConsole:     "Nav Station - navigation and jump control",
	EquipPilotConsole:   "Pilot Station - manual flight controls",
	EquipScienceConsole: "Science Station - sensor analysis",
	EquipCargoConsole:     "Cargo Console - manage and jettison cargo",
	EquipCargoTransporter: "Cargo Transporter - beams cargo to/from surface",
	EquipIncinerator:      "Incinerator - waste disposal",
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
	EquipFuelTank:       "Fuel Tank - stores jump fuel",
	EquipJumpDrive:      "Jump Drive - FTL travel (90 power per jump!)",
	EquipCargoTile:      "Cargo Pad - designated cargo space",
	EquipTerminal:       "Terminal - data access point",
	EquipLootCrate:      "Crate - searchable container",
	EquipObjective:      "Objective - mission target",
	EquipFuelCell:       "Fuel Cells - shuttle fuel supply",
	EquipSpareParts:     "Spare Parts - engine components",
	EquipPowerPack:      "Power Pack - portable battery",
}
