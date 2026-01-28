package game

import (
	"math/rand/v2"
	"strings"

	"github.com/spacehole-rogue/spacehole_rogue/internal/world"
)

// Surface map dimensions
const (
	SurfaceWidth  = 40
	SurfaceHeight = 25
)

// GenerateSurfaceMap creates a surface map from planet and POI data.
func GenerateSurfaceMap(seed int64, planetIdx int, planetKind PlanetKind, poi string) *SurfaceMap {
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|3)))

	terrain := planetToTerrain(planetKind)

	sm := &SurfaceMap{
		Width:       SurfaceWidth,
		Height:      SurfaceHeight,
		Grid:        world.NewTileGrid(SurfaceWidth, SurfaceHeight),
		TerrainType: terrain,
		Seed:        seed,
		PlanetIdx:   planetIdx,
		POI:         poi,
	}

	// Fill with ground
	for y := 0; y < SurfaceHeight; y++ {
		for x := 0; x < SurfaceWidth; x++ {
			sm.Grid.Set(x, y, world.Tile{Kind: world.TileGround})
		}
	}

	// Scatter rocks (15-20%)
	rockDensity := 0.15 + rng.Float64()*0.05
	for y := 0; y < SurfaceHeight; y++ {
		for x := 0; x < SurfaceWidth; x++ {
			if rng.Float64() < rockDensity {
				sm.Grid.Set(x, y, world.Tile{Kind: world.TileRock})
			}
		}
	}

	// Scatter some hazards (2-5%)
	hazardDensity := 0.02 + rng.Float64()*0.03
	for y := 0; y < SurfaceHeight; y++ {
		for x := 0; x < SurfaceWidth; x++ {
			if sm.Grid.Get(x, y).Kind == world.TileGround && rng.Float64() < hazardDensity {
				sm.Grid.Set(x, y, world.Tile{Kind: world.TileHazard})
			}
		}
	}

	// Place shuttle landing pad (bottom center)
	shuttleX := SurfaceWidth / 2
	shuttleY := SurfaceHeight - 3
	clearArea(sm.Grid, shuttleX-1, shuttleY-1, 3, 3)
	sm.Grid.Set(shuttleX, shuttleY, world.Tile{Kind: world.TileShuttlePad})
	sm.ShuttleX = shuttleX
	sm.ShuttleY = shuttleY

	// Player spawns next to shuttle
	sm.PlayerX = shuttleX
	sm.PlayerY = shuttleY - 1

	// Place structure based on POI
	structure := poiToStructure(poi)
	structX := 5 + rng.IntN(SurfaceWidth-15)
	structY := 3 + rng.IntN(SurfaceHeight-12)

	// Ensure structure doesn't overlap shuttle area
	if structY+len(structure) > shuttleY-2 {
		structY = shuttleY - 2 - len(structure)
	}
	if structY < 2 {
		structY = 2
	}

	placeStructure(sm.Grid, structX, structY, structure)

	// Place objective inside structure (find the terminal or place an objective marker)
	objX, objY := findEquipmentInStructure(sm.Grid, structX, structY, structure, world.EquipTerminal)
	if objX < 0 {
		// No terminal, find a floor tile and place objective marker
		objX, objY = findFloorInStructure(sm.Grid, structX, structY, structure)
		if objX >= 0 {
			sm.Grid.Set(objX, objY, world.TileWithEquipment(world.TileFloor, world.EquipObjective))
		}
	}

	// Create objective
	if objX >= 0 {
		objKind := ObjReachTerminal
		desc := "Access the terminal"
		if sm.Grid.EquipmentKindAt(objX, objY) == world.EquipObjective {
			objKind = ObjFindItem
			desc = "Retrieve the artifact"
		}
		sm.Objective = &SurfaceObjective{
			Kind:        objKind,
			Description: desc,
			TargetX:     objX,
			TargetY:     objY,
			ItemKind:    CargoRareMinerals, // default loot
		}
	}

	return sm
}

func planetToTerrain(kind PlanetKind) world.TerrainType {
	switch kind {
	case PlanetRocky:
		return world.TerrainRocky
	case PlanetIce:
		return world.TerrainIce
	case PlanetVolcanic:
		return world.TerrainVolcanic
	default:
		return world.TerrainRocky
	}
}

// clearArea clears an area to ground tiles.
func clearArea(grid *world.TileGrid, x, y, w, h int) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			grid.Set(x+dx, y+dy, world.Tile{Kind: world.TileGround})
		}
	}
}

// Structure templates - characters map to tiles:
// # = wall, . = floor, + = door, T = terminal, L = loot crate, space = unchanged

var structureRuins = []string{
	" ##### ",
	" #...# ",
	" #.T.# ",
	" #...+ ",
	" ##### ",
}

var structureMiningRig = []string{
	"########",
	"#L....L#",
	"#......#",
	"#..TT..+",
	"#......#",
	"#L....L#",
	"########",
}

var structureResearchLab = []string{
	"#######",
	"#.....#",
	"#.TTT.#",
	"#.....+",
	"#L...L#",
	"#######",
}

var structureCrashSite = []string{
	"  ###  ",
	" #L..# ",
	"#....#+",
	" #..T# ",
	"  ###  ",
}

var structureOutpost = []string{
	"#####",
	"#...#",
	"#.T.+",
	"#...#",
	"#####",
}

func poiToStructure(poi string) []string {
	poiLower := strings.ToLower(poi)
	switch {
	case strings.Contains(poiLower, "ancient") || strings.Contains(poiLower, "structure"):
		return structureRuins
	case strings.Contains(poiLower, "mining") || strings.Contains(poiLower, "rig"):
		return structureMiningRig
	case strings.Contains(poiLower, "thermal") || strings.Contains(poiLower, "research"):
		return structureResearchLab
	case strings.Contains(poiLower, "debris") || strings.Contains(poiLower, "crash"):
		return structureCrashSite
	default:
		return structureOutpost
	}
}

func placeStructure(grid *world.TileGrid, startX, startY int, template []string) {
	for dy, row := range template {
		for dx, ch := range row {
			x := startX + dx
			y := startY + dy
			tile := charToTile(ch)
			if tile.Kind != world.TileVoid { // space = don't change
				grid.Set(x, y, tile)
			}
		}
	}
}

func charToTile(ch rune) world.Tile {
	switch ch {
	case '#':
		return world.Tile{Kind: world.TileWall}
	case '.':
		return world.Tile{Kind: world.TileFloor}
	case '+':
		return world.Tile{Kind: world.TileDoor}
	case 'T':
		return world.TileWithEquipment(world.TileFloor, world.EquipTerminal)
	case 'L':
		return world.TileWithEquipment(world.TileFloor, world.EquipLootCrate)
	default:
		return world.Tile{Kind: world.TileVoid} // unchanged
	}
}

func findEquipmentInStructure(grid *world.TileGrid, startX, startY int, template []string, equip world.EquipmentKind) (int, int) {
	for dy, row := range template {
		for dx := range row {
			x := startX + dx
			y := startY + dy
			if grid.EquipmentKindAt(x, y) == equip {
				return x, y
			}
		}
	}
	return -1, -1
}

func findFloorInStructure(grid *world.TileGrid, startX, startY int, template []string) (int, int) {
	for dy, row := range template {
		for dx, ch := range row {
			if ch == '.' {
				x := startX + dx
				y := startY + dy
				if grid.Get(x, y).Kind == world.TileFloor && grid.Get(x, y).Equipment == nil {
					return x, y
				}
			}
		}
	}
	return -1, -1
}
