package game

import (
	"fmt"
	"math/rand/v2"
)

// DiscoveryLog tracks what the player has found across the sector.
type DiscoveryLog struct {
	StarTypesSeen [5]bool      // indexed by StarType
	SystemsVisited map[int]bool // indexed by system index
	PlanetsScanned map[string]PlanetScanData // key = "sysIdx:objIdx"
	StationsDocked map[int]bool // indexed by system index

	TotalScans          int
	TotalSystemsVisited int
	TotalStarTypesSeen  int
	TotalStationsDocked int

	RecentScans []PlanetScanData // ordered newest-first, capped at 10
}

// PlanetScanData holds what a planetary scan revealed.
type PlanetScanData struct {
	Name       string
	SystemName string
	PlanetType PlanetKind
	Resources  string
	Hazard     string
	POI        string // empty if no POI detected
}

// NewDiscoveryLog creates an empty discovery log.
func NewDiscoveryLog() *DiscoveryLog {
	return &DiscoveryLog{
		SystemsVisited: make(map[int]bool),
		PlanetsScanned: make(map[string]PlanetScanData),
		StationsDocked: make(map[int]bool),
	}
}

// ScanKey returns the map key for a planet scan.
func ScanKey(sysIdx, objIdx int) string {
	return fmt.Sprintf("%d:%d", sysIdx, objIdx)
}

// Flavor text pools per planet type.

var planetResources = map[PlanetKind][]string{
	PlanetRocky: {
		"Rich iron deposits",
		"Rare mineral veins detected",
		"Silicon reserves in deep crust",
		"Subsurface cave networks",
	},
	PlanetGas: {
		"Hydrogen fuel reserves",
		"Exotic gas pockets",
		"Helium-3 traces in upper atmosphere",
		"Atmospheric chemical anomaly",
	},
	PlanetIce: {
		"Pure water ice sheets",
		"Frozen organic compounds",
		"Cryogenic mineral deposits",
		"Subsurface liquid ocean",
	},
	PlanetVolcanic: {
		"Molten metal flows accessible",
		"Geothermal energy potential",
		"Volcanic glass deposits",
		"Sulphur-rich mineral fields",
	},
}

var planetHazards = map[PlanetKind][]string{
	PlanetRocky: {
		"Unstable tectonic activity",
		"No breathable atmosphere",
		"Micrometeorite storm risk",
	},
	PlanetGas: {
		"Crushing atmospheric pressure",
		"Severe electrical storms",
		"Intense radiation belts",
	},
	PlanetIce: {
		"Extreme cold (-220C surface)",
		"Methane geyser eruptions",
		"Unstable ice shelf collapse risk",
	},
	PlanetVolcanic: {
		"Surface temperature 800C+",
		"Toxic gas emissions",
		"Unpredictable lava eruptions",
	},
}

var planetPOIs = map[PlanetKind][]string{
	PlanetRocky: {
		"Anomalous energy reading",
		"Ancient structure detected",
		"Debris field in orbit",
	},
	PlanetGas: {
		"Unusual radio emissions",
		"Artificial satellite in orbit",
		"Gas composition anomaly",
	},
	PlanetIce: {
		"Thermal signature beneath ice",
		"Faint distress beacon (old)",
		"Crystalline formation of unknown origin",
	},
	PlanetVolcanic: {
		"Heat-resistant alloy deposit",
		"Abandoned mining rig on surface",
		"Unusual seismic pattern",
	},
}

// GenerateScanData creates scan results for a planet using deterministic seeding.
func GenerateScanData(sectorSeed int64, sysIdx, objIdx int, obj *SpaceObject, systemName string) PlanetScanData {
	seed := sectorSeed*10000 + int64(sysIdx)*100 + int64(objIdx)
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|7)))

	pt := obj.PlanetType

	resources := planetResources[pt]
	hazards := planetHazards[pt]
	pois := planetPOIs[pt]

	data := PlanetScanData{
		Name:       obj.Name,
		SystemName: systemName,
		PlanetType: pt,
		Resources:  resources[rng.IntN(len(resources))],
		Hazard:     hazards[rng.IntN(len(hazards))],
	}

	// 50% chance of a point of interest
	if rng.IntN(2) == 0 {
		data.POI = pois[rng.IntN(len(pois))]
	}

	return data
}
