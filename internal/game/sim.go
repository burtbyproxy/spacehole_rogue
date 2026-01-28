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

	// Skills and discovery
	Skills    PlayerSkills
	Discovery *DiscoveryLog

	// UI signals (set by Interact, cleared by Game after handling)
	NavActivated    bool
	PilotActivated  bool
	DockActivated   bool // set when player docks at a station
	CargoActivated  bool // set when player uses cargo console
	ScanActivated   bool // set when player uses science console

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

	sector := NewSector(42)
	disc := NewDiscoveryLog()
	// Mark starting system and star type as discovered
	disc.SystemsVisited[0] = true
	disc.TotalSystemsVisited = 1
	disc.StarTypesSeen[int(sector.Systems[0].Type)] = true
	disc.TotalStarTypesSeen = 1

	return &Sim{
		ECS:         w,
		Grid:        grid,
		Layout:      layout,
		Resources:   NewShuttleResources(),
		Needs:       PlayerNeeds{Hunger: 40, Thirst: 30, Hygiene: 20},
		Log:         log,
		Sector:      sector,
		Discovery:   disc,
		EngineOn:    true,
		GeneratorOn: true,
		RecyclerOn:  true,
		player:      player,
		posMap:      posMap,
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
		if s.Skills.AddXP(SkillSurvival, 2.0) {
			LogLevelUp(s.Log, SkillSurvival, s.Skills.Level(SkillSurvival))
		}

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
		if s.Skills.AddXP(SkillSurvival, 1.5) {
			LogLevelUp(s.Log, SkillSurvival, s.Skills.Level(SkillSurvival))
		}

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
		if s.Skills.AddXP(SkillSurvival, 0.5) {
			LogLevelUp(s.Log, SkillSurvival, s.Skills.Level(SkillSurvival))
		}

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
		if s.Skills.AddXP(SkillSurvival, 0.5) {
			LogLevelUp(s.Log, SkillSurvival, s.Skills.Level(SkillSurvival))
		}

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
		s.ScanActivated = true
		s.Log.Add("Science station active. Deborah left her notes. They're just hoof prints.", MsgSocial)

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
		s.Skills.AddXP(SkillEngineering, 0.5)

	case world.EquipGenerator:
		status := "OFF"
		if s.GeneratorOn {
			status = "ON"
		}
		s.Log.Add(fmt.Sprintf("Generator [%s]. Produces 1 energy/sec.", status), MsgInfo)
		s.Skills.AddXP(SkillEngineering, 0.5)

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
		s.Skills.AddXP(SkillEngineering, 0.5)

	case world.EquipViewscreen:
		star := s.Sector.Systems[s.Sector.CurrentSystem]
		s.Log.Add(fmt.Sprintf("Viewscreen: %s system. %s.", star.Name, StarTypeName(star.Type)), MsgInfo)

	case world.EquipCargoConsole:
		s.CargoActivated = true
		s.Log.Add("Cargo console activated.", MsgInfo)

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
		if s.Skills.AddXP(SkillEngineering, 3.0) {
			LogLevelUp(s.Log, SkillEngineering, s.Skills.Level(SkillEngineering))
		}

	case world.EquipGenerator:
		s.GeneratorOn = !s.GeneratorOn
		if s.GeneratorOn {
			s.Log.Add("Generator online. Producing energy.", MsgInfo)
		} else {
			s.Log.Add("Generator offline. No power generation.", MsgWarning)
		}
		if s.Skills.AddXP(SkillEngineering, 3.0) {
			LogLevelUp(s.Log, SkillEngineering, s.Skills.Level(SkillEngineering))
		}

	case world.EquipMatterRecycler:
		s.RecyclerOn = !s.RecyclerOn
		if s.RecyclerOn {
			s.Log.Add("Recycler online. Processing dirty matter.", MsgInfo)
		} else {
			s.Log.Add("Recycler offline. Saving power.", MsgInfo)
		}
		if s.Skills.AddXP(SkillEngineering, 3.0) {
			LogLevelUp(s.Log, SkillEngineering, s.Skills.Level(SkillEngineering))
		}

	default:
		s.Log.Add("This equipment can't be toggled.", MsgSocial)
	}
}

// DockAtStation docks at the current system's station.
// Performs auto-refill and returns the station data.
func (s *Sim) DockAtStation() *StationData {
	sm := s.Sector.CurrentSystemMap()
	sd := sm.EnsureStationData()
	if sd == nil {
		return nil
	}
	DockRefill(&s.Resources)
	s.Log.Add(fmt.Sprintf("Docked at %s. Tanks topped off, energy full.", sd.Name), MsgInfo)
	s.OnStationDocked(s.Sector.CurrentSystem)
	return sd
}

// RepairHull repairs hull points at 2 credits per point.
// amount = 0 means full repair. Returns credits spent and points repaired.
func (s *Sim) RepairHull(amount int) (cost int, repaired int) {
	const costPerPoint = 2
	damage := s.Resources.MaxHull - s.Resources.Hull
	if damage == 0 {
		s.Log.Add("Hull integrity at 100%. No repairs needed.", MsgInfo)
		return 0, 0
	}
	if amount <= 0 || amount > damage {
		amount = damage
	}
	maxAfford := s.Resources.Credits / costPerPoint
	if amount > maxAfford {
		amount = maxAfford
	}
	if amount == 0 {
		s.Log.Add("Not enough credits for repairs.", MsgWarning)
		return 0, 0
	}
	cost = amount * costPerPoint
	s.Resources.Credits -= cost
	s.Resources.Hull += amount
	s.Log.Add(fmt.Sprintf("Repaired %d hull pts for %dcr. Hull: %d/%d.",
		amount, cost, s.Resources.Hull, s.Resources.MaxHull), MsgInfo)
	return cost, amount
}

// BuyCargo buys one unit of cargo from the station.
func (s *Sim) BuyCargo(sd *StationData, kind CargoKind) bool {
	if !sd.Stocked[kind] || sd.Stock[kind] <= 0 {
		s.Log.Add("Station has none of that in stock.", MsgWarning)
		return false
	}
	price := sd.SellPrices[kind]
	if s.Resources.Credits < price {
		s.Log.Add(fmt.Sprintf("Need %dcr. You have %dcr.", price, s.Resources.Credits), MsgWarning)
		return false
	}
	added := s.Resources.AddCargo(kind, 1)
	if added == 0 {
		s.Log.Add("Cargo bay full. No empty pads or stacks.", MsgWarning)
		return false
	}
	s.Resources.Credits -= price
	sd.Stock[kind]--
	s.Log.Add(fmt.Sprintf("Bought %s for %dcr.", CargoName(kind), price), MsgInfo)
	if s.Skills.AddXP(SkillDiplomacy, 2.0) {
		LogLevelUp(s.Log, SkillDiplomacy, s.Skills.Level(SkillDiplomacy))
	}
	return true
}

// SellCargo sells one unit from the given pad index.
func (s *Sim) SellCargo(sd *StationData, padIdx int) bool {
	if padIdx < 0 || padIdx >= len(s.Resources.CargoPads) {
		return false
	}
	pad := &s.Resources.CargoPads[padIdx]
	if pad.Kind == CargoNone || pad.Count <= 0 {
		s.Log.Add("That pad is empty.", MsgWarning)
		return false
	}
	kind := pad.Kind
	price := sd.BuyPrices[kind]
	if price <= 0 {
		price = 1 // station always buys for at least 1
	}
	s.Resources.Credits += price
	sd.Stock[kind]++
	pad.Count--
	name := CargoName(kind)
	if pad.Count == 0 {
		pad.Kind = CargoNone
	}
	s.Log.Add(fmt.Sprintf("Sold %s for %dcr.", name, price), MsgInfo)
	if s.Skills.AddXP(SkillDiplomacy, 2.0) {
		LogLevelUp(s.Log, SkillDiplomacy, s.Skills.Level(SkillDiplomacy))
	}
	return true
}

// JettisonCargo destroys one unit from the given pad index.
func (s *Sim) JettisonCargo(padIdx int) bool {
	if padIdx < 0 || padIdx >= len(s.Resources.CargoPads) {
		return false
	}
	pad := &s.Resources.CargoPads[padIdx]
	if pad.Kind == CargoNone || pad.Count <= 0 {
		s.Log.Add("That pad is empty.", MsgWarning)
		return false
	}
	name := CargoName(pad.Kind)
	pad.Count--
	if pad.Count == 0 {
		pad.Kind = CargoNone
	}
	s.Log.Add(fmt.Sprintf("Jettisoned 1x %s into the void.", name), MsgWarning)
	return true
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
	if s.Skills.AddXP(SkillPiloting, 5.0) {
		LogLevelUp(s.Log, SkillPiloting, s.Skills.Level(SkillPiloting))
	}
	s.OnSystemVisited(targetIdx)
	return true
}

// OnSystemVisited handles first-visit discovery bonuses for a star system.
func (s *Sim) OnSystemVisited(sysIdx int) {
	star := s.Sector.Systems[sysIdx]

	// First time visiting this specific system?
	if !s.Discovery.SystemsVisited[sysIdx] {
		s.Discovery.SystemsVisited[sysIdx] = true
		s.Discovery.TotalSystemsVisited++
		s.Resources.Credits += 10
		s.Log.Add(fmt.Sprintf("New system discovered: %s! +10cr.", star.Name), MsgDiscovery)
		if s.Skills.AddXP(SkillPiloting, 8.0) {
			LogLevelUp(s.Log, SkillPiloting, s.Skills.Level(SkillPiloting))
		}
	}

	// First time seeing this star type?
	if !s.Discovery.StarTypesSeen[int(star.Type)] {
		s.Discovery.StarTypesSeen[int(star.Type)] = true
		s.Discovery.TotalStarTypesSeen++
		s.Resources.Credits += 25
		s.Log.Add(fmt.Sprintf("New star type logged: %s! +25cr.", StarTypeName(star.Type)), MsgDiscovery)
		if s.Skills.AddXP(SkillScience, 15.0) {
			LogLevelUp(s.Log, SkillScience, s.Skills.Level(SkillScience))
		}
	}
}

// ScanPlanet scans a planet in the current system map.
func (s *Sim) ScanPlanet(objIdx int) {
	sm := s.Sector.CurrentSystemMap()
	obj := &sm.Objects[objIdx]
	sysIdx := s.Sector.CurrentSystem
	key := ScanKey(sysIdx, objIdx)

	if _, already := s.Discovery.PlanetsScanned[key]; already {
		s.Log.Add(fmt.Sprintf("Re-scanning %s. No new data.", obj.Name), MsgInfo)
		s.Skills.AddXP(SkillScience, 1.0)
		return
	}

	systemName := s.Sector.Systems[sysIdx].Name
	scanData := GenerateScanData(s.Sector.Seed, sysIdx, objIdx, obj, systemName)
	s.Discovery.PlanetsScanned[key] = scanData
	s.Discovery.TotalScans++

	// Keep recent scans list (newest first, max 10)
	s.Discovery.RecentScans = append([]PlanetScanData{scanData}, s.Discovery.RecentScans...)
	if len(s.Discovery.RecentScans) > 10 {
		s.Discovery.RecentScans = s.Discovery.RecentScans[:10]
	}

	s.Resources.Credits += 15
	s.Log.Add(fmt.Sprintf("Scanned %s (%s). %s. +15cr.",
		obj.Name, PlanetKindName(obj.PlanetType), scanData.Resources), MsgDiscovery)
	if scanData.Hazard != "" {
		s.Log.Add(fmt.Sprintf("Hazard: %s", scanData.Hazard), MsgWarning)
	}
	if scanData.POI != "" {
		s.Log.Add(fmt.Sprintf("POI: %s", scanData.POI), MsgDiscovery)
	}
	if s.Skills.AddXP(SkillScience, 8.0) {
		LogLevelUp(s.Log, SkillScience, s.Skills.Level(SkillScience))
	}
}

// OnStationDocked handles first-dock discovery bonuses.
func (s *Sim) OnStationDocked(sysIdx int) {
	if !s.Discovery.StationsDocked[sysIdx] {
		s.Discovery.StationsDocked[sysIdx] = true
		s.Discovery.TotalStationsDocked++
		s.Resources.Credits += 10
		s.Log.Add("First dock at this station! +10cr.", MsgDiscovery)
	}
	if s.Skills.AddXP(SkillDiplomacy, 3.0) {
		LogLevelUp(s.Log, SkillDiplomacy, s.Skills.Level(SkillDiplomacy))
	}
}
