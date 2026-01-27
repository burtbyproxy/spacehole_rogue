# Phase 7: Sector Map + Navigation

## Summary

Player sits at navigation console (=), presses E → sector map opens. Stars shown on 80x45 grid. WASD snaps cursor between stars. E confirms jump (costs energy). Instant travel, back to ship interior. ESC returns to ship view without jumping.

---

## Files to Change

| File | Action | What |
|------|--------|------|
| `internal/game/sector.go` | NEW | StarType, StarSystem, Sector, NewSector(seed), NearestInDirection, EnergyCostTo |
| `internal/game/sim.go` | MODIFY | Add `Sector *Sector` + `ConsoleActivated bool` fields, modify console interact, add `NavigateTo()` |
| `cmd/spacehole/main.go` | MODIFY | Add ViewMode enum, split Update/drawScreen into ship/sector variants, add sector map rendering |

No changes to `render/`, `world/`, `resources.go`, `messages.go`, or `assets`.

---

## Step 1: internal/game/sector.go (new)

```go
// StarType enum
type StarType uint8
const (
    StarYellow StarType = iota
    StarRed
    StarBlue
    StarWhite
    StarOrange
)

// StarSystem struct
type StarSystem struct {
    Name    string
    X, Y    int
    Type    StarType
    Visited bool
}

// Sector struct
type Sector struct {
    Systems       []StarSystem
    CurrentSystem int
    CursorSystem  int
    Seed          int64
}
```

**NewSector(seed)** — seeded PRNG generation:
- 12-15 star systems
- Placed in [4,54] x [4,34] with min 4-cell spacing
- Starting system near center, marked Visited
- Names from preset list, shuffled by seed
- Random star types

**NearestInDirection(dx, dy) int** — snap cursor to nearest star in pressed direction

**EnergyCostTo(target) int** — distance-based, ~int(dist * 1.5), minimum 5

**DistanceBetween(a, b) float64** — Euclidean

---

## Step 2: internal/game/sim.go (modify)

- Add `Sector *Sector` and `ConsoleActivated bool` to Sim struct
- `NewSim()`: create sector with `NewSector(42)`
- `Interact()` console case: set `ConsoleActivated = true` instead of info message
- New `NavigateTo(targetIdx int) bool`: check energy, deduct cost, update CurrentSystem, set Visited, log MsgDiscovery

---

## Step 3: cmd/spacehole/main.go (modify)

**Add ViewMode enum:**
```go
type ViewMode uint8
const (
    ViewShip ViewMode = iota
    ViewSectorMap
)
```

**Split Update():**
- `updateShip()` + `updateSectorMap()` with switch
- Check `ConsoleActivated` → switch to ViewSectorMap

**Split drawScreen():**
- `drawShipView()` + `drawSectorMapView()` with switch

**updateSectorMap():**
- WASD moves cursor via `NearestInDirection`
- E calls `NavigateTo` → switch back to ViewShip
- ESC → back to ViewShip (no quit)

**drawSectorMapView():**
- Clear buffer
- Draw stars as `*` (colored by type)
- Shuttle as `^` at current system
- Cursor brackets `[*]` in yellow
- Star names in dark gray (selected name in white)
- Right panel with selected star info + energy cost
- Bottom: energy bar + comms + instructions

**Helper funcs:**
- `starColor(StarType) uint8`
- `starTypeName(StarType) string`

---

## Key Design Decisions

- **View state on Game, not Sim** — Sim is simulation, Game is UI. Sim doesn't need to know what's displayed.
- **Cursor snaps between stars** — no free movement on the grid, WASD finds nearest star in that direction
- **Instant travel** — no travel time for Phase 7, just deduct energy and switch view
- **Sim ticks during map view** — resources drain, needs rise. Creates urgency.
- **Energy = fuel** — no separate fuel resource yet. Typical jump costs 15-30 energy.
- **Star names only shown for selected star** — avoids overlap issues on the map

---

## Verification

- [ ] Walk to console (=), press E → sector map appears
- [ ] Stars visible as colored `*`, shuttle as `^` at starting system
- [ ] WASD snaps cursor between stars, info panel updates
- [ ] E on distant star → energy deducted, "Arrived at X" message, back to ship interior
- [ ] Star marked as visited (lighter color on revisit)
- [ ] E on current system → nothing happens (already here)
- [ ] Not enough energy → warning message, no jump
- [ ] ESC → back to ship interior without jumping
- [ ] Resources still tick during map view
