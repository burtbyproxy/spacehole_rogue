package game

// ItemKind identifies a type of personal inventory item.
// These are small items the player carries, separate from ship cargo.
type ItemKind uint8

const (
	ItemNone ItemKind = iota
	// Repair supplies
	ItemFuelCells   // fuel cells for shuttle tank
	ItemSpareParts  // engine repair parts
	ItemPowerPack   // power pack for battery
	// Consumables
	ItemRationPack  // emergency food
	ItemWaterPack   // emergency water
	ItemMedKit      // heals injuries
	// Utility
	ItemToolKit     // repairs equipment
	ItemScanner     // improves scan range
	ItemKindCount   // sentinel
)

// MaxInventorySlots is the personal inventory limit.
const MaxInventorySlots = 6

// InventorySlot holds one item stack.
type InventorySlot struct {
	Kind  ItemKind
	Count int
}

// Inventory is the player's personal item storage.
type Inventory struct {
	Slots [MaxInventorySlots]InventorySlot
}

// NewInventory creates an empty personal inventory.
func NewInventory() Inventory {
	return Inventory{}
}

// FindSlot returns the index of the slot holding the given kind, or -1.
func (inv *Inventory) FindSlot(kind ItemKind) int {
	for i, slot := range inv.Slots {
		if slot.Kind == kind && slot.Count > 0 {
			return i
		}
	}
	return -1
}

// FindEmptySlot returns the index of the first empty slot, or -1.
func (inv *Inventory) FindEmptySlot() int {
	for i, slot := range inv.Slots {
		if slot.Kind == ItemNone || slot.Count == 0 {
			return i
		}
	}
	return -1
}

// HasItem returns true if the inventory contains at least one of the given item.
func (inv *Inventory) HasItem(kind ItemKind) bool {
	return inv.FindSlot(kind) >= 0
}

// AddItem adds an item to the inventory. Returns true if successful.
func (inv *Inventory) AddItem(kind ItemKind, amount int) bool {
	// Try to stack with existing
	idx := inv.FindSlot(kind)
	if idx >= 0 {
		inv.Slots[idx].Count += amount
		return true
	}
	// Try to claim empty slot
	idx = inv.FindEmptySlot()
	if idx < 0 {
		return false // inventory full
	}
	inv.Slots[idx].Kind = kind
	inv.Slots[idx].Count = amount
	return true
}

// RemoveItem removes an item from the inventory. Returns true if successful.
func (inv *Inventory) RemoveItem(kind ItemKind, amount int) bool {
	idx := inv.FindSlot(kind)
	if idx < 0 {
		return false
	}
	if inv.Slots[idx].Count < amount {
		return false
	}
	inv.Slots[idx].Count -= amount
	if inv.Slots[idx].Count == 0 {
		inv.Slots[idx].Kind = ItemNone
	}
	return true
}

// Count returns the number of a specific item type.
func (inv *Inventory) Count(kind ItemKind) int {
	idx := inv.FindSlot(kind)
	if idx < 0 {
		return 0
	}
	return inv.Slots[idx].Count
}

// UsedSlots returns the number of non-empty slots.
func (inv *Inventory) UsedSlots() int {
	n := 0
	for _, slot := range inv.Slots {
		if slot.Kind != ItemNone && slot.Count > 0 {
			n++
		}
	}
	return n
}

// itemEntry holds static info about an item type.
type itemEntry struct {
	Name string
}

var itemTable = [ItemKindCount]itemEntry{
	ItemNone:       {"Empty"},
	ItemFuelCells:  {"Fuel Cells"},
	ItemSpareParts: {"Spare Parts"},
	ItemPowerPack:  {"Power Pack"},
	ItemRationPack: {"Ration Pack"},
	ItemWaterPack:  {"Water Pack"},
	ItemMedKit:     {"Med Kit"},
	ItemToolKit:    {"Tool Kit"},
	ItemScanner:    {"Scanner"},
}

// ItemName returns the display name for an item kind.
func ItemName(k ItemKind) string {
	if k < ItemKindCount {
		return itemTable[k].Name
	}
	return "Unknown"
}
