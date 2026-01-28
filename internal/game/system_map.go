package game

import (
	"math"
	"math/rand/v2"
)

// SpaceObjectKind identifies what an object in a star system is.
type SpaceObjectKind uint8

const (
	ObjStar SpaceObjectKind = iota
	ObjPlanet
	ObjStation
	ObjDerelict
	ObjAsteroid
	ObjShip
)

// PlanetKind determines planet visuals and description.
type PlanetKind uint8

const (
	PlanetRocky PlanetKind = iota
	PlanetGas
	PlanetIce
	PlanetVolcanic
)

// ShipAIKind determines NPC ship behavior.
type ShipAIKind uint8

const (
	AITrader ShipAIKind = iota
	AIPatrol
	AIPirate
)

// SpaceObject is anything in a star system — star, planet, station, ship, wreck.
type SpaceObject struct {
	Kind       SpaceObjectKind
	Name       string
	X, Y       int
	PlanetType PlanetKind // only for ObjPlanet
	AIKind     ShipAIKind // only for ObjShip
	DX, DY     int        // movement direction for ships
	MoveRate   int        // ticks between moves for ships
	moveTimer  int
	dirTimer   int // ticks until next direction change
}

// System map dimensions (scrolling space, much larger than screen).
const (
	SystemMapW = 320
	SystemMapH = 160
)

// Shuttle physics constants.
// At 60 TPS: full speed ~15 tiles/sec, takes ~1 sec to accelerate, coasts to stop in ~3 sec.
const (
	ShuttleAccel    = 0.005  // thrust per tick
	ShuttleMaxSpeed = 0.25   // tiles per tick (~15 tiles/sec)
	ShuttleDrag     = 0.982  // velocity multiplier per tick (gentle space drag)
)

// ShipPhysics tracks position and velocity for ships with Newtonian-ish movement.
// Position is sub-tile (float64); rendering snaps to nearest tile via TileX/TileY.
type ShipPhysics struct {
	X, Y           float64 // sub-tile position
	VX, VY         float64 // velocity (tiles per tick)
	FaceDX, FaceDY int     // last thrust direction (for rendering glyph)

	// Ship properties
	Accel    float64 // thrust added per tick
	MaxSpeed float64 // velocity magnitude cap
	Drag     float64 // velocity multiplier per tick (1.0 = no drag)
}

// TileX returns the nearest integer tile X coordinate.
func (p *ShipPhysics) TileX() int { return int(math.Round(p.X)) }

// TileY returns the nearest integer tile Y coordinate.
func (p *ShipPhysics) TileY() int { return int(math.Round(p.Y)) }

// Speed returns the current velocity magnitude.
func (p *ShipPhysics) Speed() float64 {
	return math.Sqrt(p.VX*p.VX + p.VY*p.VY)
}

// SpeedPct returns speed as a percentage of max speed (0-100).
func (p *ShipPhysics) SpeedPct() int {
	if p.MaxSpeed <= 0 {
		return 0
	}
	pct := int(p.Speed() / p.MaxSpeed * 100)
	if pct > 100 {
		pct = 100
	}
	return pct
}

// ApplyThrust adds acceleration in the given direction (-1, 0, or 1 per axis).
// Diagonal thrust is normalized so it's not sqrt(2) faster.
func (p *ShipPhysics) ApplyThrust(dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}
	p.FaceDX = dx
	p.FaceDY = dy

	ax := float64(dx) * p.Accel
	ay := float64(dy) * p.Accel
	if dx != 0 && dy != 0 {
		ax *= 1.0 / math.Sqrt2
		ay *= 1.0 / math.Sqrt2
	}
	p.VX += ax
	p.VY += ay
}

// Tick advances physics by one step: apply drag, cap speed, move position.
func (p *ShipPhysics) Tick() {
	// Apply drag
	p.VX *= p.Drag
	p.VY *= p.Drag

	// Cap speed
	speed := p.Speed()
	if speed > p.MaxSpeed {
		scale := p.MaxSpeed / speed
		p.VX *= scale
		p.VY *= scale
	}

	// Move
	p.X += p.VX
	p.Y += p.VY

	// Kill near-zero velocity
	if math.Abs(p.VX) < 0.001 {
		p.VX = 0
	}
	if math.Abs(p.VY) < 0.001 {
		p.VY = 0
	}
}

// ClampToBounds keeps the ship within map bounds and zeroes velocity on edge hit.
func (p *ShipPhysics) ClampToBounds(w, h int) {
	minX, maxX := 1.0, float64(w-2)
	minY, maxY := 1.0, float64(h-2)
	if p.X < minX {
		p.X = minX
		p.VX = 0
	}
	if p.X > maxX {
		p.X = maxX
		p.VX = 0
	}
	if p.Y < minY {
		p.Y = minY
		p.VY = 0
	}
	if p.Y > maxY {
		p.Y = maxY
		p.VY = 0
	}
}

// SystemMap holds the contents of a star system the player can fly around in.
type SystemMap struct {
	Width, Height int
	Objects       []SpaceObject
	Shuttle       ShipPhysics  // player shuttle (float position, velocity, physics)
	Station       *StationData // generated on first dock, nil if no station
	rng           *rand.Rand
	seed          int64 // saved for station data generation
	starName      string
}

var romanNumerals = []string{"I", "II", "III", "IV", "V"}

// GenerateSystemMap creates a new system map from a seed.
func GenerateSystemMap(seed int64, starType StarType, starName string) *SystemMap {
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|1)))

	sm := &SystemMap{
		Width:  SystemMapW,
		Height: SystemMapH,
		Shuttle: ShipPhysics{
			X: 10, Y: float64(SystemMapH / 2),
			Accel: ShuttleAccel, MaxSpeed: ShuttleMaxSpeed, Drag: ShuttleDrag,
		},
		rng:      rng,
		seed:     seed,
		starName: starName,
	}

	centerX := SystemMapW / 2
	centerY := SystemMapH / 2

	// Star at center
	sm.Objects = append(sm.Objects, SpaceObject{
		Kind: ObjStar,
		Name: starName,
		X:    centerX,
		Y:    centerY,
	})

	// Planets (2-5)
	numPlanets := 2 + rng.IntN(4)
	for i := 0; i < numPlanets; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := 25.0 + rng.Float64()*90.0
		px := centerX + int(math.Cos(angle)*dist)
		py := centerY + int(math.Sin(angle)*dist*0.6) // squash Y for pseudo-perspective
		px = clampInt(px, 3, SystemMapW-3)
		py = clampInt(py, 3, SystemMapH-3)

		sm.Objects = append(sm.Objects, SpaceObject{
			Kind:       ObjPlanet,
			Name:       starName + " " + romanNumerals[i%len(romanNumerals)],
			X:          px,
			Y:          py,
			PlanetType: PlanetKind(rng.IntN(4)),
		})
	}

	// Station (50% chance, placed near a random planet)
	if rng.IntN(2) == 0 && numPlanets > 0 {
		pidx := 1 + rng.IntN(numPlanets) // +1 because index 0 is the star
		planet := sm.Objects[pidx]
		sx := planet.X + 2 + rng.IntN(3)
		sy := planet.Y - 1 + rng.IntN(3)
		sx = clampInt(sx, 2, SystemMapW-2)
		sy = clampInt(sy, 2, SystemMapH-2)

		sm.Objects = append(sm.Objects, SpaceObject{
			Kind: ObjStation,
			Name: starName + " Station",
			X:    sx,
			Y:    sy,
		})
	}

	// NPC ships (0-2)
	numShips := rng.IntN(3)
	for i := 0; i < numShips; i++ {
		kind := ShipAIKind(rng.IntN(3))
		x := 10 + rng.IntN(SystemMapW-20)
		y := 5 + rng.IntN(SystemMapH-10)

		dx := rng.IntN(3) - 1
		dy := rng.IntN(3) - 1
		if dx == 0 && dy == 0 {
			dx = 1
		}

		moveRate, dirTime := shipAIParams(kind)

		sm.Objects = append(sm.Objects, SpaceObject{
			Kind:      ObjShip,
			Name:      shipAIName(kind, rng),
			X:         x,
			Y:         y,
			AIKind:    kind,
			DX:        dx,
			DY:        dy,
			MoveRate:  moveRate,
			moveTimer: moveRate,
			dirTimer:  dirTime + rng.IntN(60),
		})
	}

	// Derelict (30% chance)
	if rng.IntN(3) == 0 {
		x := 10 + rng.IntN(SystemMapW-20)
		y := 5 + rng.IntN(SystemMapH-10)
		sm.Objects = append(sm.Objects, SpaceObject{
			Kind: ObjDerelict,
			Name: "Derelict",
			X:    x,
			Y:    y,
		})
	}

	return sm
}

func shipAIParams(kind ShipAIKind) (moveRate, dirTime int) {
	switch kind {
	case AITrader:
		return 24, 300 // 2.5 tiles/sec — slow space freighter
	case AIPatrol:
		return 14, 200 // 4.3 tiles/sec — steady patrol
	case AIPirate:
		return 7, 90 // 8.6 tiles/sec — fast, but shuttle can outrun at max speed
	default:
		return 14, 150
	}
}

func shipAIName(kind ShipAIKind, rng *rand.Rand) string {
	var names []string
	switch kind {
	case AITrader:
		names = []string{"Merchant Vessel", "Trade Runner", "Cargo Hauler", "Supply Ship"}
	case AIPatrol:
		names = []string{"Patrol Craft", "Security Corvette", "Scout Ship", "Picket Ship"}
	case AIPirate:
		names = []string{"Raider", "Corsair", "Marauder", "Brigand"}
	default:
		return "Unknown Vessel"
	}
	return names[rng.IntN(len(names))]
}

// TickNPCs advances all NPC ships by one step.
func (sm *SystemMap) TickNPCs() {
	for i := range sm.Objects {
		obj := &sm.Objects[i]
		if obj.Kind != ObjShip {
			continue
		}

		// Movement tick
		obj.moveTimer--
		if obj.moveTimer <= 0 {
			obj.moveTimer = obj.MoveRate
			obj.X += obj.DX
			obj.Y += obj.DY

			// Bounce off edges
			if obj.X <= 1 || obj.X >= sm.Width-2 {
				obj.DX = -obj.DX
				obj.X = clampInt(obj.X, 1, sm.Width-2)
			}
			if obj.Y <= 1 || obj.Y >= sm.Height-2 {
				obj.DY = -obj.DY
				obj.Y = clampInt(obj.Y, 1, sm.Height-2)
			}
		}

		// Direction change tick
		obj.dirTimer--
		if obj.dirTimer <= 0 {
			sm.changeShipDirection(obj)
		}
	}
}

func (sm *SystemMap) changeShipDirection(obj *SpaceObject) {
	sx := sm.Shuttle.TileX()
	sy := sm.Shuttle.TileY()

	switch obj.AIKind {
	case AIPirate:
		// Bias toward shuttle
		if sx > obj.X {
			obj.DX = 1
		} else if sx < obj.X {
			obj.DX = -1
		} else {
			obj.DX = 0
		}
		if sy > obj.Y {
			obj.DY = 1
		} else if sy < obj.Y {
			obj.DY = -1
		} else {
			obj.DY = 0
		}
		obj.dirTimer = 90 + sm.rng.IntN(60)

	default:
		// Random wander
		obj.DX = sm.rng.IntN(3) - 1
		obj.DY = sm.rng.IntN(3) - 1
		if obj.DX == 0 && obj.DY == 0 {
			obj.DX = 1
		}
		if obj.AIKind == AITrader {
			obj.dirTimer = 300 + sm.rng.IntN(60)
		} else {
			obj.dirTimer = 200 + sm.rng.IntN(40)
		}
	}
}

// NearestObject returns the closest object within radius of (x, y), or nil.
func (sm *SystemMap) NearestObject(x, y, radius int) *SpaceObject {
	r2 := radius * radius
	var best *SpaceObject
	bestD2 := r2 + 1
	for i := range sm.Objects {
		o := &sm.Objects[i]
		dx := o.X - x
		dy := o.Y - y
		d2 := dx*dx + dy*dy
		if d2 <= r2 && d2 < bestD2 {
			bestD2 = d2
			best = o
		}
	}
	return best
}

// EnsureStationData generates station data if this system has a station and data hasn't been created yet.
func (sm *SystemMap) EnsureStationData() *StationData {
	if sm.Station != nil {
		return sm.Station
	}
	// Find station object
	for _, obj := range sm.Objects {
		if obj.Kind == ObjStation {
			sm.Station = GenerateStationData(sm.seed+999, obj.Name)
			return sm.Station
		}
	}
	return nil // no station in this system
}

// FindStation returns the station SpaceObject in this system, or nil.
func (sm *SystemMap) FindStation() *SpaceObject {
	for i := range sm.Objects {
		if sm.Objects[i].Kind == ObjStation {
			return &sm.Objects[i]
		}
	}
	return nil
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// PlanetKindName returns a human-readable name for a planet type.
func PlanetKindName(k PlanetKind) string {
	switch k {
	case PlanetRocky:
		return "Rocky world"
	case PlanetGas:
		return "Gas giant"
	case PlanetIce:
		return "Ice world"
	case PlanetVolcanic:
		return "Volcanic world"
	default:
		return "Unknown"
	}
}

// ShipAIKindName returns a label for an NPC ship type.
func ShipAIKindName(k ShipAIKind) string {
	switch k {
	case AITrader:
		return "Trader"
	case AIPatrol:
		return "Patrol"
	case AIPirate:
		return "Pirate"
	default:
		return "Unknown"
	}
}
