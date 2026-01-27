# Phase 7b: System Map + Local Flight + NPC Ships

## Summary

When the player presses E on the console, they enter the system map — a scrolling 2D space view where they fly their shuttle with WASD. The system contains a star, planets, stations, derelicts, and NPC ships. Approaching objects shows info; approaching stations lets you dock (future). Pressing N from the system map opens the sector nav map (existing) for interstellar jumps. After jumping, you arrive in the new system's map. NPC ships fly around autonomously, making space feel alive.

---

## View Transition Flow (updated)

```
Ship Interior (ViewShip)
  └─ E on console → System Map (ViewSystemMap)
       ├─ WASD: fly shuttle freely
       ├─ E near object: interact (approach info for now)
       ├─ N key → Sector Map (ViewSectorMap)
       │    ├─ WASD: snap between stars
       │    ├─ E: jump → System Map of NEW system
       │    └─ ESC → back to System Map
       └─ ESC → back to Ship Interior
```

**Key change:** jumping from sector map now lands in ViewSystemMap, not ViewShip.

---

## Files to Change

| File | Action | What |
|------|--------|------|
| `internal/game/system_map.go` | NEW | SpaceObject, SystemMap, generation, NPC movement, approach detection |
| `internal/game/sector.go` | MODIFY | Add `Map *SystemMap` field to StarSystem (lazy-generated on first visit) |
| `internal/game/sim.go` | MODIFY | Add shuttle position, system map tick, approach detection, console → system map signal |
| `cmd/spacehole/main.go` | MODIFY | Add ViewSystemMap mode, rendering, flight input, updated view transitions |

---

## Step 1: internal/game/system_map.go (new)

### Object Types

```go
type SpaceObjectKind uint8
const (
    ObjStar SpaceObjectKind = iota
    ObjPlanet
    ObjStation
    ObjDerelict
    ObjAsteroid
    ObjShip
)

type ShipAIKind uint8
const (
    AITrader ShipAIKind = iota
    AIPatrol
    AIPirate
)

type PlanetKind uint8
const (
    PlanetRocky PlanetKind = iota
    PlanetGas
    PlanetIce
    PlanetVolcanic
)
```

### SpaceObject Struct

```go
type SpaceObject struct {
    Kind       SpaceObjectKind
    Name       string
    X, Y       int       // position in system space
    Glyph      byte
    Color      uint8
    // Planet-specific
    PlanetType PlanetKind
    // Ship-specific
    AIKind     ShipAIKind
    DX, DY     int       // movement direction (-1, 0, 1)
    MoveRate   int       // ticks between moves (lower = faster)
    moveTimer  int       // internal countdown
}
```

### SystemMap Struct

```go
const (
    SystemMapW = 160   // 2x screen width
    SystemMapH = 80    // ~2x viewport height
)

type SystemMap struct {
    Width, Height      int
    Objects            []SpaceObject
    ShuttleX, ShuttleY int   // player shuttle position in system space
}
```

### Generation — GenerateSystemMap(seed int64, starType StarType) *SystemMap

- Star at center (80, 40), glyph `*`, color by type
- 2-5 planets at orbital distances (15-60 tiles from center)
  - Placed at random angles around the star
  - Random PlanetKind, glyph `O`, color by type
  - Names: "[Star Name] I", "[Star Name] II", etc.
- 0-1 station: placed 2-3 tiles from a random planet
  - Glyph `H`, color cyan
  - Named "[Star Name] Station" or "Outpost [Name]"
- 0-2 NPC ships: random positions, random AI types
  - Trader: glyph `T`, green, slow (MoveRate=12), wanders
  - Patrol: glyph `P`, blue, medium (MoveRate=8), orbits
  - Pirate: glyph `!`, red, fast (MoveRate=5), drifts toward shuttle
- 0-1 derelict: random position, glyph `%`, dark gray
- Shuttle starts at map edge (e.g., X=10, Y=40) — arriving from hyperspace

### NPC Movement — (sm *SystemMap) TickNPCs()

For each ObjShip:
- Decrement moveTimer; when 0, move by (DX, DY), reset timer to MoveRate
- Keep within bounds (bounce off edges)
- Every ~120 ticks, randomly change direction
- Pirates: bias DX/DY toward shuttle position (simple pursuit)

### Approach Detection — (sm *SystemMap) NearestObject(x, y, radius int) *SpaceObject

Returns closest non-ship object within radius, or nil.

---

## Step 2: internal/game/sector.go (modify)

Add `Map *SystemMap` to StarSystem:

```go
type StarSystem struct {
    Name    string
    X, Y    int
    Type    StarType
    Visited bool
    Map     *SystemMap  // nil until first visit, then lazy-generated
}
```

Add method `EnsureSystemMap(idx int)` to Sector — if `Systems[idx].Map == nil`, generate it using seed `s.Seed*1000 + int64(idx)` and the system's star type. Sets shuttle start position.

---

## Step 3: internal/game/sim.go (modify)

- Change ConsoleActivated semantics: it now signals "enter system map" (not sector map)
- Add `TickSystemMap()` — called from `Tick()`, advances NPC movement
- `NavigateTo()` now also calls `EnsureSystemMap` on the target and switches to it
- No shuttle position on Sim — that lives on the SystemMap itself

---

## Step 4: cmd/spacehole/main.go (modify)

### Add ViewSystemMap to enum

```go
const (
    ViewShip ViewMode = iota
    ViewSectorMap
    ViewSystemMap  // NEW
)
```

### System map viewport constants

```go
const (
    spaceViewportX = 0    // left edge of space viewport
    spaceViewportY = 2    // below title bar
    spaceViewportW = 58   // width (leaves room for info panel)
    spaceViewportH = 30   // height (leaves room for comms)
    spaceInfoX     = 60   // right panel for system info
)
```

### updateSystemMap() — new function

- ESC → ViewShip (back to ship interior)
- N → ViewSectorMap (open nav map, set cursor to current system)
- WASD (held, via `ebiten.IsKeyPressed`): move shuttle every 4 ticks = 15 tiles/sec
  - Track a `moveTimer` on Game struct, decrement each tick, move when 0
- Clamp shuttle position to system map bounds
- E near object: show approach message (planet info, station dock prompt, derelict scan)
- Check approach detection each frame → update info display

### drawSystemMapView() — new function

- Clear buffer
- Title: "--- [Star Name] System ---" + "[StarType]"
- Calculate camera offset: `camX = spaceViewportW/2 - shuttleX`, clamped to map bounds
- Draw background scatter stars (dim `.` at random positions, seeded)
- Draw all objects at camera-relative positions (only if within viewport)
- Draw shuttle `^` at viewport center
- Right panel: system info, nearby object details, ship status (energy/hull)
- Comms log at bottom
- Instructions: "WASD: Fly  E: Interact  N: Nav Map  ESC: Ship Interior"

### Updated view transitions

- `updateShip()`: ConsoleActivated → ViewSystemMap (was ViewSectorMap)
- `updateSectorMap()`: after successful NavigateTo → ViewSystemMap (was ViewShip)
- `updateSystemMap()`: N → ViewSectorMap, ESC → ViewShip

### Flight movement (held keys with throttle)

```go
// On Game struct:
flightTimer int  // decrements each tick, shuttle moves when 0

// In updateSystemMap():
const flightSpeed = 4  // move every 4 ticks (15 tiles/sec at 60 TPS)
g.flightTimer--
if g.flightTimer <= 0 {
    dx, dy := 0, 0
    if ebiten.IsKeyPressed(ebiten.KeyW) { dy = -1 }
    if ebiten.IsKeyPressed(ebiten.KeyS) { dy = 1 }
    if ebiten.IsKeyPressed(ebiten.KeyA) { dx = -1 }
    if ebiten.IsKeyPressed(ebiten.KeyD) { dx = 1 }
    if dx != 0 || dy != 0 {
        sm := currentSystemMap
        sm.ShuttleX = clamp(sm.ShuttleX+dx, 0, sm.Width-1)
        sm.ShuttleY = clamp(sm.ShuttleY+dy, 0, sm.Height-1)
        g.flightTimer = flightSpeed
    }
}
```

---

## Glyph & Color Reference (System Map)

| Object | Glyph | Color | Notes |
|--------|-------|-------|-------|
| Star | `*` | By type (yellow/red/blue/white/orange) | Center of map |
| Rocky planet | `O` | Brown | Most common |
| Gas planet | `O` | Magenta | Larger orbits |
| Ice planet | `O` | Light Cyan | Outer orbits |
| Volcanic planet | `O` | Light Red | Inner orbits |
| Station | `H` | Cyan | Near a planet |
| Derelict | `%` | Dark Gray | Random |
| Trader ship | `T` | Light Green | Slow wander |
| Patrol ship | `P` | Light Blue | Medium orbit |
| Pirate ship | `!` | Light Red | Fast, pursues |
| Player shuttle | `^` | White | Viewport center |
| Background stars | `.` or `·` | Dark Gray | Cosmetic scatter |

---

## NPC Ship Behavior (MVP)

All ships move on their own, creating a living space:

- **Traders (green T):** Slow drift between planets and stations. Change direction every ~180 ticks.
- **Patrols (blue P):** Circle the system at medium speed. Change direction every ~120 ticks.
- **Pirates (red !):** Fast. Bias movement toward shuttle (1 in 3 chance each direction change to aim at player). Creates tension even without combat.

When shuttle is within 3 tiles of an NPC ship:
- Trader: "A trader vessel drifts nearby."
- Patrol: "Patrol ship scanning the area."
- Pirate: "WARNING: Hostile vessel detected!" (MsgWarning)

---

## Key Design Decisions

- **System map is scrolling** — 160x80 tiles, camera follows shuttle. Gives a real sense of flying through space.
- **Held-key flight** — WASD continuous movement (not tap-per-tile). Feels like piloting, not walking.
- **Console → system map (not sector map)** — piloting is the primary function. Nav map is a sub-mode.
- **NPC ships create atmosphere** — even without combat, seeing ships move around makes space feel alive.
- **Lazy system generation** — each star system's content is generated on first visit, stored on the StarSystem struct.
- **Sim still ticks** — resources drain while flying. Need to go back inside to eat/drink/fix things.
- **Background scatter stars** — cosmetic dim dots make the black void feel like space.

---

## Verification

- [ ] Walk to console (=), press E → system map appears with star, planets, maybe station/ships
- [ ] WASD (held) flies shuttle smoothly through space
- [ ] Camera follows shuttle, scrolling across the 160x80 map
- [ ] Approaching a planet shows info: "Orbiting [Name]. [Type] planet."
- [ ] Approaching a station shows: "Docking range of [Name]. (future)"
- [ ] NPC ships visibly moving around the system
- [ ] Pirate approaches shuttle → warning message
- [ ] N key → sector nav map opens (existing functionality)
- [ ] Jump from sector map → arrive in new system map (not ship interior)
- [ ] ESC from system map → back to ship interior
- [ ] Resources still tick during flight
- [ ] Re-entering system map (E on console again) → same system, same positions
