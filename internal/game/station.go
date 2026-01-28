package game

import (
	"fmt"
	"math/rand/v2"
)

// CargoKind identifies a type of tradeable cargo.
type CargoKind uint8

const (
	CargoNone CargoKind = iota // empty pad
	CargoScrapMetal
	CargoWaterIce
	CargoRationPacks
	CargoPowerCells
	CargoMedKits
	CargoCircuitry
	CargoRareMinerals
	CargoAlienArtifacts
	// Prologue-specific repair items
	CargoShuttleFuel   // fuel cells for shuttle tank
	CargoSpareParts    // engine repair parts
	CargoShuttlePower  // power pack for battery
	CargoKindCount     // sentinel — not a real cargo type
)

// MaxPerPad is the stack limit per cargo pad.
const MaxPerPad = 10

// CargoPad is a single cargo bay pad holding a stack of one kind.
type CargoPad struct {
	Kind  CargoKind
	Count int
}

// cargoEntry holds static info about a cargo type.
type cargoEntry struct {
	Name      string
	BasePrice int
}

// cargoTable maps CargoKind to name and base price.
var cargoTable = [CargoKindCount]cargoEntry{
	CargoNone:           {"Empty", 0},
	CargoScrapMetal:     {"Scrap Metal", 3},
	CargoWaterIce:       {"Water Ice", 5},
	CargoRationPacks:    {"Ration Packs", 8},
	CargoPowerCells:     {"Power Cells", 12},
	CargoMedKits:        {"Med Kits", 18},
	CargoCircuitry:      {"Circuitry", 25},
	CargoRareMinerals:   {"Rare Minerals", 35},
	CargoAlienArtifacts: {"Alien Artifacts", 50},
	CargoShuttleFuel:    {"Shuttle Fuel", 20},
	CargoSpareParts:     {"Spare Parts", 25},
	CargoShuttlePower:   {"Power Pack", 15},
}

// CargoName returns the display name for a cargo kind.
func CargoName(k CargoKind) string {
	if k < CargoKindCount {
		return cargoTable[k].Name
	}
	return "Unknown"
}

// CargoBasePrice returns the base price for a cargo kind.
func CargoBasePrice(k CargoKind) int {
	if k < CargoKindCount {
		return cargoTable[k].BasePrice
	}
	return 0
}

// StationData is the generated state of a single station.
// Created lazily on first dock, persists for the session.
type StationData struct {
	Name       string
	Tagline    string
	SellPrices [CargoKindCount]int  // what station charges player to buy
	BuyPrices  [CargoKindCount]int  // what station pays player to sell
	Stock      [CargoKindCount]int  // units available at station
	Stocked    [CargoKindCount]bool // which types this station carries
	BarScene   string               // random bar text (generated on dock)
	Faction    string               // faction presence at this station
}

// StockedList returns the cargo kinds this station carries, in order.
func (sd *StationData) StockedList() []CargoKind {
	var list []CargoKind
	for k := CargoKind(1); k < CargoKindCount; k++ {
		if sd.Stocked[k] {
			list = append(list, k)
		}
	}
	return list
}

// GenerateStationData creates station data from a seed and name.
func GenerateStationData(seed int64, name string) *StationData {
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>16|3)))

	sd := &StationData{
		Name:    name,
		Faction: "Space Knights", // placeholder — all stations Space Knights for now
	}

	// Tagline
	taglines := []string{
		"Where the recycled air is almost breathable.",
		"Fuel up. Stock up. Try not to blow up.",
		"Now with 40% fewer hull breaches!",
		"We put the 'station' in 'desperation'.",
		"Voted 'Adequate' three years running.",
		"Free docking. Everything else costs extra.",
		"Home is wherever you can afford to stop.",
		"Our motto: it could be worse.",
	}
	sd.Tagline = taglines[rng.IntN(len(taglines))]

	// Stock 4-6 cargo types (skip CargoNone at index 0)
	numStocked := 4 + rng.IntN(3)
	// Shuffle cargo kinds to pick which ones are stocked
	kinds := make([]CargoKind, 0, int(CargoKindCount)-1)
	for k := CargoKind(1); k < CargoKindCount; k++ {
		kinds = append(kinds, k)
	}
	rng.Shuffle(len(kinds), func(i, j int) {
		kinds[i], kinds[j] = kinds[j], kinds[i]
	})
	if numStocked > len(kinds) {
		numStocked = len(kinds)
	}
	for i := 0; i < numStocked; i++ {
		k := kinds[i]
		sd.Stocked[k] = true
		// Price modifier: 0.8 to 1.4
		modifier := 0.8 + rng.Float64()*0.6
		sellPrice := int(float64(cargoTable[k].BasePrice)*modifier + 0.5)
		if sellPrice < 1 {
			sellPrice = 1
		}
		sd.SellPrices[k] = sellPrice
		sd.BuyPrices[k] = int(float64(sellPrice)*0.7 + 0.5)
		if sd.BuyPrices[k] < 1 {
			sd.BuyPrices[k] = 1
		}
		// Stock 3-10 units
		sd.Stock[k] = 3 + rng.IntN(8)
	}

	// Generate bar scene
	sd.BarScene = generateBarScene(rng)

	return sd
}

func generateBarScene(rng *rand.Rand) string {
	scenes := []string{
		"The bartender slides you something luminous.\nIt might be a drink. It might be alive.",
		"A patron at the end of the bar is arguing\nwith a potted plant. The plant is winning.",
		"The music is best described as 'aggressive\nambiance.' Nobody seems to mind.",
		"A trading crew plays cards in the corner.\nOne of them is clearly cheating.\nThe others don't seem to care.",
		"The bar smells like recycled air and\nbroken promises. The drinks are worse.",
		"A retired pilot tells you about the time\nshe outran a pirate fleet. You suspect\nshe's lying. She's definitely lying.",
	}

	rumors := []string{
		"A patron whispers: \"Heard there's good\nsalvage near the outer systems.\"",
		"Someone mutters: \"The Space Knights are\nrecruiting. Desperate times.\"",
		"A trader mentions: \"Circuitry prices are\nthrough the roof at the rim stations.\"",
		"An old spacer rasps: \"The Monkey Lion.\nThey say she's still out there.\"",
		"A drunk pilot slurs: \"Don't fly near\nthe red giants. Trust me on this.\"",
		"A mechanic sighs: \"Parts are getting\nscarce. Stock up while you can.\"",
	}

	deborahs := []string{
		"Deborah tries to eat the barstool.\nShe is a zebra.",
		"Deborah is wearing a tiny party hat.\nNobody knows where she got it.",
		"Deborah stares at you with an intensity\nthat suggests she thinks you're food.",
		"Deborah has somehow gotten behind the\nbar. The bartender has given up.",
		"Deborah is asleep on the pool table.\nThe game continues around her.",
	}

	scene := scenes[rng.IntN(len(scenes))]
	rumor := rumors[rng.IntN(len(rumors))]
	deborah := deborahs[rng.IntN(len(deborahs))]

	return fmt.Sprintf("%s\n\n%s\n\n%s", scene, rumor, deborah)
}

// DockRefill performs the auto-refill sequence when docking at a station.
// Flushes waste, converts dirty to clean, tops off all tanks and energy.
func DockRefill(r *Resources) {
	// 1. Flush body waste → dirty pools
	r.Organic.Dirty += r.WasteOrganic
	r.Water.Dirty += r.WasteWater
	r.WasteOrganic = 0
	r.WasteWater = 0

	// 2. Convert all dirty → clean
	r.Organic.Clean += r.Organic.Dirty
	r.Organic.Dirty = 0
	r.Water.Clean += r.Water.Dirty
	r.Water.Dirty = 0

	// 3. Clear recycler buffer → clean
	r.Organic.Clean += r.Recycler.OrganicBuffer
	r.Water.Clean += r.Recycler.WaterBuffer
	r.Recycler.OrganicBuffer = 0
	r.Recycler.WaterBuffer = 0

	// 4. Top off tanks to capacity (station provides the extra)
	r.Organic.Clean = r.Organic.Capacity
	r.Water.Clean = r.Water.Capacity

	// 5. Refill energy
	r.Energy = r.MaxEnergy
}
