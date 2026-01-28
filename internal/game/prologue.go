package game

import (
	"fmt"
	"math/rand/v2"
)

// PrologueScenario holds the generated starting situation.
type PrologueScenario struct {
	Location     PrologueLocation
	Stranding    StrandingReason
	ShuttleState ShuttleCondition
	Objective    string // what player needs to do
	Flavor       string // narrative description
}

// PrologueLocation is where you wake up stranded.
type PrologueLocation uint8

const (
	LocBarrenPlanet PrologueLocation = iota
	LocIcePlanet
	LocVolcanicPlanet
	LocAbandonedOutpost
	LocDerelictStation
	LocCrashSite
)

// StrandingReason is why you got left behind.
type StrandingReason uint8

const (
	StrandAwayTeamAbandoned StrandingReason = iota // ML had to bug out, left you
	StrandShuttleCrashed                           // your shuttle went down
	StrandWokeUpAlone                              // woke up in wreckage, everyone gone
	StrandEscapedCaptivity                         // broke out of pirate/hostile prison
	StrandTransporterGlitch                        // beamed down, transporter fried
	StrandMissionGoneWrong                         // mission failed, only survivor
)

// ShuttleCondition is the state of your ride out.
type ShuttleCondition uint8

const (
	ShuttleNeedsFuel   ShuttleCondition = iota // has power, just empty tank
	ShuttleNeedsRepair                         // engine busted, need parts
	ShuttleNeedsPower                          // dead battery, need to charge
	ShuttleNeedsAll                            // it's bad - fuel + repair + power
)

var prologueLocationNames = map[PrologueLocation]string{
	LocBarrenPlanet:     "barren planet",
	LocIcePlanet:        "frozen moon",
	LocVolcanicPlanet:   "volcanic hellscape",
	LocAbandonedOutpost: "abandoned outpost",
	LocDerelictStation:  "derelict station",
	LocCrashSite:        "crash site",
}

var strandingDescriptions = map[StrandingReason]string{
	StrandAwayTeamAbandoned: "The Monkey Lion jumped out without you. Emergency protocol. The away team scattered. You're the only one who made it here.",
	StrandShuttleCrashed:    "Your shuttle hit something on descent. Maybe debris, maybe hostile fire. Doesn't matter now. It's scrap.",
	StrandWokeUpAlone:       "You woke up in a med bay. The lights were flickering. Everyone was gone. You don't remember how you got here.",
	StrandEscapedCaptivity:  "You broke out. Took out a guard, found a maintenance hatch. Now you're free, but stranded.",
	StrandTransporterGlitch: "The transporter beam scattered. You rematerialized here. The Monkey Lion's signal is gone. Comms are dead.",
	StrandMissionGoneWrong:  "The mission went sideways. Hostiles. Ambush. You're the only one who made it out.",
}

var shuttleObjectives = map[ShuttleCondition]string{
	ShuttleNeedsFuel:   "Find fuel cells to refill the shuttle's tank",
	ShuttleNeedsRepair: "Find spare parts to repair the shuttle's engine",
	ShuttleNeedsPower:  "Find a power source to charge the shuttle's battery",
	ShuttleNeedsAll:    "The shuttle needs everything - fuel, parts, and power",
}

// GeneratePrologue creates a randomized starting scenario.
func GeneratePrologue(seed int64) *PrologueScenario {
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|7)))

	loc := PrologueLocation(rng.IntN(6))
	strand := StrandingReason(rng.IntN(6))
	shuttle := ShuttleCondition(rng.IntN(4))

	// Some combinations make more sense - adjust
	// Crashed shuttle + crash site = redundant, pick different location
	if strand == StrandShuttleCrashed && loc == LocCrashSite {
		loc = PrologueLocation(rng.IntN(3)) // planet type instead
	}
	// Escaped captivity usually means outpost or station
	if strand == StrandEscapedCaptivity && loc < LocAbandonedOutpost {
		loc = LocAbandonedOutpost + PrologueLocation(rng.IntN(2))
	}

	flavor := buildPrologueFlavor(loc, strand, shuttle)

	return &PrologueScenario{
		Location:     loc,
		Stranding:    strand,
		ShuttleState: shuttle,
		Objective:    shuttleObjectives[shuttle],
		Flavor:       flavor,
	}
}

func buildPrologueFlavor(loc PrologueLocation, strand StrandingReason, shuttle ShuttleCondition) string {
	locName := prologueLocationNames[loc]
	strandDesc := strandingDescriptions[strand]

	shuttleDesc := ""
	switch shuttle {
	case ShuttleNeedsFuel:
		shuttleDesc = "There's a shuttle nearby. Looks intact, but the fuel gauge reads empty."
	case ShuttleNeedsRepair:
		shuttleDesc = "You found a shuttle. The engine's shot, but maybe you can scavenge parts."
	case ShuttleNeedsPower:
		shuttleDesc = "There's a shuttle here. Dead. The battery's completely drained."
	case ShuttleNeedsAll:
		shuttleDesc = "You found a shuttle. It's a wreck - needs fuel, parts, and power. But it's your only way out."
	}

	return fmt.Sprintf("You're stranded on a %s.\n\n%s\n\n%s", locName, strandDesc, shuttleDesc)
}

// LocationToPlanetKind converts prologue location to planet type for surface gen.
func (p *PrologueScenario) LocationToPlanetKind() PlanetKind {
	switch p.Location {
	case LocIcePlanet:
		return PlanetIce
	case LocVolcanicPlanet:
		return PlanetVolcanic
	case LocAbandonedOutpost, LocDerelictStation, LocCrashSite:
		return PlanetRocky // interior/station maps will use TerrainInterior
	default:
		return PlanetRocky
	}
}

// IsInterior returns true if the location is an indoor/station environment.
func (p *PrologueScenario) IsInterior() bool {
	switch p.Location {
	case LocAbandonedOutpost, LocDerelictStation:
		return true
	default:
		return false
	}
}

// PrologueObjectiveKind returns what the player needs to find.
type PrologueObjectiveKind uint8

const (
	PrologueObjFuel   PrologueObjectiveKind = iota // find fuel cells
	PrologueObjParts                               // find engine parts
	PrologueObjPower                               // find power cell
	PrologueObjMulti                               // find multiple things
)

// GetObjectives returns the list of things the player needs to find.
func (p *PrologueScenario) GetObjectives() []PrologueObjectiveKind {
	switch p.ShuttleState {
	case ShuttleNeedsFuel:
		return []PrologueObjectiveKind{PrologueObjFuel}
	case ShuttleNeedsRepair:
		return []PrologueObjectiveKind{PrologueObjParts}
	case ShuttleNeedsPower:
		return []PrologueObjectiveKind{PrologueObjPower}
	case ShuttleNeedsAll:
		return []PrologueObjectiveKind{PrologueObjFuel, PrologueObjParts, PrologueObjPower}
	default:
		return []PrologueObjectiveKind{PrologueObjFuel}
	}
}

// PrologueLocationName returns the display name for the location.
func PrologueLocationName(loc PrologueLocation) string {
	if name, ok := prologueLocationNames[loc]; ok {
		return name
	}
	return "unknown location"
}
