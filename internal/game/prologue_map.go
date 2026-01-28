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

	// Place objectives based on what's needed
	objectives := scenario.GetObjectives()
	ps := &PrologueSurface{
		SurfaceMap:     sm,
		Scenario:       scenario,
		ObjectivesLeft: objectives,
	}

	// Place objective items around the map
	placeObjectiveItems(sm.Grid, objectives, shuttleX, shuttleY, rng)

	// Create a multi-objective description
	sm.Objective = &SurfaceObjective{
		Kind:        ObjFindItem,
		Description: scenario.Objective,
		Complete:    false,
	}

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

	// Maybe add some equipment
	if rng.Float64() < 0.4 {
		eqX := x + 1 + rng.IntN(w-2)
		eqY := y + 1 + rng.IntN(h-2)
		if grid.Get(eqX, eqY).Kind == world.TileFloor {
			grid.Set(eqX, eqY, world.TileWithEquipment(world.TileFloor, world.EquipLootCrate))
		}
	}
}

// placeSmallRuin places a small ruin structure.
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
		}
	}
}

// placeObjectiveItems places the prologue objectives around the map.
func placeObjectiveItems(grid *world.TileGrid, objectives []PrologueObjectiveKind, shuttleX, shuttleY int, rng *rand.Rand) {
	// Find valid positions away from shuttle
	var positions [][2]int
	for y := 3; y < grid.Height-5; y++ {
		for x := 3; x < grid.Width-3; x++ {
			// Skip near shuttle
			if abs(x-shuttleX) < 8 && abs(y-shuttleY) < 5 {
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
