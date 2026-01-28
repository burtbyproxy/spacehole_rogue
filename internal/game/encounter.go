package game

import (
	"fmt"
	"math/rand/v2"
)

// EncounterKind identifies the type of NPC encounter.
type EncounterKind uint8

const (
	EncounterTrader EncounterKind = iota
	EncounterPatrol
	EncounterPirate
)

// EncounterState tracks an active encounter with an NPC ship.
type EncounterState struct {
	Kind       EncounterKind
	ShipName   string
	AIKind     ShipAIKind
	ShipObj    *SpaceObject
	Greeting   string
	Options    []EncounterOption
	ResultText string // filled after player picks an option
	Resolved   bool
}

// EncounterOption is a single choice available in an encounter menu.
type EncounterOption struct {
	Label       string
	Enabled     bool
	DisableText string // shown when option is disabled
	SkillReq    SkillID
	SkillLevel  int // min level required (0 = no requirement)
}

// HailState tracks a pending incoming hail from an NPC ship.
type HailState struct {
	Ship      *SpaceObject
	TicksLeft int // countdown to hail expiry
}

const hailTimeout = 300 // ticks before hail expires

// Ship name pools (unique proper names, assigned during system generation).
var traderShipNames = []string{
	"Star Hauler", "Merchantman", "Cargo Queen", "Lucky Profit", "Silk Road",
}

var patrolShipNames = []string{
	"Sentinel VII", "Watchdog", "Iron Law", "Peacekeeper", "Blue Line",
}

var pirateShipNames = []string{
	"Void Fang", "Black Marlin", "Skull Dancer", "Dread Nail", "Gut Ripper",
}

// ShipProperName returns a proper name for a ship based on its AI type and a seed index.
func ShipProperName(kind ShipAIKind, idx int) string {
	switch kind {
	case AITrader:
		return traderShipNames[idx%len(traderShipNames)]
	case AIPatrol:
		return patrolShipNames[idx%len(patrolShipNames)]
	case AIPirate:
		return pirateShipNames[idx%len(pirateShipNames)]
	default:
		return "Unknown"
	}
}

// Greeting pools per encounter kind.
var traderGreetings = []string{
	"Greetings, traveler! Looking to trade?",
	"Well met, pilot. I've got goods if you've got credits.",
	"Ahoy there! Business is slow out here. Care to browse?",
}

var patrolGreetings = []string{
	"Unidentified shuttle, transmit credentials.",
	"This is sector patrol. State your business.",
	"Routine inspection. Please hold position.",
}

var pirateGreetings = []string{
	"Cut your engines. Hand over your cargo.",
	"Nice shuttle. It's ours now. Unless you've got something better to offer.",
	"Resistance is expensive. Surrender is free.",
}

// NewEncounter creates an encounter from a hailed NPC ship.
func NewEncounter(ship *SpaceObject, sectorSeed int64, skills *PlayerSkills) *EncounterState {
	seed := sectorSeed*777 + int64(ship.X)*31 + int64(ship.Y)*17
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>8|3)))

	kind := encounterKindFromAI(ship.AIKind)

	// Pick greeting
	var greeting string
	switch kind {
	case EncounterTrader:
		greeting = traderGreetings[rng.IntN(len(traderGreetings))]
	case EncounterPatrol:
		greeting = patrolGreetings[rng.IntN(len(patrolGreetings))]
	case EncounterPirate:
		greeting = pirateGreetings[rng.IntN(len(pirateGreetings))]
	}

	enc := &EncounterState{
		Kind:     kind,
		ShipName: ship.Name,
		AIKind:   ship.AIKind,
		ShipObj:  ship,
		Greeting: greeting,
	}

	// Build options based on encounter type
	switch kind {
	case EncounterTrader:
		enc.Options = []EncounterOption{
			{Label: "Hail back (friendly chat)", Enabled: true},
			{Label: "Trade goods", Enabled: true},
			{Label: "Request supplies", Enabled: true},
			{Label: "Ignore transmission", Enabled: true},
		}
	case EncounterPatrol:
		enc.Options = []EncounterOption{
			{Label: "Identify yourself", Enabled: true},
			{Label: "Report pirate activity", Enabled: true},
			{Label: "Request escort", Enabled: true},
			{Label: "Ignore transmission", Enabled: true},
		}
	case EncounterPirate:
		bluffEnabled := skills.Level(SkillDiplomacy) >= 3
		enc.Options = []EncounterOption{
			{Label: "Surrender cargo", Enabled: true},
			{Label: "Bribe (30cr)", Enabled: true},
			{Label: "Bluff (Diplomacy Lv 3+)", Enabled: bluffEnabled, DisableText: "Diplomacy too low", SkillReq: SkillDiplomacy, SkillLevel: 3},
			{Label: "Flee", Enabled: true},
			{Label: "Fight", Enabled: false, DisableText: "Combat systems offline"},
		}
	}

	return enc
}

func encounterKindFromAI(ai ShipAIKind) EncounterKind {
	switch ai {
	case AITrader:
		return EncounterTrader
	case AIPatrol:
		return EncounterPatrol
	case AIPirate:
		return EncounterPirate
	default:
		return EncounterTrader
	}
}

// ResolveEncounter handles the player's choice in an encounter.
// Returns result text to display.
func ResolveEncounter(sim *Sim, enc *EncounterState, optionIdx int) string {
	if optionIdx < 0 || optionIdx >= len(enc.Options) {
		return ""
	}
	opt := enc.Options[optionIdx]
	if !opt.Enabled {
		return opt.DisableText
	}

	enc.Resolved = true

	switch enc.Kind {
	case EncounterTrader:
		return resolveTrader(sim, enc, optionIdx)
	case EncounterPatrol:
		return resolvePatrol(sim, enc, optionIdx)
	case EncounterPirate:
		return resolvePirate(sim, enc, optionIdx)
	}
	return "Transmission ended."
}

func resolveTrader(sim *Sim, enc *EncounterState, idx int) string {
	switch idx {
	case 0: // Hail back
		if sim.Skills.AddXP(SkillDiplomacy, 1.0) {
			LogLevelUp(sim.Log, SkillDiplomacy, sim.Skills.Level(SkillDiplomacy))
		}
		responses := []string{
			"\"Safe travels, friend. The void is kinder to those who talk first.\"",
			"\"Always nice to meet a friendly face out here. Most just shoot.\"",
			"\"May your cargo hold stay full and your hull stay intact!\"",
		}
		seed := sim.Ticks
		return responses[seed%uint64(len(responses))]

	case 1: // Trade goods
		// Signal to open trade view - handled by main.go
		sim.Log.Add("Opening trade channel with merchant vessel.", MsgInfo)
		return "TRADE" // special signal

	case 2: // Request supplies
		if sim.Skills.AddXP(SkillDiplomacy, 2.0) {
			LogLevelUp(sim.Log, SkillDiplomacy, sim.Skills.Level(SkillDiplomacy))
		}
		// 40% chance of success
		seed := sim.Sector.Seed*333 + int64(enc.ShipObj.X) + int64(sim.Ticks)
		rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>4|5)))
		if rng.IntN(5) < 2 {
			// Success — give some water and food
			sim.Resources.Water.Clean += 5
			sim.Resources.Organic.Clean += 3
			return "\"Here, take these. We've got plenty.\"\n+5 clean water, +3 clean organics."
		}
		return "\"Sorry, friend. We're running lean ourselves. Can't spare any.\""

	case 3: // Ignore
		return "Transmission closed. The merchant vessel drifts on."
	}
	return ""
}

func resolvePatrol(sim *Sim, enc *EncounterState, idx int) string {
	switch idx {
	case 0: // Identify yourself
		if sim.Skills.AddXP(SkillDiplomacy, 1.0) {
			LogLevelUp(sim.Log, SkillDiplomacy, sim.Skills.Level(SkillDiplomacy))
		}
		return "\"Credentials check out. Carry on, civilian. Stay safe.\""

	case 1: // Report pirate activity
		// Check if there's a pirate in the current system
		sm := sim.Sector.CurrentSystemMap()
		hasPirate := false
		for i := range sm.Objects {
			if sm.Objects[i].Kind == ObjShip && sm.Objects[i].AIKind == AIPirate {
				hasPirate = true
				break
			}
		}
		if hasPirate {
			sim.Resources.Credits += 25
			if sim.Skills.AddXP(SkillDiplomacy, 3.0) {
				LogLevelUp(sim.Log, SkillDiplomacy, sim.Skills.Level(SkillDiplomacy))
			}
			sim.Log.Add("Bounty received: +25cr.", MsgDiscovery)
			return "\"Confirmed. We've dispatched units. Here's a bounty for the intel. +25cr.\""
		}
		return "\"We have no reports of pirate activity in this sector. False alarm.\""

	case 2: // Request escort
		return "\"We can't spare the resources for an escort right now.\nStay on marked lanes and you'll be fine. Probably.\""

	case 3: // Ignore
		sim.Log.Add("Patrol logs you as uncooperative.", MsgWarning)
		return "The patrol vessel notes your non-compliance and moves on."
	}
	return ""
}

func resolvePirate(sim *Sim, enc *EncounterState, idx int) string {
	switch idx {
	case 0: // Surrender cargo
		lostCargo := 0
		for i := range sim.Resources.CargoPads {
			pad := &sim.Resources.CargoPads[i]
			if pad.Kind != CargoNone {
				lostCargo += pad.Count
				pad.Count = 0
				pad.Kind = CargoNone
			}
		}
		lostCredits := sim.Resources.Credits / 2
		sim.Resources.Credits -= lostCredits
		sim.Log.Add(fmt.Sprintf("Lost %d cargo units and %dcr to pirates.", lostCargo, lostCredits), MsgCritical)
		return fmt.Sprintf("\"Pleasure doing business.\"\nYou lost %d cargo units and %dcr.", lostCargo, lostCredits)

	case 1: // Bribe
		cost := 30
		if sim.Resources.Credits < cost {
			enc.Resolved = false // let them pick again
			return fmt.Sprintf("You don't have %dcr. The pirate is not amused.", cost)
		}
		sim.Resources.Credits -= cost
		if sim.Skills.AddXP(SkillDiplomacy, 2.0) {
			LogLevelUp(sim.Log, SkillDiplomacy, sim.Skills.Level(SkillDiplomacy))
		}
		sim.Log.Add(fmt.Sprintf("Bribed pirate: -%dcr.", cost), MsgWarning)
		return fmt.Sprintf("\"Smart choice.\" The pirate pockets your %dcr and warps away.", cost)

	case 2: // Bluff
		if sim.Skills.AddXP(SkillDiplomacy, 5.0) {
			LogLevelUp(sim.Log, SkillDiplomacy, sim.Skills.Level(SkillDiplomacy))
		}
		// Skill check: higher diplomacy = better odds
		dipLevel := sim.Skills.Level(SkillDiplomacy)
		seed := sim.Sector.Seed*999 + int64(enc.ShipObj.X) + int64(sim.Ticks)
		rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>4|7)))
		roll := rng.IntN(10)
		// Need to roll under diplomacy level (min 3 to even try)
		if roll < dipLevel {
			sim.Log.Add("Bluff successful! The pirate backs down.", MsgDiscovery)
			return "\"Wait... you're with the Patrol? Forget it, we're leaving!\"\nYour bluff worked."
		}
		sim.Log.Add("Bluff failed. The pirate sees through you.", MsgWarning)
		enc.Resolved = false // let them pick again — still in the encounter
		return "\"Nice try, but I wasn't born yesterday.\" The pirate isn't fooled."

	case 3: // Flee
		// Make the pirate chase harder
		enc.ShipObj.MoveRate = max(enc.ShipObj.MoveRate-2, 3)
		enc.ShipObj.dirTimer = 30
		sim.Log.Add("Fleeing! The pirate gives chase.", MsgWarning)
		return "You break off communications and gun the engines. The pirate follows."

	case 4: // Fight (disabled)
		return "Combat systems offline."
	}
	return ""
}

// encounterKindLabel returns a display label for the encounter kind.
func EncounterKindLabel(kind EncounterKind) string {
	switch kind {
	case EncounterTrader:
		return "Merchant Vessel"
	case EncounterPatrol:
		return "Patrol Vessel"
	case EncounterPirate:
		return "Pirate Vessel"
	default:
		return "Unknown Vessel"
	}
}
