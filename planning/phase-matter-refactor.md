# Phase: Matter System Refactor + Pipe Infrastructure

## Summary

Four-step implementation:

1. **Matter refactor** — Split Replicator into dispensers + combined recycler with buffer tanks
2. **Pipe infrastructure** — Add pipe/wire data to tiles, SimCity-style overlay toggle to see them
3. **Breakage + leaks** — Pipes/equipment can break, leaking matter onto floors, player repairs
4. **Visual polish** — Tank brightness by fullness, pipe brightness by flow

---

## Step 1: Matter Refactor (Dispensers + Combined Recycler)

### New Matter Flow

```
CLEAN TANKS ──→ DISPENSERS ──→ PLAYER BODY ──→ WASTE
  ↑                              ↓
  └── MATTER RECYCLER ←── DIRTY POOLS ←── TOILET/SHOWER
```

- **Food Station:** vending machine for food. E to eat. 5 clean organic → body, -35 hunger.
- **Drink Station:** vending machine for water. E to drink. 3 clean water → body, -25 thirst.
- **Matter Recycler:** combined unit. Toggleable (T). Has internal buffer per type (capacity 5 each on shuttle). Pulls dirty matter in, processes over time, spits clean matter out. Costs 1 energy per unit processed.
- **Tanks (Water, Organic):** info only. No eating/drinking at tanks.

### Equipment Changes

| Remove | Replace With |
|--------|--------------|
| EquipReplicator | EquipFoodStation (food dispenser) |
| EquipWaterRecycler | EquipMatterRecycler (combined recycler) |

| Add | Purpose |
|-----|---------|
| EquipFoodStation | Food vending machine (dispenses clean organic) |
| EquipDrinkStation | Water vending machine (dispenses clean water) |
| EquipMatterRecycler | Combined dirty→clean for water + organic, with buffer tanks |
| EquipGenerator | Power generator (generates energy at fixed rate) |

**Engine no longer generates power — Generator does. Engine provides thrust only.**

### Files: Step 1

**internal/world/tilemap.go** — Replace EquipReplicator + EquipWaterRecycler with EquipFoodStation, EquipDrinkStation, EquipMatterRecycler. Update descriptions.

**internal/world/loader.go** — Remove 'R', 'w'. Add 'F'→FoodStation, 'D'→DrinkStation, 'r'→MatterRecycler, 'g'→Generator.

**internal/game/resources.go** — Add:

```go
type RecyclerState struct {
    WaterBuffer   int
    OrganicBuffer int
    Capacity      int // max per type
}
```

Add `Recycler RecyclerState` to Resources. Init with `Capacity: 5`.

**internal/game/sim.go** — Replace `WaterRecyclerOn` + `ReplicatorOn` with `RecyclerOn` + `GeneratorOn`. Rewrite `tickEngine()` → `tickGenerator()` (generator makes power, not engine). Rewrite `tickRecyclers()` with intake+process phases using recycler buffers. Rewrite `Interact()` and `ToggleEquipment()` for new equipment types.

**internal/render/tiles.go** — New visuals:

| Equipment | Glyph | FG |
|-----------|-------|----|
| FoodStation | F | LightGreen |
| DrinkStation | D | LightCyan |
| MatterRecycler | 177 ▒ | LightMagenta |

**assets/ships/shuttle.json** — New 15×16 layout with drink/food stations, combined recycler, cargo bay.

**cmd/spacehole/main.go** — Update legend + equipment status panels.

---

## Step 2: Pipe Infrastructure + Overlay

### Concept

Like SimCity's underground view: press a key to cycle overlay modes (Normal → Water → Organic → Power). In overlay mode, pipes/wires glow through walls and floors, showing the physical infrastructure. **Never show all pipe types at once — too cluttered.**

### Data Model

Add pipe flags to Tile (bitmask for multiple pipe types per tile):

```go
type PipeFlags uint8
const (
    PipeWater   PipeFlags = 1 << iota // blue
    PipeOrganic                        // green
    PipePower                          // yellow
)
```

Add to Tile struct:

```go
type Tile struct {
    Kind      TileKind
    Equipment EquipmentKind
    Pipes     PipeFlags  // which pipes run through this tile
    Broken    bool       // pipe broken at this tile
}
```

### Layout Format

Second layer in shuttle.json for pipe routing:

```json
"pipes": [
    "...............",
    "...............",
    "...............",
    "######p########",
    ".............p.",
    ".............p.",
    "#####p####p####",
    "wwww#mmm#p....",
    "oooo+mmm+p....",
    "....#mmm#p....",
    "wwww+mmm+p....",
    "wwww#...#p....",
    "#####mmm#p####",
    "ww.mmmmmmgp...",
    ".....p........",
    "..............."
]
```

Characters: `w`=water, `o`=organic, `m`=mixed(water+organic), `p`=power, `g`=organic, `.`=none.

Pipe connections auto-computed from adjacency (check 4 neighbors for matching pipe flags → pick correct box-drawing glyph: `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼`).

### Overlay Rendering

New OverlayMode on Game struct:

```go
type OverlayMode uint8
const (
    OverlayNone OverlayMode = iota
    OverlayWater
    OverlayOrganic
    OverlayPower
)
```

Press Tab (or O) to cycle overlay modes. When active:
- Normal tiles render dimmed (darker FG)
- Pipe tiles matching the active overlay render with bright type color
- Box-drawing glyph auto-selected based on adjacent pipe connections
- Pipe brightness scales with resource flow (resource level / capacity)
- Title bar shows overlay name: "[WATER OVERLAY]" in blue

### Pipe Routing Logic

Equipment implicitly defines pipe endpoints:
- Water Tank → (water pipes) → Recycler → (water pipes) → Drink Station
- Organic Tank → (organic pipes) → Recycler → (organic pipes) → Food Station
- Toilet → (mixed pipes) → Recycler (waste water + waste organic)
- Shower → (water pipes) → Recycler (waste water)
- Engine → (power wires) → Power Cell → (power wires) → Recycler

Pipes manually placed in layout JSON. Auto-tiling computes glyphs.

### Files: Step 2

**internal/world/tilemap.go** — Add PipeFlags, Broken field to Tile.

**internal/world/loader.go** — Parse "pipes" layer from JSON. Apply pipe flags to grid tiles.

**internal/render/tiles.go** — Add `RenderPipeOverlay()` function. Auto-tile pipe glyphs. Color/brightness from resource state.

**cmd/spacehole/main.go** — Add OverlayMode, Tab key to cycle, overlay rendering pass, title bar indicator.

---

## Step 3: Breakage + Leaks + Repair

### Deterioration Mechanics

Pipes and equipment have a condition value (0-100). Condition degrades over time. At 0, the item breaks.

```go
// On Tile struct:
PipeCondition uint8 // 100 = pristine, 0 = broken

// Deterioration rate:
const pipeDeterioration = 1 // lose 1 condition per deterioration tick
const deteriorationInterval = 3000 // check every ~50 real seconds (1 game hour)
```

Each deterioration tick, each pipe/equipment tile loses 1 condition (randomized — not all degrade at the same rate, roll per tile). The player can SEE deterioration in the overlay: pipe brightness/color shifts from green (healthy) → yellow (worn) → red (critical) → broken.

When condition hits 0:
- **Pipe breaks:** matter leaks, alert fires
- **Equipment breaks:** stops functioning, alert fires

### Visual deterioration in overlay

- 100-75: bright pipe color (healthy)
- 74-50: slightly dimmer
- 49-25: yellow tint (warning)
- 24-1: red tint (critical)
- 0: broken glyph (`~`), red, blinking

Equipment and tanks also deteriorate. Tanks at 0 = leak stored matter. Recycler at 0 = stops processing. Engine at 0 = no thrust.

### Leak Visualization

Add `LeakType PipeFlags` to floor tiles near broken pipes. Renderer shows leak color on floor BG:
- Water leak: dark blue BG
- Organic leak: dark green BG
- Both: dark cyan BG

Leaking matter is LOST from the system (reduces pool total). Conservation breaks — this is the threat.

### Repair

Walk to a deteriorated/broken pipe or equipment tile, press E → repair. Restores condition to 100. Costs nothing on shuttle (just time to walk there). Future: repair costs structural matter, requires engineering skill, takes time.

The overlay shows condition at a glance — find the yellow/red pipes and fix them before they break. On the shuttle this is simple (small ship, few pipes). On larger ships, engineering crew handles routine repairs and you get alerts for critical failures.

### Files: Step 3

**internal/world/tilemap.go** — Add LeakType field to Tile.

**internal/game/sim.go** — Add `tickBreakage()` to Tick loop. Breakage rolls, leak processing, leak spread to floor tiles. Add repair case to `Interact()`.

**internal/render/tiles.go** — Broken pipe rendering (glyph `~`, red), leak floor coloring.

---

## Step 4: Visual Polish (Brightness)

### Tank Brightness

Tanks (■ glyph 254) render with dynamic foreground brightness based on fullness:
- Full (>75%): bright color (LightBlue, LightGreen, Yellow)
- Medium (25-75%): normal color (Blue, Green, Brown)
- Low (<25%): dark color (DarkGray tinted)
- Empty: very dark / blinking

This requires passing resource state into the tile renderer. Add a `TileRenderContext` that includes resource levels, which `tileVisuals()` can reference for equipment tiles.

### Pipe Brightness

In overlay mode, pipe segments glow based on flow rate (clean+dirty resource level / capacity):
- High flow: bright type color
- Low flow: dim type color
- Zero flow: very dark (pipe exists but nothing moving)

### Files: Step 4

**internal/render/tiles.go** — Add `TileRenderContext` parameter to `RenderTileGrid()`. Dynamic color selection for tanks and pipes based on resource levels.

**cmd/spacehole/main.go** — Pass resource state into render context.

---

## Shuttle Layout (15 wide × 15 tall)

Matches the reference drawing: large cockpit/living area with drink+food, power room top-right with generator+battery, bedroom and bathroom stacked on left, tall cargo bay on right, utility closet (tanks+recycler) and engine room at bottom.

```
###############    0   width=15
#.VVVVVVV.#gp.#    1   living area (viewscreen) | power room (generator+battery)
#.........+...#    2   living area | door to hallway
#....=....#...#    3   nav console | hallway
#.D.......#...#    4   drink stn   | hallway
#.F.......#...#    5   food stn    | hallway
####+#####++###    6   door(3)→bedroom | doors(10,11)→cargo/hall
#__..#...#cccc#    7   bedroom | connector | cargo bay
#....+...+cccc#    8   bedroom | connector | cargo bay
####+#...#cccc#    9   door(3)→bathroom
#ts..+...+cccc#   10   bathroom | connector | cargo bay
#....#...#cccc#   11   bathroom | connector | cargo bay
####+#####+####   12   door(3)→utility | door(10)→engine
#WGr#.E...C...#   13   utility closet | engine room + cargo console
###############   14
```

### Room breakdown

- **Cockpit/Living Area (rows 1-5, cols 1-9):** BIG main room. Viewscreen across top wall. Nav console, drink station, food station all here.
- **Power Room (rows 1-2, cols 11-13):** Generator (g) + battery/power cell (p). Accessed from living area via door.
- **Hallway (rows 3-5, cols 11-13):** Right-side corridor above cargo bay.
- **Bedroom (rows 7-8, cols 1-4):** Sleeping berths (_). Accessed from living area via door.
- **Connector (rows 7-11, cols 6-8):** Mid-corridor connecting bedroom, bathroom, and cargo.
- **Bathroom (rows 10-11, cols 1-4):** Toilet (t) + shower (s) in ONE room.
- **Cargo Bay (rows 7-11, cols 10-13):** Tall cargo space, 4×5 = 20 cargo pads.
- **Utility Closet (row 13, cols 1-3):** Water tank (W), organic tank (G), matter recycler (r).
- **Engine Room (row 13, cols 5-13):** Engine (E) centered, cargo console (C).

**Spawn:** [5, 3] — center of living area, near nav console.

---

## Verification

### Step 1
- [ ] Food/drink stations work
- [ ] Recycler buffers fill and process
- [ ] Toggles work
- [ ] Legend updated

### Step 2
- [ ] Tab cycles overlay
- [ ] Water overlay shows blue pipes
- [ ] Organic shows green
- [ ] Power shows yellow
- [ ] Pipe glyphs auto-tile correctly

### Step 3
- [ ] Pipe breaks randomly
- [ ] Leak appears on floor
- [ ] Alert fires
- [ ] Walk to break, press E to fix
- [ ] Matter lost while leaking

### Step 4
- [ ] Tanks glow bright when full, dim when empty
- [ ] Pipes glow with flow
