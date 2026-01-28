package game

import "github.com/spacehole-rogue/spacehole_rogue/internal/world"

// SurfaceMap holds planetary surface exploration state.
type SurfaceMap struct {
	Width, Height int
	Grid          *world.TileGrid   // reuses existing TileGrid
	TerrainType   world.TerrainType // determines visual palette

	PlayerX, PlayerY   int // current player position
	ShuttleX, ShuttleY int // exit point (return to orbit)

	Objective     *SurfaceObjective // current mission objective
	LootCollected int               // count of crates searched

	Seed      int64 // for deterministic generation
	PlanetIdx int   // index in SystemMap.Objects
	POI       string // POI string that triggered this landing
}

// ObjectiveKind identifies what the player needs to do.
type ObjectiveKind uint8

const (
	ObjFindItem      ObjectiveKind = iota // locate and collect an item
	ObjReachTerminal                      // reach and interact with terminal
)

// SurfaceObjective tracks a single mission goal on a surface.
type SurfaceObjective struct {
	Kind        ObjectiveKind
	Description string
	TargetX     int       // target location X
	TargetY     int       // target location Y
	ItemKind    CargoKind // item to find (for ObjFindItem)
	Complete    bool
}

// IsWalkable checks if the player can walk to (x, y) on this surface.
func (sm *SurfaceMap) IsWalkable(x, y int) bool {
	return sm.Grid.IsWalkable(x, y)
}

// TryMove attempts to move the player by (dx, dy). Returns true if moved.
func (sm *SurfaceMap) TryMove(dx, dy int) bool {
	newX := sm.PlayerX + dx
	newY := sm.PlayerY + dy
	if sm.IsWalkable(newX, newY) {
		sm.PlayerX = newX
		sm.PlayerY = newY
		return true
	}
	return false
}

// AtShuttle returns true if the player is standing on the shuttle pad.
func (sm *SurfaceMap) AtShuttle() bool {
	return sm.PlayerX == sm.ShuttleX && sm.PlayerY == sm.ShuttleY
}

// GetTile returns the tile at (x, y).
func (sm *SurfaceMap) GetTile(x, y int) world.Tile {
	return sm.Grid.Get(x, y)
}
