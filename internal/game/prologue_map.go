package game

import (
	"math/rand/v2"

	"github.com/spacehole-rogue/spacehole_rogue/internal/world"
)

// Prologue map dimensions (slightly larger than regular surface)
const (
	PrologueMapWidth  = 50
	PrologueMapHeight = 30
)

// PrologueSurface extends SurfaceMap with prologue-specific state.
type PrologueSurface struct {
	*SurfaceMap
	Scenario       *PrologueScenario
	ObjectivesLeft []PrologueObjectiveKind // what's still needed
	FuelFound      bool
	PartsFound     bool
	PowerFound     bool
}

// GeneratePrologueMap creates the starting surface map based on the prologue scenario.
func GeneratePrologueMap(scenario *PrologueScenario, seed int64) *PrologueSurface {
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|11)))

	// Determine terrain type
	var terrain world.TerrainType
	switch scenario.Location {
	case LocIcePlanet:
		terrain = world.TerrainIce
	case LocVolcanicPlanet:
		terrain = world.TerrainVolcanic
	case LocAbandonedOutpost, LocDerelictStation:
		terrain = world.TerrainInterior
	default:
		terrain = world.TerrainRocky
	}

	sm := &SurfaceMap{
		Width:       PrologueMapWidth,
		Height:      PrologueMapHeight,
		Grid:        world.NewTileGrid(PrologueMapWidth, PrologueMapHeight),
		TerrainType: terrain,
		Seed:        seed,
		PlanetIdx:   -1, // not a real planet
		POI:         "Starting Location",
	}

	// Fill based on location type
	isInterior := scenario.IsInterior()
	if isInterior {
		fillInteriorMap(sm.Grid, rng)
	} else {
		fillPlanetMap(sm.Grid, terrain, rng)
	}

	// Place the broken shuttle
	var shuttleX, shuttleY int
	if isInterior {
		// For interior maps, place shuttle at end of main corridor
		corridorY := PrologueMapHeight / 2
		shuttleX = PrologueMapWidth/2 + rng.IntN(5) - 2
		shuttleY = corridorY
		// Clear a small docking area
		for dx := -1; dx <= 1; dx++ {
			sm.Grid.Set(shuttleX+dx, shuttleY, world.Tile{Kind: world.TileFloor})
		}
		sm.Grid.Set(shuttleX, shuttleY, world.Tile{Kind: world.TileShuttlePad})
		sm.ShuttleX = shuttleX
		sm.ShuttleY = shuttleY
		// Player spawns adjacent on corridor
		sm.PlayerX = shuttleX + 2
		sm.PlayerY = shuttleY
	} else {
		// Outdoor: shuttle near bottom center
		shuttleX = PrologueMapWidth/2 + rng.IntN(5) - 2
		shuttleY = PrologueMapHeight - 5
		clearArea(sm.Grid, shuttleX-2, shuttleY-1, 5, 3)
		sm.Grid.Set(shuttleX, shuttleY, world.Tile{Kind: world.TileShuttlePad})
		sm.ShuttleX = shuttleX
		sm.ShuttleY = shuttleY
		// Player spawns near shuttle
		sm.PlayerX = shuttleX
		sm.PlayerY = shuttleY - 1
	}

	// Ensure all structure doors are reachable from player spawn
	ensureDoorsReachable(sm.Grid, sm.PlayerX, sm.PlayerY)

	// Place objectives based on what's needed
	objectives := scenario.GetObjectives()
	ps := &PrologueSurface{
		SurfaceMap:     sm,
		Scenario:       scenario,
		ObjectivesLeft: objectives,
	}

	// Place objective items around the map (only in reachable positions)
	placeObjectiveItems(sm.Grid, objectives, shuttleX, shuttleY, sm.PlayerX, sm.PlayerY, rng)

	// Place loot crates in reachable positions
	placeLootCrates(sm.Grid, sm.PlayerX, sm.PlayerY, rng)

	// Create a multi-objective description
	sm.Objective = &SurfaceObjective{
		Kind:        ObjFindItem,
		Description: scenario.Objective,
		Complete:    false,
	}

	// Initialize fog of war
	sm.InitVisibility()
	sm.UpdateVisibility()

	return ps
}

// fillPlanetMap fills with outdoor terrain.
func fillPlanetMap(grid *world.TileGrid, terrain world.TerrainType, rng *rand.Rand) {
	// Fill with ground
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			grid.Set(x, y, world.Tile{Kind: world.TileGround})
		}
	}

	// Scatter rocks (12-18%)
	rockDensity := 0.12 + rng.Float64()*0.06
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			if rng.Float64() < rockDensity {
				grid.Set(x, y, world.Tile{Kind: world.TileRock})
			}
		}
	}

	// Scatter hazards based on terrain (2-4%)
	hazardDensity := 0.02 + rng.Float64()*0.02
	if terrain == world.TerrainVolcanic {
		hazardDensity = 0.04 + rng.Float64()*0.03 // more hazards on volcanic
	}
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			if grid.Get(x, y).Kind == world.TileGround && rng.Float64() < hazardDensity {
				grid.Set(x, y, world.Tile{Kind: world.TileHazard})
			}
		}
	}

	// Place some ruins/structures for scavenging
	numStructures := 2 + rng.IntN(2)
	for i := 0; i < numStructures; i++ {
		structX := 5 + rng.IntN(grid.Width-15)
		structY := 3 + rng.IntN(grid.Height-12)
		placeSmallRuin(grid, structX, structY, rng)
	}
}

// fillInteriorMap fills with interior/station terrain.
func fillInteriorMap(grid *world.TileGrid, rng *rand.Rand) {
	// Fill with void first
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			grid.Set(x, y, world.Tile{Kind: world.TileVoid})
		}
	}

	// Create a rough station layout with rooms
	// Main corridor
	corridorY := grid.Height / 2
	for x := 5; x < grid.Width-5; x++ {
		grid.Set(x, corridorY, world.Tile{Kind: world.TileFloor})
		grid.Set(x, corridorY-1, world.Tile{Kind: world.TileWall})
		grid.Set(x, corridorY+1, world.Tile{Kind: world.TileWall})
	}

	// Add rooms branching off
	numRooms := 4 + rng.IntN(3)
	for i := 0; i < numRooms; i++ {
		roomX := 8 + rng.IntN(grid.Width-20)
		roomW := 5 + rng.IntN(4)
		roomH := 4 + rng.IntN(3)

		// Alternate above/below corridor
		var roomY int
		if i%2 == 0 {
			roomY = corridorY - 2 - roomH
		} else {
			roomY = corridorY + 2
		}

		placeRoom(grid, roomX, roomY, roomW, roomH, corridorY, rng)
	}

	// Add some hazards (breaches)
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			if grid.Get(x, y).Kind == world.TileFloor && rng.Float64() < 0.02 {
				grid.Set(x, y, world.Tile{Kind: world.TileHazard})
			}
		}
	}
}

// placeRoom carves a room and connects it to the corridor.
func placeRoom(grid *world.TileGrid, x, y, w, h, corridorY int, rng *rand.Rand) {
	// Carve room
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			if dy == 0 || dy == h-1 || dx == 0 || dx == w-1 {
				grid.Set(x+dx, y+dy, world.Tile{Kind: world.TileWall})
			} else {
				grid.Set(x+dx, y+dy, world.Tile{Kind: world.TileFloor})
			}
		}
	}

	// Connect to corridor with door
	doorX := x + 1 + rng.IntN(w-2)
	if y < corridorY {
		// Room above corridor
		for cy := y + h - 1; cy <= corridorY-1; cy++ {
			grid.Set(doorX, cy, world.Tile{Kind: world.TileFloor})
		}
		grid.Set(doorX, y+h-1, world.Tile{Kind: world.TileDoor})
	} else {
		// Room below corridor
		for cy := corridorY + 1; cy <= y; cy++ {
			grid.Set(doorX, cy, world.Tile{Kind: world.TileFloor})
		}
		grid.Set(doorX, y, world.Tile{Kind: world.TileDoor})
	}
	// Loot crates are placed separately via placeLootCrates() after reachability check
}

// placeSmallRuin places a small ruin structure and clears an approach path to its door.
func placeSmallRuin(grid *world.TileGrid, x, y int, rng *rand.Rand) {
	templates := [][]string{
		{
			"####",
			"#..#",
			"#..+",
			"####",
		},
		{
			" ### ",
			"##.##",
			"#...+",
			"##.##",
			" ### ",
		},
		{
			"#####",
			"#L..#",
			"#...+",
			"#####",
		},
	}

	template := templates[rng.IntN(len(templates))]
	var doorX, doorY int
	for dy, row := range template {
		for dx, ch := range row {
			px, py := x+dx, y+dy
			if px < 0 || px >= grid.Width || py < 0 || py >= grid.Height {
				continue
			}
			tile := charToTile(ch)
			if tile.Kind != world.TileVoid {
				grid.Set(px, py, tile)
			}
			// Track door position
			if ch == '+' {
				doorX, doorY = px, py
			}
		}
	}

	// Clear an approach path from the door (doors face right in these templates)
	// Clear 3-4 tiles to the right of the door to ensure access
	for dx := 1; dx <= 4; dx++ {
		px := doorX + dx
		if px >= 0 && px < grid.Width && doorY >= 0 && doorY < grid.Height {
			tile := grid.Get(px, doorY)
			// Only clear rocks/hazards, don't overwrite other structures
			if tile.Kind == world.TileRock || tile.Kind == world.TileHazard {
				grid.Set(px, doorY, world.Tile{Kind: world.TileGround})
			}
		}
	}
}

// ensureDoorsReachable finds all doors and carves paths to any that aren't reachable.
func ensureDoorsReachable(grid *world.TileGrid, playerX, playerY int) {
	// Find all reachable positions
	reachable := floodFillReachable(grid, playerX, playerY)

	// Find all door positions
	var doors [][2]int
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			if grid.Get(x, y).Kind == world.TileDoor {
				doors = append(doors, [2]int{x, y})
			}
		}
	}

	// For each unreachable door, carve a path to it
	for _, door := range doors {
		dx, dy := door[0], door[1]
		if reachable[door] {
			continue // Already reachable
		}

		// Find nearest reachable tile using BFS
		nearest := findNearestReachable(grid, dx, dy, reachable)
		if nearest[0] == -1 {
			continue // No reachable tile found (shouldn't happen)
		}

		// Carve a straight-ish path from nearest reachable to the door
		carvePath(grid, nearest[0], nearest[1], dx, dy)

		// Update reachability after carving
		reachable = floodFillReachable(grid, playerX, playerY)
	}
}

// findNearestReachable finds the nearest reachable tile to (startX, startY).
func findNearestReachable(grid *world.TileGrid, startX, startY int, reachable map[[2]int]bool) [2]int {
	visited := make(map[[2]int]bool)
	queue := [][2]int{{startX, startY}}

	for len(queue) > 0 {
		pos := queue[0]
		queue = queue[1:]

		if visited[pos] {
			continue
		}
		visited[pos] = true

		x, y := pos[0], pos[1]
		if x < 0 || x >= grid.Width || y < 0 || y >= grid.Height {
			continue
		}

		// Found a reachable tile
		if reachable[pos] {
			return pos
		}

		// Expand search (including through walls/rocks for search purposes)
		queue = append(queue, [2]int{x - 1, y})
		queue = append(queue, [2]int{x + 1, y})
		queue = append(queue, [2]int{x, y - 1})
		queue = append(queue, [2]int{x, y + 1})
	}

	return [2]int{-1, -1} // Not found
}

// carvePath clears a path from (x1,y1) to (x2,y2) by clearing rocks/hazards.
func carvePath(grid *world.TileGrid, x1, y1, x2, y2 int) {
	x, y := x1, y1
	for x != x2 || y != y2 {
		// Clear this tile if it's blocking
		tile := grid.Get(x, y)
		if tile.Kind == world.TileRock || tile.Kind == world.TileHazard {
			grid.Set(x, y, world.Tile{Kind: world.TileGround})
		}

		// Move toward target (simple approach: horizontal then vertical)
		if x < x2 {
			x++
		} else if x > x2 {
			x--
		} else if y < y2 {
			y++
		} else if y > y2 {
			y--
		}
	}
}

// floodFillReachable returns a set of all positions reachable from (startX, startY).
func floodFillReachable(grid *world.TileGrid, startX, startY int) map[[2]int]bool {
	reachable := make(map[[2]int]bool)
	queue := [][2]int{{startX, startY}}

	for len(queue) > 0 {
		pos := queue[0]
		queue = queue[1:]

		if reachable[pos] {
			continue
		}

		x, y := pos[0], pos[1]
		if x < 0 || x >= grid.Width || y < 0 || y >= grid.Height {
			continue
		}

		if !grid.IsWalkable(x, y) {
			continue
		}

		reachable[pos] = true

		// Add neighbors
		queue = append(queue, [2]int{x - 1, y})
		queue = append(queue, [2]int{x + 1, y})
		queue = append(queue, [2]int{x, y - 1})
		queue = append(queue, [2]int{x, y + 1})
	}

	return reachable
}

// placeObjectiveItems places the prologue objectives around the map.
// Only places items in positions reachable from the player spawn.
func placeObjectiveItems(grid *world.TileGrid, objectives []PrologueObjectiveKind, shuttleX, shuttleY, playerX, playerY int, rng *rand.Rand) {
	// First, find all reachable positions from player spawn
	reachable := floodFillReachable(grid, playerX, playerY)

	// Find valid positions away from shuttle that are reachable
	var positions [][2]int
	for y := 3; y < grid.Height-5; y++ {
		for x := 3; x < grid.Width-3; x++ {
			// Skip near shuttle
			if abs(x-shuttleX) < 8 && abs(y-shuttleY) < 5 {
				continue
			}
			// Must be reachable
			if !reachable[[2]int{x, y}] {
				continue
			}
			if grid.Get(x, y).Kind == world.TileFloor || grid.Get(x, y).Kind == world.TileGround {
				if grid.Get(x, y).Equipment == nil {
					positions = append(positions, [2]int{x, y})
				}
			}
		}
	}

	// Shuffle positions
	rng.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Place each objective
	for i, obj := range objectives {
		if i >= len(positions) {
			break
		}
		pos := positions[i]
		var equip world.EquipmentKind
		switch obj {
		case PrologueObjFuel:
			equip = world.EquipFuelCell
		case PrologueObjParts:
			equip = world.EquipSpareParts
		case PrologueObjPower:
			equip = world.EquipPowerPack
		}
		grid.Set(pos[0], pos[1], world.TileWithEquipment(grid.Get(pos[0], pos[1]).Kind, equip))
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// placeLootCrates places loot crates in reachable floor positions.
func placeLootCrates(grid *world.TileGrid, playerX, playerY int, rng *rand.Rand) {
	// Find all reachable positions from player spawn
	reachable := floodFillReachable(grid, playerX, playerY)

	// Find valid floor positions for loot crates
	var positions [][2]int
	for y := 1; y < grid.Height-1; y++ {
		for x := 1; x < grid.Width-1; x++ {
			// Must be reachable
			if !reachable[[2]int{x, y}] {
				continue
			}
			// Must be floor tile without equipment
			tile := grid.Get(x, y)
			if tile.Kind == world.TileFloor && tile.Equipment == nil {
				positions = append(positions, [2]int{x, y})
			}
		}
	}

	// Shuffle positions
	rng.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Place 2-4 loot crates
	numCrates := 2 + rng.IntN(3)
	for i := 0; i < numCrates && i < len(positions); i++ {
		pos := positions[i]
		grid.Set(pos[0], pos[1], world.TileWithEquipment(world.TileFloor, world.EquipLootCrate))
	}
}

// CheckPrologueComplete checks if all objectives are found.
func (ps *PrologueSurface) CheckPrologueComplete() bool {
	needed := ps.Scenario.GetObjectives()
	for _, obj := range needed {
		switch obj {
		case PrologueObjFuel:
			if !ps.FuelFound {
				return false
			}
		case PrologueObjParts:
			if !ps.PartsFound {
				return false
			}
		case PrologueObjPower:
			if !ps.PowerFound {
				return false
			}
		}
	}
	return true
}

// MarkObjectiveFound marks a prologue objective as found.
func (ps *PrologueSurface) MarkObjectiveFound(kind PrologueObjectiveKind) {
	switch kind {
	case PrologueObjFuel:
		ps.FuelFound = true
	case PrologueObjParts:
		ps.PartsFound = true
	case PrologueObjPower:
		ps.PowerFound = true
	}
	// Update remaining list
	var remaining []PrologueObjectiveKind
	for _, obj := range ps.ObjectivesLeft {
		found := false
		switch obj {
		case PrologueObjFuel:
			found = ps.FuelFound
		case PrologueObjParts:
			found = ps.PartsFound
		case PrologueObjPower:
			found = ps.PowerFound
		}
		if !found {
			remaining = append(remaining, obj)
		}
	}
	ps.ObjectivesLeft = remaining

	// Update objective complete status
	if ps.CheckPrologueComplete() && ps.SurfaceMap.Objective != nil {
		ps.SurfaceMap.Objective.Complete = true
	}
}

// ObjectiveStatus returns a human-readable status of prologue objectives.
func (ps *PrologueSurface) ObjectiveStatus() string {
	if ps.CheckPrologueComplete() {
		return "Shuttle ready - return to launch!"
	}

	status := "Still need:"
	needed := ps.Scenario.GetObjectives()
	for _, obj := range needed {
		found := false
		name := ""
		switch obj {
		case PrologueObjFuel:
			found = ps.FuelFound
			name = "Fuel"
		case PrologueObjParts:
			found = ps.PartsFound
			name = "Parts"
		case PrologueObjPower:
			found = ps.PowerFound
			name = "Power"
		}
		if found {
			status += " [X]" + name
		} else {
			status += " [ ]" + name
		}
	}
	return status
}
