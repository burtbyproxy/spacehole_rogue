package game

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// MissionCategory groups missions by gameplay style.
type MissionCategory uint8

const (
	CatInvestigate MissionCategory = iota
	CatMilitary
	CatResearch
	CatSupport
)

// MissionType identifies a specific mission scenario.
type MissionType uint8

const (
	// Investigate
	MissionCrashSite MissionType = iota
	MissionDerelictShip
	MissionDistressSignal
	MissionEveryoneDisappeared
	MissionMissingScientist
	MissionVanishedShip
	// Military
	MissionInspection
	MissionPursuit
	MissionShoreLeave
	// Research
	MissionBlackhole
	MissionCivilization
	MissionNebula
	MissionPlanet
	MissionQuasar
	MissionStar
	MissionUnexploredSystem
	MissionWormhole
	// Support
	MissionDefend
	MissionDeliver
	MissionMedical
	MissionRescue
	MissionTransport
	MissionCount // sentinel
)

// TwistType adds a complication or surprise to an episode.
type TwistType uint8

const (
	TwistAssassinationAttempt TwistType = iota
	TwistCrewInfected
	TwistEquipmentMalfunction
	TwistMarooned
	TwistOfficerInsane
	TwistSeriesOfMurders
	TwistShipCaptured
	TwistShipDamaged
	TwistSurpriseAttack
	TwistTakenPrisoner
	TwistThoughtsManifested
	TwistTimeTravel
	TwistCourtMartialed
	TwistCount
)

// LocationType describes the setting of an episode.
type LocationType uint8

const (
	LocAlternateDimension LocationType = iota
	LocCivilianColony
	LocParadiseGarden
	LocFalseUtopia
	LocMilitaryOutpost
	LocMiningBase
	LocPleasureDistrict
	LocPrehistoricPlanet
	LocPrison
	LocResearchBase
	LocStarknightCommand
	LocCount
)

// CharacterType identifies the featured NPC archetype.
type CharacterType uint8

const (
	CharAlienAmbassador CharacterType = iota
	CharAmbitiousOfficer
	CharBrainwashedColonists
	CharCreepyChildren
	CharDerangedScientist
	CharEccentricTrader
	CharEvilTwin
	CharGeneticSuperhuman
	CharGiantCube
	CharHistoricalFigure
	CharHonorableEnemyCaptain
	CharHotshotPilot
	CharLonelyGodling
	CharLoveInterest
	CharMassiveSingleCelledOrganisms
	CharMoltenStoneCreature
	CharOldRival
	CharPowerfulPsychic
	CharPrimitiveMonster
	CharReclusiveDictator
	CharRobotOverlord
	CharRogueSatellite
	CharSecretWeapon
	CharSentientCloud
	CharShadyDiplomat
	CharShakespeareanActingTroupe
	CharSpaceHippies
	CharSupercomputer
	CharWarCriminal
	CharCount
)

// ---------------------------------------------------------------------------
// Episode state
// ---------------------------------------------------------------------------

// EpisodeState holds the generated episode for a system visit.
type EpisodeState struct {
	Mission   MissionType
	Twist     TwistType
	Location  LocationType
	Character CharacterType
	Title     string
	Briefing  string // multi-line, assembled from templates
	Options   []EpisodeOption
	ResultText string
	Resolved   bool
	MLClue     bool // true if this episode dropped an ML clue
}

// EpisodeOption is a single choice the player can make.
type EpisodeOption struct {
	Label       string
	Enabled     bool
	DisableText string
	SkillReq    SkillID
	SkillLevel  int
}

// ---------------------------------------------------------------------------
// Data tables
// ---------------------------------------------------------------------------

var missionNames = [MissionCount]string{
	"Crash Site", "Derelict Ship", "Distress Signal",
	"Everyone Disappeared", "Missing Scientist", "Vanished Ship",
	"Inspection", "Pursuit", "Shore Leave",
	"Blackhole", "Civilization", "Nebula", "Planet", "Quasar",
	"Star", "Unexplored System", "Wormhole",
	"Defend", "Deliver", "Medical", "Rescue", "Transport",
}

var missionCategories = [MissionCount]MissionCategory{
	CatInvestigate, CatInvestigate, CatInvestigate,
	CatInvestigate, CatInvestigate, CatInvestigate,
	CatMilitary, CatMilitary, CatMilitary,
	CatResearch, CatResearch, CatResearch, CatResearch, CatResearch,
	CatResearch, CatResearch, CatResearch,
	CatSupport, CatSupport, CatSupport, CatSupport, CatSupport,
}

// Mission briefing templates. Slots: {system}, {location}, {char}
var missionBriefings = [MissionCount]string{
	// Investigate
	"Sensors detect a crash site near {location} in the {system} system. Wreckage is scattered across the surface. {char}",
	"A derelict vessel drifts near {location} in the {system} system. No life signs detected. Hull breached in multiple places. {char}",
	"A weak distress signal broadcasts from {location} in the {system} system. The transmission is fragmented but urgent. {char}",
	"All contact with {location} in the {system} system has been lost. The settlement went silent three days ago. No signals, no beacons. {char}",
	"A research team at {location} reports their lead scientist has vanished. Equipment is still running but the lab is empty. {char}",
	"A registered vessel was last tracked near {location} in the {system} system. It never arrived at its destination. {char}",
	// Military
	"A patrol authority near {location} in the {system} system demands you submit to inspection. They claim jurisdiction over this sector. {char}",
	"An alert broadcasts from {location} in the {system} system — a fugitive vessel was spotted in the area. Authorities request pursuit assistance. {char}",
	"You pick up a shore leave beacon from {location} in the {system} system. It's a designated rest stop for passing ships. {char}",
	// Research
	"A stable blackhole has been detected near {location} in the {system} system. Its accretion disk pulses with unusual energy patterns. {char}",
	"Sensors reveal a pre-warp civilization on a planet near {location} in the {system} system. They appear unaware of spacefaring species. {char}",
	"A dense nebula near {location} in the {system} system is emitting anomalous radiation patterns. Standard sensors can barely penetrate it. {char}",
	"An uncharted planet near {location} in the {system} system shows unusual geological activity. The surface readings don't match any known world type. {char}",
	"A quasar-like energy source has appeared near {location} in the {system} system. It shouldn't be possible at this scale. {char}",
	"The star in the {system} system is behaving erratically near {location}. Solar output is fluctuating wildly outside predicted models. {char}",
	"The region around {location} in the {system} system has never been charted. Long-range scans show... something. {char}",
	"A spatial anomaly near {location} in the {system} system reads as a stable wormhole. The other end is unknown. {char}",
	// Support
	"A settlement at {location} in the {system} system is under attack. They're broadcasting on all frequencies for help. {char}",
	"A supply request originates from {location} in the {system} system. They need essential cargo delivered urgently. {char}",
	"A medical emergency has been declared at {location} in the {system} system. A pathogen is spreading through the population. {char}",
	"A distress call from {location} in the {system} system reports people trapped in a structural collapse. Time is running out. {char}",
	"Refugees near {location} in the {system} system need transport to safety. Their ship's engines are failing. {char}",
}

var twistNames = [TwistCount]string{
	"Assassination Attempt", "Crew Infected", "Equipment Malfunction",
	"Marooned", "Officer Goes Insane", "Series of Murders",
	"Ship Captured", "Ship Damaged", "Surprise Attack",
	"Taken Prisoner", "Thoughts Manifested", "Time Travel",
	"Unjustly Court Martialed",
}

// Twist foreshadowing — subtle hint in the briefing.
var twistHints = [TwistCount]string{
	"You notice suspicious movement on sensors. Someone is watching.",
	"Bioscans detect unusual pathogen signatures in the area.",
	"Warning: shuttle systems showing intermittent faults.",
	"Navigation confirms you're far from any safe harbor.",
	"Comms chatter sounds... erratic. Something's off with someone out there.",
	"Scanners detect anomalous bio-signatures. Multiple readings, all fading.",
	"Sensors pick up vessels maneuvering into position around you.",
	"Your hull integrity sensors are picking up micro-fracture warnings.",
	"Your sensors flicker. Something feels wrong.",
	"A tractor beam signature registers briefly, then disappears.",
	"Reality feels... thin here. Instruments confirm spatial instability.",
	"Chronometric readings are... inconsistent. Time is behaving strangely.",
	"A judicial transmission is broadcasting on official frequencies.",
}

// Twist reveals — what actually happens (appended to result text).
var twistReveals = [TwistCount]string{
	"An assassin strikes! A hidden weapon fires at your shuttle. You take evasive action.",
	"Warning — pathogen detected aboard! Your systems are contaminated.",
	"Critical malfunction! Your recycler overloads, venting matter into space.",
	"Your nav systems lock up. You're stranded until you can reroute power.",
	"The officer aboard the nearby vessel has gone completely unhinged, raving about voices.",
	"You discover evidence of multiple deaths. This wasn't an accident.",
	"Energy dampeners activate — your shuttle is caught in a tractor web!",
	"An impact rocks the shuttle. Hull breach warning!",
	"Ambush! Vessels decloak and open fire! You take evasive action.",
	"Force fields snap on around you. You've been captured!",
	"Your thoughts become real. The shuttle fills with... was that always there?",
	"Space warps around you. When it clears, the stars have shifted. Time has jumped.",
	"You're being charged with crimes you didn't commit. A tribunal convenes.",
}

var locationNames = [LocCount]string{
	"an alternate dimension rift", "a civilian colony", "a paradise garden world",
	"a false utopia", "a military outpost", "a mining base",
	"a pleasure district station", "a prehistoric planet", "a prison facility",
	"a research base", "Starknight Command",
}

var characterTypeNames = [CharCount]string{
	"Alien Ambassador", "Ambitious Officer", "Brainwashed Colonists",
	"Creepy Children", "Deranged Scientist", "Eccentric Trader",
	"Evil Twin", "Genetic Superhuman", "Giant Cube",
	"Historical Figure", "Honorable Enemy Captain", "Hotshot Pilot",
	"Lonely Godling", "Love Interest", "Massive Single Celled Organisms",
	"Molten Stone Creature", "Old Rival", "Powerful Psychic",
	"Primitive Monster", "Reclusive Dictator", "Robot Overlord",
	"Rogue Satellite", "Secret Weapon", "Sentient Cloud",
	"Shady Diplomat", "Shakespearean Acting Troupe", "Space Hippies",
	"Supercomputer", "War Criminal",
}

// Character introductions — {name} slot for proper name.
var characterIntros = [CharCount]string{
	"{name}, an alien ambassador, hails your shuttle with formal greetings.",
	"An ambitious young officer named {name} demands your attention on comms.",
	"The colonists at the settlement stare blankly. They speak in unison.",
	"Children peer at you through the viewport. Their eyes are... wrong.",
	"{name}, a wild-eyed researcher, hails your shuttle frantically.",
	"{name}, an eccentric trader, broadcasts a deal too good to be true.",
	"Someone who looks exactly like you appears on the viewscreen. They smile.",
	"{name}, a genetically enhanced being, radiates an unsettling calm.",
	"A massive metallic cube orbits silently nearby. It does not respond to hails.",
	"Historical records identify {name} — but that's impossible. They died centuries ago.",
	"{name}, an enemy captain of considerable reputation, hails with unexpected courtesy.",
	"A hotshot pilot called {name} buzzes your shuttle, showing off.",
	"An entity calling itself {name} claims to be a god. A lonely one.",
	"{name} appears on your viewscreen. Something about them is... magnetic.",
	"Sensors show a single organism. It's the size of a moon. It's alive.",
	"A creature of molten stone lumbers across the surface, radiating incredible heat.",
	"You recognize {name} on the comm channel. An old rival. This won't be simple.",
	"{name}, a powerful psychic, contacts you telepathically before you even hail.",
	"Something massive and primitive roars on the surface. It's territorial.",
	"{name}, a reclusive dictator, grants you a rare audience via encrypted channel.",
	"An artificial intelligence designated {name} controls everything here.",
	"A rogue satellite locks onto your shuttle and begins transmitting data.",
	"Intelligence reports reference a secret weapon hidden in this system.",
	"A sentient cloud drifts toward you, pulsing with bioluminescent patterns.",
	"{name}, a diplomat of questionable reputation, offers to negotiate.",
	"A troupe of Shakespearean actors hails you, mid-performance of Hamlet.",
	"A commune of Space Hippies broadcasts peace symbols and folk music.",
	"A vast supercomputer called {name} speaks in perfect monotone.",
	"{name}, a wanted war criminal, is reportedly hiding in this system.",
}

// Proper names for characters — seeded RNG picks from these per character type.
var characterProperNamePool = [CharCount][]string{
	{"Thex", "Zira", "Ambassador Kol", "Envoy Tal'Set", "Diplomat Vreen"},       // Alien Ambassador
	{"Lt. Harker", "Cmdr. Voss", "Ensign Zhao", "Lt. Pryce", "Cmdr. Nash"},     // Ambitious Officer
	{"the colonists", "the settlers", "the inhabitants", "the population"},       // Brainwashed Colonists
	{"the children", "the young ones", "the watchers"},                           // Creepy Children
	{"Dr. Voss", "Dr. Krell", "Professor Zahn", "Dr. Mira", "Dr. Okkonen"},     // Deranged Scientist
	{"Korb", "Madame Luxe", "Trader Nim", "the Merchant", "Dealmaker Fenn"},     // Eccentric Trader
	{"your double", "the impostor", "the mirror", "the other you"},               // Evil Twin
	{"Apex", "Nova", "Subject Zero", "The Perfected", "Augment Kael"},           // Genetic Superhuman
	{"the Cube", "Object Seven", "the Monolith", "Grid Alpha"},                  // Giant Cube
	{"Admiral Chen", "Captain Kirk", "General Voss", "Commander Shran"},          // Historical Figure
	{"Captain D'vak", "Commander Torek", "Captain Sela", "Captain Krenn"},       // Honorable Enemy Captain
	{"Ace", "Maverick", "Flash", "Daredevil", "Stardust"},                       // Hotshot Pilot
	{"Eternus", "The Solitary", "Monad", "The Forsaken"},                        // Lonely Godling
	{"Alex", "Morgan", "Casey", "Jordan", "Quinn"},                              // Love Interest
	{"the Organism", "Macro-Entity", "the Living Moon", "Bio-Mass Alpha"},       // Massive Single Celled Organisms
	{"the Golem", "Ignis", "the Colossus", "Pyrax"},                             // Molten Stone Creature
	{"Rennick", "Castillo", "your old nemesis", "Torres", "Blake"},              // Old Rival
	{"Sylar", "Mindkeeper", "The Oracle", "Psion", "Thought-weaver"},            // Powerful Psychic
	{"the Beast", "the Leviathan", "the Horror", "the Predator"},                // Primitive Monster
	{"Dictator Vorn", "Supreme Leader Kael", "Tyrant Drexx", "the Overlord"},    // Reclusive Dictator
	{"NEXUS-9", "Sovereign", "the Machine", "AXIOM", "Unit Prime"},              // Robot Overlord
	{"SAT-7", "Orbital-X", "the Probe", "Deep Eye"},                            // Rogue Satellite
	{"Project Omega", "the Device", "Codename Fist", "the Prototype"},           // Secret Weapon
	{"the Cloud", "Nimbus", "the Mist", "Vapor", "the Drift"},                  // Sentient Cloud
	{"Ambassador Krel", "Envoy Shade", "Diplomat Vex", "Consul Nix"},            // Shady Diplomat
	{"the Players", "the Troupe", "the Company", "the Thespians"},               // Shakespearean Acting Troupe
	{"the Collective", "the Commune", "Star Children", "the Free Folk"},          // Space Hippies
	{"ORACLE", "CORE", "ATLAS", "MINERVA", "LOGOS"},                             // Supercomputer
	{"Dax Vrenn", "General Thule", "the Butcher", "Colonel Hask", "Krell"},      // War Criminal
}

// ML clue texts — dropped ~10% for eligible characters.
var mlClueTexts = []string{
	"In the data you recovered, there's a fragment: '...the Monkey Lion\nawaits beyond the veil of stars...'",
	"Among the records, star charts reference a massive vessel\ndesignation ML-7 — last seen jumping to unknown coordinates.",
	"A corrupted log entry reads: '...SpaceHole Drive engaged.\nThe Monkey Lion vanished from all sensors...'",
	"You find a personal journal mentioning the USS Monkey Lion:\n'They said it could punch holes in space itself.'",
	"Encrypted coordinates recovered. Cross-referencing against\nknown SH Drive signatures... partial match detected.",
}

// Characters eligible for ML clue drops.
var mlClueEligible = map[CharacterType]bool{
	CharEccentricTrader:  true,
	CharHistoricalFigure: true,
	CharDerangedScientist: true,
	CharSentientCloud:    true,
	CharSupercomputer:    true,
}

// ---------------------------------------------------------------------------
// Options per mission category
// ---------------------------------------------------------------------------

type categoryOptions struct {
	labels     [4]string
	skills     [4]SkillID // primary skill XP for each option
	skillAmts  [4]float64 // XP amounts
}

var investigateOpts = categoryOptions{
	labels:    [4]string{"Investigate up close", "Scan from safe distance", "Hail and request information", "Move on"},
	skills:    [4]SkillID{SkillSurvival, SkillScience, SkillDiplomacy, SkillPiloting},
	skillAmts: [4]float64{5, 4, 3, 0},
}

var militaryOpts = categoryOptions{
	labels:    [4]string{"Comply and engage", "Negotiate", "Attempt evasion", "Move on"},
	skills:    [4]SkillID{SkillCombat, SkillDiplomacy, SkillPiloting, SkillPiloting},
	skillAmts: [4]float64{5, 4, 3, 0},
}

var researchOpts = categoryOptions{
	labels:    [4]string{"Conduct detailed scans", "Collect samples", "Record observations and log", "Move on"},
	skills:    [4]SkillID{SkillScience, SkillEngineering, SkillScience, SkillPiloting},
	skillAmts: [4]float64{5, 4, 2, 0},
}

var supportOpts = categoryOptions{
	labels:    [4]string{"Provide full assistance", "Offer partial help", "Advise and move on", "Decline"},
	skills:    [4]SkillID{SkillDiplomacy, SkillDiplomacy, SkillDiplomacy, SkillPiloting},
	skillAmts: [4]float64{5, 3, 2, 0},
}

var categoryOptionSets = [4]categoryOptions{
	investigateOpts, militaryOpts, researchOpts, supportOpts,
}

// ---------------------------------------------------------------------------
// Rolling
// ---------------------------------------------------------------------------

// RollEpisode generates a procedural episode for a system visit.
// Returns nil if no episode triggers (~40% chance).
func RollEpisode(sectorSeed int64, sysIdx int, systemName string, skills *PlayerSkills) *EpisodeState {
	seed := sectorSeed*2000 + int64(sysIdx)*13
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|7)))

	// 60% chance of episode
	if rng.IntN(100) >= 60 {
		return nil
	}

	// Roll the 4 tables
	mission := MissionType(rng.IntN(int(MissionCount)))
	twist := TwistType(rng.IntN(int(TwistCount)))
	location := LocationType(rng.IntN(int(LocCount)))
	character := CharacterType(rng.IntN(int(CharCount)))

	// Pick a proper name for the character
	pool := characterProperNamePool[character]
	charName := pool[rng.IntN(len(pool))]

	// Assemble title
	title := strings.ToUpper(missionNames[mission]) + " at " + strings.ToUpper(locationNames[location])

	// Assemble briefing
	cat := missionCategories[mission]
	briefingTemplate := missionBriefings[mission]
	briefing := briefingTemplate
	briefing = strings.ReplaceAll(briefing, "{system}", systemName)
	briefing = strings.ReplaceAll(briefing, "{location}", locationNames[location])
	briefing = strings.ReplaceAll(briefing, "{char}", "")

	charIntro := characterIntros[character]
	charIntro = strings.ReplaceAll(charIntro, "{name}", charName)

	twistHint := twistHints[twist]

	fullBriefing := strings.TrimSpace(briefing) + "\n\n" + charIntro + "\n\n" + twistHint

	// Build options from category
	opts := categoryOptionSets[cat]
	options := make([]EpisodeOption, 4)
	for i := 0; i < 4; i++ {
		options[i] = EpisodeOption{
			Label:   opts.labels[i],
			Enabled: true,
		}
	}

	// Skill gate: option 1 for military requires Combat Lv 2
	if cat == CatMilitary {
		if skills.Level(SkillCombat) < 2 {
			options[0].Enabled = false
			options[0].DisableText = "Requires Combat Lv 2"
			options[0].SkillReq = SkillCombat
			options[0].SkillLevel = 2
		}
	}

	return &EpisodeState{
		Mission:   mission,
		Twist:     twist,
		Location:  location,
		Character: character,
		Title:     title,
		Briefing:  fullBriefing,
		Options:   options,
	}
}

// ---------------------------------------------------------------------------
// Resolution
// ---------------------------------------------------------------------------

// ResolveEpisode processes the player's choice and returns result text.
func ResolveEpisode(sim *Sim, ep *EpisodeState, optionIdx int) string {
	if optionIdx < 0 || optionIdx >= len(ep.Options) {
		return ""
	}
	opt := ep.Options[optionIdx]
	if !opt.Enabled {
		return ""
	}

	ep.Resolved = true

	cat := missionCategories[ep.Mission]
	opts := categoryOptionSets[cat]

	// Award skill XP
	if opts.skillAmts[optionIdx] > 0 {
		if sim.Skills.AddXP(opts.skills[optionIdx], opts.skillAmts[optionIdx]) {
			LogLevelUp(sim.Log, opts.skills[optionIdx], sim.Skills.Level(opts.skills[optionIdx]))
		}
	}

	// Option 4 (index 3) is always "move on" — no effects
	if optionIdx == 3 {
		sim.Log.Add("You decide to move on.", MsgInfo)
		return "You decide to move on. The episode fades behind you\nas the shuttle continues through the system."
	}

	seed := sim.Sector.Seed*3000 + int64(ep.Mission)*17 + int64(ep.Twist)*31
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|11)))

	charPool := characterProperNamePool[ep.Character]
	charName := charPool[rng.IntN(len(charPool))]

	var baseText string
	var twistText string
	credits := 0
	hullDmg := 0
	energyCost := 0
	var cargoKind CargoKind
	cargoAmt := 0

	// --- Base outcome by option index ---
	switch optionIdx {
	case 0: // Bold option — high risk/reward
		switch cat {
		case CatInvestigate:
			baseText = fmt.Sprintf("You move in close to investigate. %s\nYou find useful salvage among the wreckage.", charName)
			credits = 20 + rng.IntN(20)
			cargoKind = CargoKind(3 + rng.IntN(5)) // random cargo type (ScrapMetal through MedKits)
			cargoAmt = 1 + rng.IntN(2)
		case CatMilitary:
			baseText = "You comply with the directive and engage directly.\nYour combat training serves you well."
			credits = 15 + rng.IntN(15)
		case CatResearch:
			baseText = "You perform a detailed scan of the phenomenon.\nThe data is extraordinary — unlike anything on record."
			credits = 15 + rng.IntN(20)
		case CatSupport:
			baseText = "You provide full assistance, spending time and resources.\nThe gratitude is genuine."
			credits = 25 + rng.IntN(15)
			energyCost = 5
			// Share supplies
			if sim.Resources.Water.Clean >= 10 {
				sim.Resources.Water.Clean -= 10
			}
			if sim.Resources.Organic.Clean >= 10 {
				sim.Resources.Organic.Clean -= 10
			}
		}

	case 1: // Moderate option
		switch cat {
		case CatInvestigate:
			baseText = "You scan from a safe distance. The readings reveal\nuseful data about the area."
			credits = 10 + rng.IntN(10)
		case CatMilitary:
			baseText = fmt.Sprintf("You open negotiations with %s.\nA tense exchange, but you reach an understanding.", charName)
			credits = 10 + rng.IntN(10)
		case CatResearch:
			baseText = "You carefully collect samples from the phenomenon.\nEngineering analysis reveals interesting properties."
			credits = 10 + rng.IntN(10)
			cargoKind = CargoScrapMetal
			cargoAmt = 1
		case CatSupport:
			baseText = "You offer what help you can without overextending.\nIt's enough to make a difference."
			credits = 15 + rng.IntN(10)
		}

	case 2: // Cautious option
		switch cat {
		case CatInvestigate:
			baseText = fmt.Sprintf("You hail %s and request information.\nThey share what they know. Useful intel.", charName)
			credits = 5 + rng.IntN(5)
		case CatMilitary:
			baseText = "You attempt evasion, pushing the shuttle's engines hard.\nYou slip away cleanly."
			credits = 5
		case CatResearch:
			baseText = "You record your observations and add them to the log.\nEvery data point counts for science."
			credits = 5 + rng.IntN(5)
		case CatSupport:
			baseText = "You offer advice and guidance over comms.\nIt's not much, but they appreciate the thought."
			credits = 5
		}
	}

	// --- Twist modifier ---
	twistApplied := false
	switch ep.Twist {
	case TwistSurpriseAttack:
		if optionIdx == 0 {
			if rng.IntN(2) == 0 {
				hullDmg += 15
				twistText = "Ambush! Vessels decloak and open fire!\nYou take evasive action but sustain damage. -" + fmt.Sprintf("%d hull.", hullDmg)
				twistApplied = true
				sim.Skills.AddXP(SkillCombat, 3)
			}
		}
	case TwistEquipmentMalfunction:
		energyCost += 5
		twistText = "Your shuttle systems malfunction mid-operation.\nYou lose power rerouting around the fault. -5 energy."
		twistApplied = true
		sim.Skills.AddXP(SkillEngineering, 3)
	case TwistCrewInfected:
		if optionIdx == 0 {
			if sim.Resources.Organic.Clean >= 5 {
				sim.Resources.Organic.Clean -= 5
				twistText = "Pathogen contamination! Organic matter corrupted. -5 organic."
			} else {
				twistText = "Pathogen detected but containment holds. Close call."
			}
			twistApplied = true
			sim.Skills.AddXP(SkillScience, 3)
		}
	case TwistMarooned:
		energyCost += 10
		twistText = "Nav systems glitch. Getting out of here costs extra fuel. -10 energy."
		twistApplied = true
		sim.Skills.AddXP(SkillSurvival, 3)
	case TwistTimeTravel:
		twistText = "Space warps around you. When it clears, the stars have\nshifted. Your chronometer jumps. What just happened?"
		twistApplied = true
		sim.Skills.AddXP(SkillScience, 5)
	case TwistThoughtsManifested:
		twistText = "For a moment, your thoughts become real. The shuttle fills\nwith something that shouldn't exist. Then it's gone."
		twistApplied = true
		sim.Skills.AddXP(SkillScience, 3)
	case TwistShipDamaged:
		if optionIdx == 0 {
			hullDmg += 10
			twistText = fmt.Sprintf("Debris impact! Your hull takes damage. -%d hull.", hullDmg)
			twistApplied = true
			sim.Skills.AddXP(SkillEngineering, 3)
		}
	case TwistShipCaptured:
		if optionIdx == 0 && rng.IntN(10) < 3 {
			credits = credits / 2 // lose half the reward
			twistText = "Energy dampeners activate! You're temporarily captured.\nYou negotiate your way out but lose half your findings."
			twistApplied = true
			sim.Skills.AddXP(SkillDiplomacy, 3)
		}
	case TwistTakenPrisoner:
		if optionIdx == 0 && rng.IntN(10) < 2 {
			credits = credits / 2
			twistText = "Force fields activate around you. Captured! You talk your\nway out, but it costs you time and findings."
			twistApplied = true
			sim.Skills.AddXP(SkillDiplomacy, 4)
		}
	case TwistSeriesOfMurders:
		twistText = "You discover evidence of multiple deaths. This wasn't\nan accident — someone here is dangerous."
		twistApplied = true
		sim.Skills.AddXP(SkillScience, 2)
	case TwistOfficerInsane:
		twistText = fmt.Sprintf("%s becomes increasingly erratic during the encounter.\nYou calm them down, but it's unsettling.", charName)
		twistApplied = true
		sim.Skills.AddXP(SkillDiplomacy, 3)
	case TwistAssassinationAttempt:
		if rng.IntN(5) == 0 {
			hullDmg += 8
			twistText = fmt.Sprintf("A hidden weapon fires at your shuttle! -%d hull.", hullDmg)
			twistApplied = true
			sim.Skills.AddXP(SkillCombat, 3)
		}
	case TwistCourtMartialed:
		if optionIdx == 0 {
			courtFee := 20
			if credits >= courtFee {
				credits -= courtFee
			} else {
				credits = 0
			}
			twistText = "You're charged with violating sector regulations.\nLegal fees eat into your earnings. -20cr."
			twistApplied = true
			sim.Skills.AddXP(SkillDiplomacy, 3)
		}
	}

	// Apply mechanical effects
	sim.Resources.Credits += credits
	if hullDmg > 0 {
		sim.Resources.Hull -= hullDmg
		if sim.Resources.Hull < 0 {
			sim.Resources.Hull = 0
		}
	}
	if energyCost > 0 {
		sim.Resources.Energy -= energyCost
		if sim.Resources.Energy < 0 {
			sim.Resources.Energy = 0
		}
	}
	if cargoAmt > 0 {
		added := sim.Resources.AddCargo(cargoKind, cargoAmt)
		if added > 0 {
			baseText += fmt.Sprintf("\n+%dx %s added to cargo.", added, CargoName(cargoKind))
		}
	}

	// Credits log
	if credits > 0 {
		baseText += fmt.Sprintf("\n+%dcr.", credits)
	}
	if hullDmg > 0 {
		sim.Log.Add(fmt.Sprintf("Hull damage: -%d. Hull at %d%%.", hullDmg, sim.Resources.HullPct()), MsgWarning)
	}
	if credits > 0 {
		sim.Log.Add(fmt.Sprintf("Earned %dcr. Credits: %d.", credits, sim.Resources.Credits), MsgDiscovery)
	}

	// ML clue check (~10% for eligible characters)
	mlClue := ""
	if mlClueEligible[ep.Character] && rng.IntN(10) == 0 {
		ep.MLClue = true
		mlClue = "\n\n" + mlClueTexts[rng.IntN(len(mlClueTexts))]
		sim.Log.Add("USS Monkey Lion clue discovered!", MsgDiscovery)
	}

	// Assemble final result
	result := baseText
	if twistApplied {
		result += "\n\n" + twistText
	}
	if mlClue != "" {
		result += mlClue
	}

	// Discovery tracking
	sim.Discovery.EpisodesCompleted++
	if ep.MLClue {
		sim.Discovery.MLCluesFound++
	}

	return result
}
