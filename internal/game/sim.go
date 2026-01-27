package game

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"
	"github.com/spacehole-rogue/spacehole_rogue/internal/world"
)

// Time scale: 1 game day = 20 real minutes = 72,000 ticks at 60 TPS.
// 1 game hour = 50 real seconds = 3,000 ticks.

// Tick intervals (at 60 TPS)
const (
	generatorInterval       = 60   // generator produces 1 energy per real second
	recyclerIntakeInterval  = 60   // recycler pulls dirty matter into buffer every 1 sec
	recyclerProcessInterval = 100  // recycler converts 1 buffered dirty → clean every ~1.7 sec
	bodyDigestInterval      = 300  // body converts 1 clean organic → waste every 5 sec
	bodyWaterInterval       = 240  // body converts 1 clean water → waste every 4 sec
	hungerInterval          = 480  // hunger rises 1 every 8 sec (~13 min 0→100, ~16 game hours)
	thirstInterval          = 300  // thirst rises 1 every 5 sec (~8 min 0→100, ~10 game hours)
	hygieneInterval         = 720  // hygiene worsens 1 every 12 sec (~20 min 0→100, ~24 game hours)
	warningInterval         = 900  // check warnings every 15 sec
)

// Sim is the game simulation. It owns all gameplay state.
type Sim struct {
	ECS       *ecs.World
	Grid      *world.TileGrid
	Layout    *world.ShipLayout
	Resources Resources
	Needs     PlayerNeeds
	Log       *MessageLog
	Ticks     uint64
	Sector    *Sector

	// Equipment on/off state
	EngineOn    bool
	GeneratorOn bool
	RecyclerOn  bool

	// UI signals (set by Interact, cleared by Game after handling)
	NavActivated   bool
	PilotActivated bool

	player ecs.Entity
	posMap *ecs.Map[Position]
}

// NewSim creates a simulation from a ship layout.
func NewSim(layout *world.ShipLayout) *Sim {
	w := ecs.NewWorld(256)
	grid := layout.ToTileGrid()

	posMap := ecs.NewMap[Position](w)

	player := ecs.NewMap2[Position, PlayerControlled](w).NewEntity(
		&Position{X: layout.SpawnX(), Y: layout.SpawnY()},
		&PlayerControlled{},
	)

	log := NewMessageLog(50)
	log.Add("Waking from cryo aboard the Nomad.", MsgInfo)
	log.Add("All systems online. Generator, recycler running.", MsgInfo)
	log.Add("You should probably find the toilet.", MsgWarning)

	return &Sim{
		ECS:             w,
		Grid:            grid,
		Layout:          layout,
		Resources:       NewShuttleResources(),
		Needs:           PlayerNeeds{Hunger: 40, Thirst: 30, Hygiene: 20},
		Log:             log,
		Sector:          NewSector(42),
		EngineOn:    true,
		GeneratorOn: true,
		RecyclerOn:  true,
		player:          player,
		posMap:          posMap,
	}
}

// PlayerPos returns the player's current tile coordinates.
func (s *Sim) PlayerPos() (int, int) {
	pos := s.posMap.Get(s.player)
	return pos.X, pos.Y
}

// TryMovePlayer attempts to move the player by (dx, dy).
func (s *Sim) TryMovePlayer(dx, dy int) bool {
	pos := s.posMap.Get(s.player)
	newX := pos.X + dx
	newY := pos.Y + dy
	if s.Grid.IsWalkable(newX, newY) {
		pos.X = newX
		pos.Y = newY
		return true
	}
	return false
}

// Tick advances the simulation by one step.
func (s *Sim) Tick() {
	s.Ticks++
	s.tickGenerator()
	s.tickRecycler()
	s.tickBody()
	s.tickNeeds()
	s.tickSystemMapNPCs()
	s.tickSystemMapShuttle()
	if s.Ticks%warningInterval == 0 {
		s.checkWarnings()
	}
}

func (s *Sim) tickGenerator() {
	if !s.GeneratorOn {
		return
	}
	if s.Ticks%generatorInterval == 0 {
		if s.Resources.Energy < s.Resources.MaxEnergy {
			s.Resources.Energy++
		}
	}
}

func (s *Sim) tickRecycler() {
	if !s.RecyclerOn {
		return
	}
	r := &s.Resources
	rc := &r.Recycler

	// Intake phase: pull dirty matter from ship pools into recycler buffer
	if s.Ticks%recyclerIntakeInterval == 0 {
		if r.Water.Dirty > 0 && rc.WaterBuffer < rc.Capacity {
			r.Water.Dirty--
			rc.WaterBuffer++
		}
		if r.Organic.Dirty > 0 && rc.OrganicBuffer < rc.Capacity {
			r.Organic.Dirty--
			rc.OrganicBuffer++
		}
	}

	// Process phase: convert buffered dirty → clean, costs 1 energy each
	if s.Ticks%recyclerProcessInterval == 0 {
		if rc.WaterBuffer > 0 && r.Energy > 0 {
			rc.WaterBuffer--
			r.Water.Clean++
			r.Energy--
		}
		if rc.OrganicBuffer > 0 && r.Energy > 0 {
			rc.OrganicBuffer--
			r.Organic.Clean++
			r.Energy--
		}
	}
}

// tickBody processes matter inside the player's body.
// Clean matter (food/water) slowly becomes dirty waste.
func (s *Sim) tickBody() {
	r := &s.Resources

	// Digest food: body organic → waste organic
	if s.Ticks%bodyDigestInterval == 0 && r.BodyOrganic > 0 {
		r.BodyOrganic--
		r.WasteOrganic++
	}

	// Process water: body water → waste water
	if s.Ticks%bodyWaterInterval == 0 && r.BodyWater > 0 {
		r.BodyWater--
		r.WasteWater++
	}
}

func (s *Sim) tickNeeds() {
	n := &s.Needs

	// Hunger rises over time
	if s.Ticks%hungerInterval == 0 {
		n.Hunger = min(n.Hunger+1, 100)
	}

	// Thirst rises over time
	if s.Ticks%thirstInterval == 0 {
		n.Thirst = min(n.Thirst+1, 100)
	}

	// Hygiene degrades over time
	if s.Ticks%hygieneInterval == 0 {
		n.Hygiene = min(n.Hygiene+1, 100)
	}
}

func (s *Sim) tickSystemMapNPCs() {
	sm := s.Sector.Systems[s.Sector.CurrentSystem].Map
	if sm != nil {
		sm.TickNPCs()
	}
}

func (s *Sim) tickSystemMapShuttle() {
	sm := s.Sector.Systems[s.Sector.CurrentSystem].Map
	if sm != nil {
		sm.Shuttle.Tick()
		sm.Shuttle.ClampToBounds(sm.Width, sm.Height)
	}
}

func (s *Sim) checkWarnings() {
	r := &s.Resources
	n := &s.Needs

	if r.Energy <= 10 && r.Energy > 0 {
		s.Log.Add(fmt.Sprintf("Power low: %d. Consider turning off equipment.", r.Energy), MsgWarning)
	} else if r.Energy == 0 {
		s.Log.Add("POWER DEPLETED. Recyclers offline.", MsgCritical)
	}

	if r.Water.Clean <= 15 && r.Water.Clean > 0 {
		s.Log.Add(fmt.Sprintf("Clean water low: %d.", r.Water.Clean), MsgWarning)
	} else if r.Water.Clean == 0 {
		s.Log.Add("NO CLEAN WATER. Dehydration imminent.", MsgCritical)
	}

	if r.Organic.Clean <= 15 && r.Organic.Clean > 0 {
		s.Log.Add(fmt.Sprintf("Clean organics low: %d.", r.Organic.Clean), MsgWarning)
	} else if r.Organic.Clean == 0 {
		s.Log.Add("NO CLEAN ORGANICS. Starvation imminent.", MsgCritical)
	}

	if r.TotalWaste() >= 20 {
		s.Log.Add("You really need to find the toilet.", MsgWarning)
	}

	if n.Hunger >= 80 {
		s.Log.Add("You're starving. Find the food station.", MsgCritical)
	} else if n.Hunger >= 60 {
		s.Log.Add("Getting hungry.", MsgWarning)
	}

	if n.Thirst >= 80 {
		s.Log.Add("Severely dehydrated!", MsgCritical)
	} else if n.Thirst >= 60 {
		s.Log.Add("Getting thirsty.", MsgWarning)
	}

	if n.Hygiene >= 80 {
		s.Log.Add("You reek. Consider a shower.", MsgWarning)
	}
}

// Interact handles the player pressing E on the tile they're standing on.
func (s *Sim) Interact() {
	px, py := s.PlayerPos()
	tile := s.Grid.Get(px, py)
	r := &s.Resources

	switch tile.Equipment {
	case world.EquipFoodStation:
		// Eat: clean organic leaves ship → enters body, hunger drops
		if r.Organic.Clean < 5 {
			s.Log.Add("Not enough clean organics to dispense a meal.", MsgWarning)
			return
		}
		if r.BodyFullness()+5 > MaxBodyFullness {
			s.Log.Add("Too full to eat. Use the toilet first.", MsgWarning)
			return
		}
		r.Organic.Clean -= 5
		r.BodyOrganic += 5
		s.Needs.Hunger = max(s.Needs.Hunger-35, 0)
		s.Log.Add("Dispensed a meal. Tastes like... Tuesday.", MsgInfo)

	case world.EquipDrinkStation:
		// Drink: clean water leaves ship → enters body, thirst drops
		if r.Water.Clean < 3 {
			s.Log.Add(fmt.Sprintf("Not enough clean water. %dc %dd.", r.Water.Clean, r.Water.Dirty), MsgWarning)
			return
		}
		if r.BodyFullness()+3 > MaxBodyFullness {
			s.Log.Add("Too full to drink. Use the toilet first.", MsgWarning)
			return
		}
		r.Water.Clean -= 3
		r.BodyWater += 3
		s.Needs.Thirst = max(s.Needs.Thirst-25, 0)
		s.Log.Add(fmt.Sprintf("Gulped some recycled water. %dc remaining.", r.Water.Clean), MsgInfo)

	case world.EquipToilet:
		// Flush: waste from body → dirty matter back to ship pools
		waste := r.TotalWaste()
		if waste == 0 {
			s.Log.Add("Nothing to deposit. Efficient.", MsgSocial)
			return
		}
		s.Log.Add(fmt.Sprintf("Waste flushed. +%d dirty organic, +%d dirty water back in system.",
			r.WasteOrganic, r.WasteWater), MsgInfo)
		r.Organic.Dirty += r.WasteOrganic
		r.Water.Dirty += r.WasteWater
		r.WasteOrganic = 0
		r.WasteWater = 0

	case world.EquipShower:
		// Uses clean water → dirty water (external, doesn't go through body)
		if r.Water.Clean < 3 {
			s.Log.Add("Not enough clean water for a shower.", MsgWarning)
			return
		}
		r.Water.Clean -= 3
		r.Water.Dirty += 3
		s.Needs.Hygiene = max(s.Needs.Hygiene-40, 0)
		s.Log.Add("Quick shower. Refreshing.", MsgInfo)

	case world.EquipBed:
		s.Log.Add("You rest briefly. The void doesn't care.", MsgSocial)

	case world.EquipLocker:
		s.Log.Add("Storage locker. Empty for now.", MsgSocial)

	case world.EquipNavConsole:
		s.NavActivated = true
		s.Log.Add("Navigation console activated.", MsgInfo)

	case world.EquipPilotConsole:
		s.PilotActivated = true
		s.Log.Add("Pilot station. Launching system view.", MsgInfo)

	case world.EquipScienceConsole:
		s.Log.Add("Science station. Deborah usually sits here. She has no idea what she's doing.", MsgSocial)

	case world.EquipIncinerator:
		s.Log.Add("Incinerator. Not yet operational.", MsgInfo)

	case world.EquipMedical:
		s.Log.Add("Medical station. Not yet operational.", MsgInfo)

	case world.EquipEngine:
		status := "OFF"
		if s.EngineOn {
			status = "ON"
		}
		s.Log.Add(fmt.Sprintf("Engine [%s]. Provides thrust.", status), MsgInfo)

	case world.EquipGenerator:
		status := "OFF"
		if s.GeneratorOn {
			status = "ON"
		}
		s.Log.Add(fmt.Sprintf("Generator [%s]. Produces 1 energy/sec.", status), MsgInfo)

	case world.EquipPowerCell:
		s.Log.Add(fmt.Sprintf("Power cell: %d / %d.", r.Energy, r.MaxEnergy), MsgInfo)

	case world.EquipOrganicTank:
		s.Log.Add(fmt.Sprintf("Organic tank: %dc %dd. Digesting: %d, waste: %d.",
			r.Organic.Clean, r.Organic.Dirty, r.BodyOrganic, r.WasteOrganic), MsgInfo)

	case world.EquipWaterTank:
		s.Log.Add(fmt.Sprintf("Water tank: %dc %dd. Body: %d, waste: %d.",
			r.Water.Clean, r.Water.Dirty, r.BodyWater, r.WasteWater), MsgInfo)

	case world.EquipMatterRecycler:
		status := "OFF"
		if s.RecyclerOn {
			status = "ON"
		}
		rc := &r.Recycler
		s.Log.Add(fmt.Sprintf("Recycler [%s]. Buffer: %dw %do / %d cap.",
			status, rc.WaterBuffer, rc.OrganicBuffer, rc.Capacity), MsgInfo)

	case world.EquipViewscreen:
		star := s.Sector.Systems[s.Sector.CurrentSystem]
		s.Log.Add(fmt.Sprintf("Viewscreen: %s system. %s.", star.Name, StarTypeName(star.Type)), MsgInfo)

	case world.EquipCargoConsole:
		s.Log.Add("Cargo console. Jettison not yet available.", MsgInfo)

	case world.EquipCargoTile:
		s.Log.Add("Cargo pad. Empty.", MsgInfo)

	default:
		s.Log.Add("Nothing to interact with here.", MsgSocial)
	}
}

// ToggleEquipment handles the player pressing T on equipment to toggle it on/off.
func (s *Sim) ToggleEquipment() {
	px, py := s.PlayerPos()
	tile := s.Grid.Get(px, py)

	switch tile.Equipment {
	case world.EquipEngine:
		s.EngineOn = !s.EngineOn
		if s.EngineOn {
			s.Log.Add("Engine started.", MsgInfo)
		} else {
			s.Log.Add("Engine shut down.", MsgWarning)
		}

	case world.EquipGenerator:
		s.GeneratorOn = !s.GeneratorOn
		if s.GeneratorOn {
			s.Log.Add("Generator online. Producing energy.", MsgInfo)
		} else {
			s.Log.Add("Generator offline. No power generation.", MsgWarning)
		}

	case world.EquipMatterRecycler:
		s.RecyclerOn = !s.RecyclerOn
		if s.RecyclerOn {
			s.Log.Add("Recycler online. Processing dirty matter.", MsgInfo)
		} else {
			s.Log.Add("Recycler offline. Saving power.", MsgInfo)
		}

	default:
		s.Log.Add("This equipment can't be toggled.", MsgSocial)
	}
}

// NavigateTo attempts to jump the shuttle to the target star system.
func (s *Sim) NavigateTo(targetIdx int) bool {
	cost := s.Sector.EnergyCostTo(targetIdx)
	if s.Resources.Energy < cost {
		s.Log.Add(fmt.Sprintf("Not enough energy. Need %d, have %d.", cost, s.Resources.Energy), MsgWarning)
		return false
	}
	s.Resources.Energy -= cost
	s.Sector.CurrentSystem = targetIdx
	s.Sector.Systems[targetIdx].Visited = true
	s.Sector.EnsureSystemMap(targetIdx)
	star := s.Sector.Systems[targetIdx]
	s.Log.Add(fmt.Sprintf("Arrived at %s. %s. Energy: -%d.", star.Name, StarTypeName(star.Type), cost), MsgDiscovery)
	return true
}
