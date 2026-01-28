package world

// PowerMode defines how equipment consumes energy.
type PowerMode uint8

const (
	PowerNone     PowerMode = iota // no power required
	PowerConstant                  // drains energy every tick while ON
	PowerOnUse                     // drains energy only when used/activated
)

// Equipment is an instance of a piece of ship equipment.
// Each instance has its own state that can be upgraded or degraded.
type Equipment struct {
	Kind EquipmentKind
	On   bool // is it powered on?

	// Power characteristics (can be modified by upgrades/damage)
	PowerMode PowerMode
	PowerCost int // energy per tick (constant) or per use (on-use)

	// Condition affects efficiency (0-100, starts at 100)
	Condition int

	// Efficiency multiplier (1.0 = normal, >1 = upgraded, <1 = degraded)
	Efficiency float64

	// Future: upgrades, damage history, etc.
}

// EquipmentTemplate defines the base stats for an equipment type.
type EquipmentTemplate struct {
	Kind       EquipmentKind
	PowerMode  PowerMode
	PowerCost  int
	Efficiency float64
}

// EquipmentTemplates defines base stats for each equipment kind.
var EquipmentTemplates = map[EquipmentKind]EquipmentTemplate{
	// --- No power required ---
	EquipNone:       {EquipNone, PowerNone, 0, 1.0},
	EquipBed:        {EquipBed, PowerNone, 0, 1.0},
	EquipLocker:     {EquipLocker, PowerNone, 0, 1.0},
	EquipCargoTile:  {EquipCargoTile, PowerNone, 0, 1.0},
	EquipOrganicTank: {EquipOrganicTank, PowerNone, 0, 1.0},
	EquipWaterTank:   {EquipWaterTank, PowerNone, 0, 1.0},
	EquipFuelTank:    {EquipFuelTank, PowerNone, 0, 1.0},
	EquipPowerCell:   {EquipPowerCell, PowerNone, 0, 1.0},

	// --- Constant draw (per tick while ON) ---
	EquipEngine:          {EquipEngine, PowerConstant, 2, 1.0},
	EquipGenerator:       {EquipGenerator, PowerConstant, 1, 1.0},
	EquipMatterRecycler:  {EquipMatterRecycler, PowerConstant, 1, 1.0},
	EquipCargoTransporter: {EquipCargoTransporter, PowerConstant, 1, 1.0},

	// --- On-use (per interaction) ---
	EquipViewscreen:     {EquipViewscreen, PowerOnUse, 1, 1.0},
	EquipNavConsole:     {EquipNavConsole, PowerOnUse, 2, 1.0},
	EquipPilotConsole:   {EquipPilotConsole, PowerOnUse, 2, 1.0},
	EquipScienceConsole: {EquipScienceConsole, PowerOnUse, 2, 1.0},
	EquipCargoConsole:   {EquipCargoConsole, PowerOnUse, 1, 1.0},
	EquipMedical:        {EquipMedical, PowerOnUse, 3, 1.0},
	EquipFoodStation:    {EquipFoodStation, PowerOnUse, 2, 1.0},
	EquipDrinkStation:   {EquipDrinkStation, PowerOnUse, 1, 1.0},
	EquipIncinerator:    {EquipIncinerator, PowerOnUse, 2, 1.0},
	EquipJumpDrive:      {EquipJumpDrive, PowerOnUse, 90, 1.0}, // massive power draw - shut systems down to jump!
	EquipToilet:         {EquipToilet, PowerNone, 0, 1.0},      // gravity-fed, no power needed
	EquipShower:         {EquipShower, PowerOnUse, 1, 1.0},

	// --- Surface equipment (no ship power) ---
	EquipTerminal:   {EquipTerminal, PowerNone, 0, 1.0},
	EquipLootCrate:  {EquipLootCrate, PowerNone, 0, 1.0},
	EquipObjective:  {EquipObjective, PowerNone, 0, 1.0},
	EquipFuelCell:   {EquipFuelCell, PowerNone, 0, 1.0},
	EquipSpareParts: {EquipSpareParts, PowerNone, 0, 1.0},
	EquipPowerPack:  {EquipPowerPack, PowerNone, 0, 1.0},
}

// NewEquipment creates a new equipment instance from a template.
func NewEquipment(kind EquipmentKind) *Equipment {
	template, ok := EquipmentTemplates[kind]
	if !ok {
		template = EquipmentTemplate{kind, PowerNone, 0, 1.0}
	}
	return &Equipment{
		Kind:       kind,
		On:         false,
		PowerMode:  template.PowerMode,
		PowerCost:  template.PowerCost,
		Condition:  100,
		Efficiency: template.Efficiency,
	}
}

// TryDrawPower attempts to draw power for this equipment.
// Returns true if power was drawn (or not needed), false if insufficient power.
func (e *Equipment) TryDrawPower(available *int) bool {
	if e.PowerMode != PowerConstant || e.PowerCost == 0 {
		return true // no constant power needed
	}
	if *available >= e.PowerCost {
		*available -= e.PowerCost
		return true
	}
	return false
}

// CanUse checks if there's enough power for an on-use action.
func (e *Equipment) CanUse(available int) bool {
	if e.PowerMode != PowerOnUse || e.PowerCost == 0 {
		return true
	}
	return available >= e.PowerCost
}

// Use deducts power for an on-use action. Call CanUse first!
func (e *Equipment) Use(available *int) {
	if e.PowerMode == PowerOnUse && e.PowerCost > 0 {
		*available -= e.PowerCost
	}
}

// Degrade reduces condition by amount. Returns new condition.
func (e *Equipment) Degrade(amount int) int {
	e.Condition = max(0, e.Condition-amount)
	// Efficiency drops as condition drops
	if e.Condition < 50 {
		e.Efficiency = 0.5 + float64(e.Condition)/100.0
	}
	return e.Condition
}

// Repair increases condition by amount. Returns new condition.
func (e *Equipment) Repair(amount int) int {
	e.Condition = min(100, e.Condition+amount)
	// Restore efficiency
	if e.Condition >= 50 {
		e.Efficiency = 1.0
	} else {
		e.Efficiency = 0.5 + float64(e.Condition)/100.0
	}
	return e.Condition
}

// Name returns human-readable name for this equipment.
func (e *Equipment) Name() string {
	return equipmentNames[e.Kind]
}

var equipmentNames = map[EquipmentKind]string{
	EquipEngine:          "Engine",
	EquipGenerator:       "Generator",
	EquipMatterRecycler:  "Recycler",
	EquipCargoTransporter: "Cargo Transporter",
	EquipViewscreen:      "Viewscreen",
	EquipNavConsole:      "Nav Console",
	EquipPilotConsole:    "Pilot Console",
	EquipScienceConsole:  "Science Console",
	EquipCargoConsole:    "Cargo Console",
	EquipMedical:         "Medical Station",
	EquipFoodStation:     "Food Replicator",
	EquipDrinkStation:    "Drink Replicator",
	EquipIncinerator:     "Incinerator",
	EquipToilet:          "Toilet",
	EquipShower:          "Shower",
	EquipBed:             "Bed",
	EquipLocker:          "Locker",
	EquipOrganicTank:     "Organic Tank",
	EquipWaterTank:       "Water Tank",
	EquipFuelTank:        "Fuel Tank",
	EquipJumpDrive:       "Jump Drive",
	EquipPowerCell:       "Battery",
	EquipCargoTile:       "Cargo Pad",
	EquipTerminal:        "Terminal",
	EquipLootCrate:       "Loot Crate",
	EquipObjective:       "Objective",
	EquipFuelCell:        "Fuel Cells",
	EquipSpareParts:      "Spare Parts",
	EquipPowerPack:       "Power Pack",
}
