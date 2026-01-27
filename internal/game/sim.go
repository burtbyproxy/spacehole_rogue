package game

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"
	"github.com/spacehole-rogue/spacehole_rogue/internal/world"
)

// Tick intervals (at 60 TPS)
const (
	engineGenInterval       = 60  // engine generates 1 energy per second
	waterRecyclerInterval   = 100 // converts 1 dirty water → clean, costs 1 energy
	organicRecyclerInterval = 100 // converts 1 dirty organic → clean, costs 1 energy
	bodyDigestInterval      = 180 // body converts 1 clean organic → waste every 3 sec
	bodyWaterInterval       = 120 // body converts 1 clean water → waste every 2 sec
	hungerInterval          = 180 // hunger rises 1 every 3 sec
	thirstInterval          = 120 // thirst rises 1 every 2 sec
	hygieneInterval         = 600 // hygiene worsens 1 every 10 sec
	warningInterval         = 300 // check warnings every 5 sec
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

	// Equipment on/off state
	EngineOn        bool
	WaterRecyclerOn bool
	ReplicatorOn    bool // acts as organic recycler when ON

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
	log.Add("All systems online. Engine, recyclers running.", MsgInfo)
	log.Add("You should probably find the toilet.", MsgWarning)

	return &Sim{
		ECS:             w,
		Grid:            grid,
		Layout:          layout,
		Resources:       NewShuttleResources(),
		Needs:           PlayerNeeds{Hunger: 40, Thirst: 30, Hygiene: 20},
		Log:             log,
		EngineOn:        true,
		WaterRecyclerOn: true,
		ReplicatorOn:    true,
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
	s.tickEngine()
	s.tickRecyclers()
	s.tickBody()
	s.tickNeeds()
	if s.Ticks%warningInterval == 0 {
		s.checkWarnings()
	}
}

func (s *Sim) tickEngine() {
	if !s.EngineOn {
		return
	}
	if s.Ticks%engineGenInterval == 0 {
		if s.Resources.Energy < s.Resources.MaxEnergy {
			s.Resources.Energy++
		}
	}
}

func (s *Sim) tickRecyclers() {
	r := &s.Resources

	// Water recycler: dirty water → clean water, costs 1 energy
	if s.WaterRecyclerOn && s.Ticks%waterRecyclerInterval == 0 {
		if r.Water.Dirty > 0 && r.Energy > 0 {
			r.Water.Dirty--
			r.Water.Clean++
			r.Energy--
		}
	}

	// Replicator as organic recycler: dirty organic → clean organic, costs 1 energy
	if s.ReplicatorOn && s.Ticks%organicRecyclerInterval == 0 {
		if r.Organic.Dirty > 0 && r.Energy > 0 {
			r.Organic.Dirty--
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
		s.Log.Add("You're starving. Find the replicator.", MsgCritical)
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
	case world.EquipReplicator:
		// Eat: clean organic leaves ship → enters body, hunger drops
		if r.Organic.Clean < 5 {
			s.Log.Add("Not enough clean organics to replicate a meal.", MsgWarning)
			return
		}
		if r.BodyFullness()+5 > MaxBodyFullness {
			s.Log.Add("Too full to eat. Use the toilet first.", MsgWarning)
			return
		}
		r.Organic.Clean -= 5
		r.BodyOrganic += 5
		s.Needs.Hunger = max(s.Needs.Hunger-35, 0)
		s.Log.Add("Replicated a meal. Tastes like... Tuesday.", MsgInfo)

	case world.EquipWaterTank:
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

	case world.EquipConsole:
		s.Log.Add("Navigation console online. No destinations yet.", MsgInfo)

	case world.EquipEngine:
		status := "OFF"
		if s.EngineOn {
			status = "ON"
		}
		s.Log.Add(fmt.Sprintf("Engine [%s]. Generates 1 energy/sec.", status), MsgInfo)

	case world.EquipPowerCell:
		s.Log.Add(fmt.Sprintf("Power cell: %d / %d.", r.Energy, r.MaxEnergy), MsgInfo)

	case world.EquipFoodStore:
		s.Log.Add(fmt.Sprintf("Organics: %dc %dd. Digesting: %d, waste: %d.",
			r.Organic.Clean, r.Organic.Dirty, r.BodyOrganic, r.WasteOrganic), MsgInfo)

	case world.EquipWaterRecycler:
		status := "OFF"
		if s.WaterRecyclerOn {
			status = "ON"
		}
		s.Log.Add(fmt.Sprintf("Water recycler [%s]. dirty→clean, costs energy.", status), MsgInfo)

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
			s.Log.Add("Engine shut down. No power generation.", MsgWarning)
		}

	case world.EquipWaterRecycler:
		s.WaterRecyclerOn = !s.WaterRecyclerOn
		if s.WaterRecyclerOn {
			s.Log.Add("Water recycler online.", MsgInfo)
		} else {
			s.Log.Add("Water recycler offline. Saving power.", MsgInfo)
		}

	case world.EquipReplicator:
		s.ReplicatorOn = !s.ReplicatorOn
		if s.ReplicatorOn {
			s.Log.Add("Replicator online. Processing dirty organics.", MsgInfo)
		} else {
			s.Log.Add("Replicator offline. Saving power.", MsgInfo)
		}

	default:
		s.Log.Add("This equipment can't be toggled.", MsgSocial)
	}
}
