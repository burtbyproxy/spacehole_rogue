package game

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

	// Player body — matter "in transit" through the player.
	// Eating/drinking moves CLEAN matter from ship → body.
	// Body slowly processes clean → dirty (waste).
	// Toilet flushes waste → DIRTY matter back to ship.
	BodyOrganic  int // food being digested (clean, in body)
	BodyWater    int // water being processed (clean, in body)
	WasteOrganic int // digested food (dirty, needs toilet)
	WasteWater   int // processed water (dirty, needs toilet)
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
		// Cryo aftermath: body full of waste, need the toilet
		WasteOrganic: 10,
		WasteWater:   5,
	}
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
