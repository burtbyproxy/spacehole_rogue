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

	Seed      int64  // for deterministic generation
	PlanetIdx int    // index in SystemMap.Objects
	POI       string // POI string that triggered this landing

	// Fog of war - visibility tracking
	Visible []bool // tiles currently visible from player position
	Seen    []bool // tiles player has ever seen (memory)
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
		sm.UpdateVisibility() // Update fog of war
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

// InitVisibility initializes the visibility tracking slices.
func (sm *SurfaceMap) InitVisibility() {
	size := sm.Width * sm.Height
	sm.Visible = make([]bool, size)
	sm.Seen = make([]bool, size)
}

// UpdateVisibility recalculates which tiles are visible from the player position.
// Visibility spreads through walkable tiles but stops at walls and doors.
func (sm *SurfaceMap) UpdateVisibility() {
	if sm.Visible == nil {
		sm.InitVisibility()
	}

	// Clear current visibility
	for i := range sm.Visible {
		sm.Visible[i] = false
	}

	// Flood fill from player position, stopping at walls and doors
	visited := make(map[[2]int]bool)
	queue := [][2]int{{sm.PlayerX, sm.PlayerY}}

	for len(queue) > 0 {
		pos := queue[0]
		queue = queue[1:]

		if visited[pos] {
			continue
		}
		visited[pos] = true

		x, y := pos[0], pos[1]
		if x < 0 || x >= sm.Width || y < 0 || y >= sm.Height {
			continue
		}

		tile := sm.Grid.Get(x, y)

		// Mark as visible
		idx := y*sm.Width + x
		sm.Visible[idx] = true
		sm.Seen[idx] = true

		// Stop spreading at walls, rocks, hazards, and doors
		// (We can see the door/wall itself, but not through it)
		if tile.Kind == world.TileWall || tile.Kind == world.TileRock ||
			tile.Kind == world.TileHazard || tile.Kind == world.TileDoor {
			continue
		}

		// Spread to neighbors
		queue = append(queue, [2]int{x - 1, y})
		queue = append(queue, [2]int{x + 1, y})
		queue = append(queue, [2]int{x, y - 1})
		queue = append(queue, [2]int{x, y + 1})
	}
}

// IsVisible returns true if the tile at (x, y) is currently visible.
func (sm *SurfaceMap) IsVisible(x, y int) bool {
	if sm.Visible == nil || x < 0 || x >= sm.Width || y < 0 || y >= sm.Height {
		return false
	}
	return sm.Visible[y*sm.Width+x]
}

// IsSeen returns true if the tile at (x, y) has ever been seen.
func (sm *SurfaceMap) IsSeen(x, y int) bool {
	if sm.Seen == nil || x < 0 || x >= sm.Width || y < 0 || y >= sm.Height {
		return false
	}
	return sm.Seen[y*sm.Width+x]
}
