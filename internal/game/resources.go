package game

import "github.com/spacehole-rogue/spacehole_rogue/internal/world"

// MatterPool tracks a finite resource that cycles between clean and dirty states.
// Clean + Dirty + any matter held outside the system = Capacity.
type MatterPool struct {
	Clean    int
	Dirty    int
	Capacity int // max total matter in the system
}

// Total returns the amount of matter currently in the ship's systems.
func (m *MatterPool) Total() int { return m.Clean + m.Dirty }

// Free returns capacity not accounted for (matter held by player, lost to leaks, etc.)
func (m *MatterPool) Free() int { return m.Capacity - m.Clean - m.Dirty }

// MaxBodyFullness is the max matter the player's body can hold.
const MaxBodyFullness = 30

// RecyclerState tracks the internal buffer of the combined matter recycler.
// Dirty matter is pulled from ship pools into the buffer, then processed into clean.
type RecyclerState struct {
	WaterBuffer   int // dirty water waiting to be processed
	OrganicBuffer int // dirty organic waiting to be processed
	Capacity      int // max per type in buffer
}

// Resources tracks all matter and energy on the shuttle.
type Resources struct {
	Water   MatterPool
	Organic MatterPool
	Energy    int
	MaxEnergy int
	Hull    int
	MaxHull int
	Recycler RecyclerState

	// Jump fuel — separate from energy, used for FTL jumps
	// Incinerator converts matter → fuel
	JumpFuel    int
	MaxJumpFuel int

	// Player body — matter "in transit" through the player.
	// Eating/drinking moves CLEAN matter from ship → body.
	// Body slowly processes clean → dirty (waste).
	// Toilet flushes waste → DIRTY matter back to ship.
	BodyOrganic  int // food being digested (clean, in body)
	BodyWater    int // water being processed (clean, in body)
	WasteOrganic int // digested food (dirty, needs toilet)
	WasteWater   int // processed water (dirty, needs toilet)

	// Economy
	Credits   int        // currency
	CargoPads []CargoPad // cargo bay — each pad holds a stack of one kind

	// Personal inventory — small items the player carries
	Inventory Inventory
}

// BodyFullness returns total matter in the player's body.
func (r *Resources) BodyFullness() int {
	return r.BodyOrganic + r.BodyWater + r.WasteOrganic + r.WasteWater
}

// TotalWaste returns dirty matter in the player ready to flush.
func (r *Resources) TotalWaste() int {
	return r.WasteOrganic + r.WasteWater
}

// NewShuttleResources creates the starting resource state.
// You just woke from cryo — you've got some waste to deal with.
// Matter is conserved: Water.Clean + Water.Dirty + BodyWater + WasteWater = Water.Capacity
func NewShuttleResources() Resources {
	return Resources{
		Water:    MatterPool{Clean: 78, Dirty: 17, Capacity: 100},
		Organic:  MatterPool{Clean: 55, Dirty: 35, Capacity: 100},
		Energy:   95,
		MaxEnergy: 100,
		Hull:     100,
		MaxHull:  100,
		Recycler: RecyclerState{Capacity: 5},
		// Jump fuel — starts with just enough for one jump
		JumpFuel:    100, // one jump costs ~90
		MaxJumpFuel: 100,
		// Cryo aftermath: body full of waste, need the toilet
		WasteOrganic: 10,
		WasteWater:   5,
		// Economy
		Credits:   100,
		CargoPads: make([]CargoPad, 12), // 12 cargo pads in shuttle
	}
}

// CargoCount returns total cargo units across all pads.
func (r *Resources) CargoCount() int {
	n := 0
	for _, p := range r.CargoPads {
		n += p.Count
	}
	return n
}

// PadsUsed returns the number of non-empty cargo pads.
func (r *Resources) PadsUsed() int {
	n := 0
	for _, p := range r.CargoPads {
		if p.Kind != CargoNone {
			n++
		}
	}
	return n
}

// FindPad returns the index of the pad holding the given kind, or -1.
func (r *Resources) FindPad(kind CargoKind) int {
	for i, p := range r.CargoPads {
		if p.Kind == kind && p.Count > 0 {
			return i
		}
	}
	return -1
}

// FindEmptyPad returns the index of the first empty pad, or -1.
func (r *Resources) FindEmptyPad() int {
	for i, p := range r.CargoPads {
		if p.Kind == CargoNone {
			return i
		}
	}
	return -1
}

// AddCargo adds units of a cargo kind to the bay.
// Uses an existing stack or claims an empty pad. Returns amount actually added.
func (r *Resources) AddCargo(kind CargoKind, amount int) int {
	idx := r.FindPad(kind)
	if idx < 0 {
		idx = r.FindEmptyPad()
		if idx < 0 {
			return 0 // no room
		}
		r.CargoPads[idx].Kind = kind
	}
	pad := &r.CargoPads[idx]
	space := MaxPerPad - pad.Count
	if amount > space {
		amount = space
	}
	pad.Count += amount
	return amount
}

// RemoveCargo removes units of a cargo kind. Returns amount actually removed.
func (r *Resources) RemoveCargo(kind CargoKind, amount int) int {
	idx := r.FindPad(kind)
	if idx < 0 {
		return 0
	}
	pad := &r.CargoPads[idx]
	if amount > pad.Count {
		amount = pad.Count
	}
	pad.Count -= amount
	if pad.Count == 0 {
		pad.Kind = CargoNone // free the pad
	}
	return amount
}

func (r *Resources) EnergyPct() int { return r.Energy * 100 / r.MaxEnergy }
func (r *Resources) HullPct() int   { return r.Hull * 100 / r.MaxHull }

// PlayerNeeds tracks the player's bodily state.
type PlayerNeeds struct {
	Hunger  int // 0 = full, 100 = starving
	Thirst  int // 0 = hydrated, 100 = dehydrated
	Hygiene int // 0 = clean, 100 = filthy
}

// NeedLevel returns a human-readable label for a need value.
func NeedLevel(val int) string {
	switch {
	case val <= 10:
		return "OK"
	case val <= 30:
		return "Mild"
	case val <= 60:
		return "Moderate"
	case val <= 80:
		return "Severe"
	default:
		return "CRITICAL"
	}
}

// MatterType identifies a type of matter that flows through ship systems.
type MatterType uint8

const (
	MatterNone    MatterType = iota
	MatterPower              // energy/electricity
	MatterWater              // H2O
	MatterOrganic            // food/biological matter
	MatterFuel               // jump drive fuel
)

// MatterName returns a human-readable name for a matter type.
func MatterName(m MatterType) string {
	switch m {
	case MatterPower:
		return "power"
	case MatterWater:
		return "water"
	case MatterOrganic:
		return "organics"
	case MatterFuel:
		return "fuel"
	default:
		return "unknown"
	}
}

// ComponentType identifies the role of equipment in the matter flow system.
type ComponentType uint8

const (
	CompNone      ComponentType = iota
	CompTank                    // stores matter
	CompGenerator               // produces matter (uses power)
	CompRecycler                // converts dirty → clean (same matter type)
	CompOutput                  // dispenses clean matter to player
	CompInput                   // accepts matter from player (waste)
	CompConverter               // converts one matter type to another
)

// EquipmentMatterInfo describes how a piece of equipment interacts with matter.
type EquipmentMatterInfo struct {
	Component  ComponentType
	InputType  MatterType // matter consumed (or MatterNone)
	OutputType MatterType // matter produced (or MatterNone)
}

// EquipmentMatter maps equipment to its matter flow behavior.
// InputType = what it consumes, OutputType = what it produces
var EquipmentMatter = map[world.EquipmentKind]EquipmentMatterInfo{
	// Tanks - store a specific matter type
	world.EquipPowerCell:   {CompTank, MatterNone, MatterPower},
	world.EquipWaterTank:   {CompTank, MatterNone, MatterWater},
	world.EquipOrganicTank: {CompTank, MatterNone, MatterOrganic},
	world.EquipFuelTank:    {CompTank, MatterNone, MatterFuel},

	// Generator - produces power (consumes nothing, just needs to be ON)
	world.EquipGenerator: {CompGenerator, MatterNone, MatterPower},

	// Recycler - converts dirty → clean (handles both water and organic)
	// Special case: processes multiple matter types, dirty→clean conversion
	world.EquipMatterRecycler: {CompRecycler, MatterNone, MatterNone},

	// Outputs (replicators) - dispense clean matter from tanks to player
	// InputType = what they draw from tanks
	world.EquipFoodStation:  {CompOutput, MatterOrganic, MatterNone},
	world.EquipDrinkStation: {CompOutput, MatterWater, MatterNone},

	// Inputs - accept matter from player, output to dirty pools
	// Toilet: waste from body → dirty water + organic
	// Shower: clean water → dirty water (hygiene)
	world.EquipToilet: {CompInput, MatterNone, MatterNone}, // accepts both waste types
	world.EquipShower: {CompInput, MatterWater, MatterWater}, // clean water in, dirty water out

	// Converters - transform one matter type to another
	// Incinerator: burns cargo (organic) → produces fuel
	world.EquipIncinerator: {CompConverter, MatterOrganic, MatterFuel},
}

// PackFillAmount is how much a single pack replenishes.
// Small amount since these are carryable personal items.
const PackFillAmount = 5

// TankMatterType returns the matter type a tank holds, or MatterNone.
func TankMatterType(eq world.EquipmentKind) MatterType {
	if info, ok := EquipmentMatter[eq]; ok && info.Component == CompTank {
		return info.OutputType
	}
	return MatterNone
}

// OutputMatterType returns the matter type an output dispenses, or MatterNone.
func OutputMatterType(eq world.EquipmentKind) MatterType {
	if info, ok := EquipmentMatter[eq]; ok && info.Component == CompOutput {
		return info.InputType // output dispenses its input type from tanks
	}
	return MatterNone
}

// ConverterMatter returns the input and output matter types for a converter.
func ConverterMatter(eq world.EquipmentKind) (input, output MatterType) {
	if info, ok := EquipmentMatter[eq]; ok && info.Component == CompConverter {
		return info.InputType, info.OutputType
	}
	return MatterNone, MatterNone
}

// PackMatterType returns the matter type a pack contains, or MatterNone.
func PackMatterType(item ItemKind) MatterType {
	switch item {
	case ItemPowerPack:
		return MatterPower
	case ItemWaterPack:
		return MatterWater
	case ItemRationPack:
		return MatterOrganic
	case ItemFuelCells:
		return MatterFuel
	default:
		return MatterNone
	}
}

// PackForMatter returns the pack item for a given matter type.
func PackForMatter(m MatterType) ItemKind {
	switch m {
	case MatterPower:
		return ItemPowerPack
	case MatterWater:
		return ItemWaterPack
	case MatterOrganic:
		return ItemRationPack
	case MatterFuel:
		return ItemFuelCells
	default:
		return ItemNone
	}
}

// TryFillTank attempts to use a pack from inventory to fill the given tank.
// Matches pack matter type to tank matter type automatically.
// Returns (filled amount, pack name) if successful, (0, "") if not.
func (r *Resources) TryFillTank(eq world.EquipmentKind) (int, string) {
	matterType := TankMatterType(eq)
	if matterType == MatterNone {
		return 0, ""
	}

	pack := PackForMatter(matterType)
	if pack == ItemNone || !r.Inventory.HasItem(pack) {
		return 0, ""
	}

	var filled int
	switch matterType {
	case MatterPower:
		space := r.MaxEnergy - r.Energy
		filled = min(PackFillAmount, space)
		r.Energy += filled
	case MatterWater:
		space := r.Water.Capacity - r.Water.Clean - r.Water.Dirty
		filled = min(PackFillAmount, space)
		r.Water.Clean += filled
	case MatterOrganic:
		space := r.Organic.Capacity - r.Organic.Clean - r.Organic.Dirty
		filled = min(PackFillAmount, space)
		r.Organic.Clean += filled
	case MatterFuel:
		space := r.MaxJumpFuel - r.JumpFuel
		filled = min(PackFillAmount, space)
		r.JumpFuel += filled
	default:
		return 0, ""
	}

	if filled > 0 {
		r.Inventory.RemoveItem(pack, 1)
	}
	return filled, ItemName(pack)
}
