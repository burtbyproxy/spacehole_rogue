package game

import "fmt"

// SkillID identifies a player skill.
type SkillID uint8

const (
	SkillEngineering SkillID = iota
	SkillCombat
	SkillPiloting
	SkillScience
	SkillDiplomacy
	SkillLeadership
	SkillSurvival
	SkillCount // sentinel
)

// PlayerSkills holds XP and computed levels for each skill.
type PlayerSkills struct {
	XP [SkillCount]float64
}

// skillXPTable maps level → cumulative XP required.
// Level 1 = 0 XP (starting), Level 10 = 1000 XP (max).
var skillXPTable = [11]float64{
	0,    // 0 (unused)
	0,    // level 1
	40,   // level 2
	90,   // level 3
	160,  // level 4
	250,  // level 5
	360,  // level 6
	490,  // level 7
	640,  // level 8
	810,  // level 9
	1000, // level 10
}

// xpToLevel converts cumulative XP to a level (1-10).
func xpToLevel(xp float64) int {
	for i := 10; i >= 1; i-- {
		if xp >= skillXPTable[i] {
			return i
		}
	}
	return 1
}

// Level returns the integer level (1-10) for a skill.
func (ps *PlayerSkills) Level(id SkillID) int {
	return xpToLevel(ps.XP[id])
}

// AddXP adds XP to a skill and returns true if the player leveled up.
func (ps *PlayerSkills) AddXP(id SkillID, amount float64) bool {
	oldLevel := ps.Level(id)
	ps.XP[id] += amount
	if ps.XP[id] > skillXPTable[10] {
		ps.XP[id] = skillXPTable[10]
	}
	return ps.Level(id) > oldLevel
}

// XPProgress returns (current XP into this level, XP needed for next level).
func (ps *PlayerSkills) XPProgress(id SkillID) (current, needed float64) {
	lvl := ps.Level(id)
	if lvl >= 10 {
		return ps.XP[id] - skillXPTable[10], 0
	}
	floor := skillXPTable[lvl]
	ceiling := skillXPTable[lvl+1]
	return ps.XP[id] - floor, ceiling - floor
}

// SkillName returns the display name for a skill.
func SkillName(id SkillID) string {
	switch id {
	case SkillEngineering:
		return "Engineering"
	case SkillCombat:
		return "Combat"
	case SkillPiloting:
		return "Piloting"
	case SkillScience:
		return "Science"
	case SkillDiplomacy:
		return "Diplomacy"
	case SkillLeadership:
		return "Leadership"
	case SkillSurvival:
		return "Survival"
	default:
		return "Unknown"
	}
}

// skillPerks holds flavor text for each skill level.
var skillPerks = [SkillCount][11]string{
	SkillEngineering: {
		"",
		"Knows which end of the wrench to hold",
		"Can identify pipe leaks by sound",
		"Faster repairs",
		"Diagnose problems before they break",
		"Jury-rig solutions from spare parts",
		"Pipe condition visible at a glance",
		"Subsystem efficiency +15%",
		"Can repair shield emitters",
		"Master engineer: half repair time",
		"Miracle worker (literally)",
	},
	SkillCombat: {
		"",
		"Knows which end of the phaser to point",
		"Can hit a stationary target (usually)",
		"Decent aim under pressure",
		"Tactical awareness in firefights",
		"Combat reflexes sharpened",
		"Can dual-wield (inadvisable but cool)",
		"Predictive targeting",
		"Battlefield commander",
		"One-person army",
		"They write songs about you",
	},
	SkillPiloting: {
		"",
		"Can find the throttle",
		"Smooth docking maneuvers",
		"Fuel-efficient flight paths",
		"Evasive maneuvers unlocked",
		"Fuel efficiency +10%",
		"Asteroid field navigation",
		"Combat pilot certification",
		"Drift king of the outer rim",
		"Can fly anything with thrusters",
		"The shuttle is an extension of your body",
	},
	SkillScience: {
		"",
		"Can read a scanner without squinting",
		"Better scan resolution",
		"Anomaly pattern recognition",
		"Deep-scan mineral analysis",
		"Identify alien tech on sight",
		"Research speed doubled",
		"Predict stellar phenomena",
		"Xenobiology specialist",
		"Theoretical physics breakthrough",
		"Deborah is impressed (she's a zebra)",
	},
	SkillDiplomacy: {
		"",
		"Can say 'hello' without offending anyone",
		"Haggling instincts",
		"Better trade prices",
		"Read NPC intentions",
		"Faction reputation bonus",
		"Negotiate under fire",
		"Alliance broker",
		"Silver tongue",
		"Galactic diplomat",
		"Could sell ice to an ice planet",
	},
	SkillLeadership: {
		"",
		"Crew tolerates your presence",
		"Basic crew management",
		"Crew morale bonus",
		"Inspire loyalty",
		"Tactical command",
		"Crisis management",
		"Crew follows without question",
		"Legendary commander",
		"Fleet admiral material",
		"They'd follow you into a black hole",
	},
	SkillSurvival: {
		"",
		"Remembers to eat and drink",
		"Efficient metabolism",
		"Slower hunger and thirst",
		"Resist minor infections",
		"Environmental resilience",
		"Survive on minimal rations",
		"Extreme environment tolerance",
		"Can eat almost anything",
		"Cockroach-level survivability",
		"Death keeps losing your address",
	},
}

// SkillPerk returns the perk description for a skill at a given level.
func SkillPerk(id SkillID, level int) string {
	if level < 1 || level > 10 {
		return ""
	}
	return skillPerks[id][level]
}

// LogLevelUp logs a skill level-up to the message log.
func LogLevelUp(log *MessageLog, id SkillID, level int) {
	log.Add(fmt.Sprintf("%s leveled up to %d! %s", SkillName(id), level, SkillPerk(id, level)), MsgDiscovery)
}
