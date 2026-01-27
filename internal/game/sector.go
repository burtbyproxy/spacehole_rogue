package game

import (
	"math"
	"math/rand/v2"
)

// StarType determines the visual color of a star on the sector map.
type StarType uint8

const (
	StarYellow StarType = iota
	StarRed
	StarBlue
	StarWhite
	StarOrange
)

// StarSystem represents a single star system on the sector map.
type StarSystem struct {
	Name    string
	X, Y    int // position on the sector map grid
	Type    StarType
	Visited bool
	Map     *SystemMap // nil until first visit, then lazy-generated
}

// Sector holds the generated star map and current navigation state.
type Sector struct {
	Systems       []StarSystem
	CurrentSystem int // index — where the player is now
	CursorSystem  int // index — where the cursor is pointing
	Seed          int64
}

// Sector map bounds (within the 80x45 grid)
const (
	sectorMinX = 4
	sectorMaxX = 54
	sectorMinY = 4
	sectorMaxY = 34
	minStarDist = 4 // minimum distance between stars
)

var starNames = []string{
	"Vega Prime", "Kepler's Rest", "Nyx", "Caelum", "Draconis",
	"Forge", "Hadal Deep", "Meridian", "Obsidian", "Solis",
	"Tempest", "Umbra", "Zenith", "Arcturus", "Cygnus",
	"Eridani", "Lyra", "Procyon", "Rigel", "Sirius",
}

// NewSector generates a sector from a seed.
func NewSector(seed int64) *Sector {
	rng := rand.New(rand.NewPCG(uint64(seed), 0))

	numSystems := 12 + rng.IntN(4) // 12-15

	// Shuffle names
	names := make([]string, len(starNames))
	copy(names, starNames)
	rng.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	systems := make([]StarSystem, 0, numSystems)

	// Starting system near center
	systems = append(systems, StarSystem{
		Name:    names[0],
		X:       (sectorMinX + sectorMaxX) / 2,
		Y:       (sectorMinY + sectorMaxY) / 2,
		Type:    StarYellow,
		Visited: true,
	})

	// Generate remaining systems
	for i := 1; i < numSystems; i++ {
		var x, y int
		for attempts := 0; attempts < 100; attempts++ {
			x = sectorMinX + rng.IntN(sectorMaxX-sectorMinX+1)
			y = sectorMinY + rng.IntN(sectorMaxY-sectorMinY+1)
			if !tooClose(systems, x, y) {
				break
			}
		}

		systems = append(systems, StarSystem{
			Name:    names[i%len(names)],
			X:       x,
			Y:       y,
			Type:    StarType(rng.IntN(5)),
			Visited: false,
		})
	}

	return &Sector{
		Systems:       systems,
		CurrentSystem: 0,
		CursorSystem:  0,
		Seed:          seed,
	}
}

func tooClose(systems []StarSystem, x, y int) bool {
	for _, s := range systems {
		dx := s.X - x
		dy := s.Y - y
		if dx*dx+dy*dy < minStarDist*minStarDist {
			return true
		}
	}
	return false
}

// DistanceBetween returns the Euclidean distance between two star systems.
func (s *Sector) DistanceBetween(a, b int) float64 {
	sa := s.Systems[a]
	sb := s.Systems[b]
	dx := float64(sa.X - sb.X)
	dy := float64(sa.Y - sb.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// EnergyCostTo returns the energy cost to travel from current system to target.
func (s *Sector) EnergyCostTo(target int) int {
	dist := s.DistanceBetween(s.CurrentSystem, target)
	cost := int(dist * 1.5)
	if cost < 5 {
		cost = 5
	}
	return cost
}

// NearestInDirection returns the index of the nearest star from the cursor
// in the given direction (dx, dy), or -1 if none found.
func (s *Sector) NearestInDirection(dx, dy int) int {
	cur := s.Systems[s.CursorSystem]
	bestIdx := -1
	bestDist := math.MaxFloat64

	for i, sys := range s.Systems {
		if i == s.CursorSystem {
			continue
		}
		relX := sys.X - cur.X
		relY := sys.Y - cur.Y

		// Candidate must be in the requested direction
		if dx > 0 && relX <= 0 {
			continue
		}
		if dx < 0 && relX >= 0 {
			continue
		}
		if dy > 0 && relY <= 0 {
			continue
		}
		if dy < 0 && relY >= 0 {
			continue
		}

		dist := math.Sqrt(float64(relX*relX + relY*relY))
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}
	return bestIdx
}

// EnsureSystemMap generates the system map for a star system if it doesn't exist yet.
func (s *Sector) EnsureSystemMap(idx int) {
	sys := &s.Systems[idx]
	if sys.Map != nil {
		return
	}
	seed := s.Seed*1000 + int64(idx)
	sys.Map = GenerateSystemMap(seed, sys.Type, sys.Name)
}

// CurrentSystemMap returns the system map for the current star system, generating it if needed.
func (s *Sector) CurrentSystemMap() *SystemMap {
	s.EnsureSystemMap(s.CurrentSystem)
	return s.Systems[s.CurrentSystem].Map
}

// StarTypeName returns a human-readable name for a star type.
func StarTypeName(t StarType) string {
	switch t {
	case StarYellow:
		return "Yellow dwarf"
	case StarRed:
		return "Red giant"
	case StarBlue:
		return "Blue supergiant"
	case StarWhite:
		return "White dwarf"
	case StarOrange:
		return "Orange star"
	default:
		return "Unknown"
	}
}
