package world

// TileKind represents the structural type of a tile.
type TileKind uint8

const (
	TileVoid  TileKind = iota // empty space outside the hull
	TileFloor                 // walkable floor
	TileWall                  // impassable wall
	TileDoor                  // door (walkable, can open/close)
)

// EquipmentKind identifies fixed equipment placed on a floor tile.
type EquipmentKind uint8

const (
	EquipNone           EquipmentKind = iota
	EquipConsole                      // piloting/navigation console
	EquipBed                          // sleeping berth
	EquipPowerCell                    // energy storage
	EquipEngine                       // propulsion engine
	EquipToilet                       // waste processing
	EquipShower                       // hygiene
	EquipReplicator                   // food replicator (organic → food)
	EquipWaterRecycler                // water recycler (dirty → clean)
	EquipFoodStore                    // food/organic matter storage
	EquipWaterTank                    // water storage tank
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
	return t.Kind == TileFloor || t.Kind == TileDoor
}

// Describe returns a human-readable description of a tile.
func (t Tile) Describe() string {
	if t.Equipment != EquipNone {
		return equipDescriptions[t.Equipment]
	}
	return tileDescriptions[t.Kind]
}

var tileDescriptions = map[TileKind]string{
	TileVoid:  "Empty space",
	TileFloor: "Floor",
	TileWall:  "Hull wall",
	TileDoor:  "Door",
}

var equipDescriptions = map[EquipmentKind]string{
	EquipConsole:       "Navigation Console - pilot the ship",
	EquipBed:           "Sleeping Berth - rest to recover",
	EquipPowerCell:     "Power Cell - stores energy for ship systems",
	EquipEngine:        "Engine - provides thrust and generates power",
	EquipToilet:        "Toilet - waste processing",
	EquipShower:        "Shower - hygiene station",
	EquipReplicator:    "Food Replicator - converts organic matter to meals",
	EquipWaterRecycler: "Water Recycler - purifies and recycles water",
	EquipFoodStore:     "Food Stores - organic matter supply",
	EquipWaterTank:     "Water Tank - fresh water storage",
}
