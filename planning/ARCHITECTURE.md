# SpaceHole Rogue — Architecture

## Tech Stack

| Layer | Choice | Notes |
|-------|--------|-------|
| Language | Go | Single binary, fast simulation |
| Rendering | Ebitengine v2 | CP437 bitmap font atlas, 2D tile-based ASCII |
| ECS | Ark (github.com/mlange-42/ark) | Entity relations for ship/deck/crew/matter hierarchies |
| Persistence | SQLite via repository pattern | Abstraction layer for future swap to Redis/Postgres |
| Build/CI | Docker (multi-stage cross-compilation) | Native dev, Docker for release builds |
| Distribution | Single binary | Steam/itch.io/direct download |

---

## Directory Structure

```
cmd/spacehole/main.go             ← Entry point, Ebitengine RunGame()
    │
    ├── internal/game/             ← Game simulation (ECS, systems, tick loop)
    │   ├── game.go                ← Game struct, Init(), Tick()
    │   ├── tick.go                ← TickPolicy interface (real-time / turn-based)
    │   ├── ecs/
    │   │   ├── components.go      ← All ECS component definitions
    │   │   └── relations.go       ← Ark relation definitions
    │   ├── systems/               ← ECS systems (see System Execution Order)
    │   └── procgen/
    │       ├── sector.go          ← Sector generation from seed
    │       ├── system.go          ← Star system generation
    │       ├── names.go           ← Name generation
    │       └── episode.go         ← Episode table rolls
    │
    ├── internal/render/           ← All Ebitengine rendering
    │   ├── atlas.go               ← CP437 glyph atlas loader
    │   ├── grid.go                ← CellBuffer, ASCII grid rendering
    │   ├── camera.go              ← Viewport/camera
    │   ├── minimap.go             ← Minimap renderer
    │   ├── hud.go                 ← HUD overlay (resource bars, alerts)
    │   ├── colors.go              ← Color palette
    │   ├── viewmanager.go         ← Context-driven automatic view switching
    │   └── views/                 ← Individual view renderers
    │
    ├── internal/world/            ← World data structures & loading
    │   ├── location.go            ← Base Location type
    │   ├── ship.go                ← Ship-specific definitions
    │   ├── outpost.go             ← Outpost-specific definitions
    │   ├── sector.go              ← Sector, star system, planet definitions
    │   ├── tilemap.go             ← Tile types and grids
    │   ├── loader.go              ← Load layouts from JSON
    │   └── matter_network.go      ← Matter flow graph construction
    │
    ├── internal/data/             ← Repository pattern data layer
    │   ├── repository.go          ← Interfaces
    │   ├── sqlite/
    │   │   ├── store.go           ← SQLite connection & migrations
    │   │   └── saverepo.go        ← Save/load game state
    │   └── models.go              ← DTOs
    │
    ├── internal/input/
    │   ├── handler.go             ← Keyboard + mouse input mapping
    │   └── keybinds.go            ← Configurable key bindings
    │
    ├── pkg/types/
    │   └── types.go               ← Vec2, Direction, enums
    │
    ├── pkg/config/
    │   └── config.go
    │
    └── assets/
        ├── fonts/cp437_16x16.png
        ├── ships/                 ← Ship layout JSON files
        ├── outposts/              ← Outpost layout JSON files
        └── data/                  ← Game data (races, episodes, items, skills)
```

---

## Core Architecture Pattern

**Ebitengine owns the main loop.** Game simulation is a library called by `Update()`. Rendering reads game state in `Draw()`.

```
┌─────────────────────────────────────────┐
│  Ebitengine RunGame()                   │
│  ┌───────────────┐  ┌────────────────┐  │
│  │   Update()    │  │    Draw()      │  │
│  │  ┌─────────┐  │  │  ┌──────────┐  │  │
│  │  │  Game   │  │  │  │ Renderer │  │  │
│  │  │  Tick() │  │  │  │ (reads   │  │  │
│  │  │         │  │  │  │  state)  │  │  │
│  │  └─────────┘  │  │  └──────────┘  │  │
│  └───────────────┘  └────────────────┘  │
└─────────────────────────────────────────┘
```

---

## Tick System — Dual Mode via TickPolicy

```go
type TickPolicy interface {
    ShouldTick() bool
    OnPlayerAction()
    TickRate() int
}

type RealTimePolicy struct { interval time.Duration; lastTick time.Time }
type TurnBasedPolicy struct { playerActed bool }
```

Configurable as a game setting. Systems don't know or care which mode is active.

---

## ECS Components (Grouped by Category)

### Spatial
- `Position { X, Y int }`
- `DeckLevel { Index int }`
- `Facing { Dir Direction }`
- `MovementIntent { TargetX, TargetY int }`
- `SectorPosition { X, Y float64 }`

### Rendering
- `Renderable { Glyph rune; FG, BG uint32; Priority int }`

### Identity
- `PlayerControlled {}`
- `Named { Name string }`
- `Race { Type RaceType; Traits []RaceTrait }`

### Player
- `PlayerSkills { Engineering, Combat, Piloting, Science, Diplomacy, Leadership, Survival int }`
- `SkillXP { ... float64 }` — fractional XP toward next level
- `DiscoveryLog { Systems, Races, Tech, Anomalies map[string]bool }`

### Crew
- `CrewMember { Role CrewRole; Rank CrewRank }`
- `CrewStats { Health, Morale, SkillLevel int }`
- `CrewAI { State AIState; Goal AIGoal; IdleTicks int }`
- `DutySchedule { Shift int }`
- `Relationships { Friends, Rivals []entity }`
- `Needs { Fatigue, Hunger, Social int }`
- `Personality { Traits []PersonalityTrait }`

### Ship Structure
- `Ship { Name, Class string }`
- `Deck { Number int; Name string; Width, Height int }`
- `Door { Open, Locked bool }`
- `Turbolift { LiftID int; ConnectedDecks []int }`
- `Station { Type StationType }`

### Ship Subsystems
- `Subsystem { Type SubsystemType; Health, MaxHealth int; PowerDraw float64; Active bool }`
- `PowerGrid { TotalPower, AllocatedPower float64 }`
- `ShipHull { HP, MaxHP int; Armor int }`
- `ShipShields { Strength, MaxStrength int; RechargeRate float64 }`

### Matter Flow
- `MatterInput { Capacity int; Volume MatterVolume; Active bool }`
- `MatterStorage { Capacity int; Volume MatterVolume }`
- `MatterOutput { Capacity, Efficiency int; Active bool }`
- `MatterTreatment { Type MatterType; Capacity, Efficiency, Duration int; Active bool; Progress int }`
- `MatterLeak { Rate int; MatterType MatterType }`

### Combat
- `Weapon { Type WeaponType; Damage, Range, Cooldown, CooldownRemaining int }`
- `Health { Current, Max int }`
- `Combatant { AttackPower, Defense int }`

### Economy & Items
- `Equipment { Name string; Type EquipmentType; Fixed bool; UseVerb string }`
- `Inventory { Items []InventorySlot; MaxSlots int }`
- `Credits { Amount int }`
- `TradeGoods { Type TradeGoodType; Quantity, BasePrice int }`

### World
- `StarSystem { Name string; X, Y float64; StarType StarType }`
- `Planet { Name string; Type PlanetType; OrbitIndex int }`
- `SpaceStation { Name string; Type OutpostType; Services []ServiceType }`
- `Outpost { Type OutpostType }`

### Episode & Story
- `ActiveEpisode { Mission MissionType; Twist TwistType; Location LocationType; Character CharacterType }`
- `MonkeyLionClue { ClueIndex int; Description string }`

### Relations (Ark)
- `LocatedOn` — entity on a Deck
- `BelongsTo` — entity belongs to Ship/Outpost
- `AssignedTo` — crew to Station
- `OrbitOf` — planet orbits StarSystem
- `DockedAt` — ship at station
- `ConnectedTo` — matter output → matter input
- `InsideOf` — item in container

### World Resources (Singletons)
- `TickCounter { Tick uint64 }`
- `DeckTiles { Grids map[entity]*TileGrid }`
- `SpatialIndex { Grid map[SpatialKey][]entity }`
- `SectorState { Systems, Ships []entity }`
- `GameSeed { Seed int64 }`
- `MatterNetworkState { TotalWater, TotalOrganic, TotalEnergy, TotalStructural int }`
- `EpisodeState { Current *ActiveEpisode; CluesFound int }`
- `AlertQueue { Alerts []Alert }`

---

## System Execution Order

1. `InputSystem` — player input → commands
2. `CrewAISystem` — NPC behavior, A* pathfinding
3. `NeedsSystem` — fatigue/hunger/social → morale
4. `DutySystem` — shift changes, station assignments
5. `DoorSystem` — auto-open/close
6. `MovementSystem` — movement, collision, SpatialIndex
7. `TurboliftSystem` — deck transitions
8. `MatterFlowSystem` — resource flow, leaks
9. `SubsystemSystem` — power, damage, repairs
10. `CombatShipSystem` — ship-to-ship
11. `CombatPersonalSystem` — on-foot combat
12. `EconomySystem` — prices, supply/demand
13. `SectorSimSystem` — NPC ships, events
14. `EpisodeSystem` — episode management, clues
15. `SkillSystem` — XP accumulation, level-ups
16. `DiscoverySystem` — log new discoveries
17. `AlertSystem` — generate/prioritize alerts
18. `ViewStateSystem` — determine active view

---

## View System — Context-Driven Automatic

```go
const (
    ViewShipInterior  // Walking inside ship or outpost
    ViewSectorMap     // Ships in sector space
    ViewSystemMap     // Planets/stations in a star system
    ViewStationMenu   // Docked (trade, recruit, repair, refuel)
    ViewShipCombat    // Ship-to-ship combat
    ViewLandingParty  // On-foot away mission
    ViewCharacter     // Character sheet, skills, inventory (manual toggle)
    ViewDiscovery     // Discovery log (manual toggle)
)
```

Transitions are automatic based on game state. Character/Discovery views are manual toggles (Tab key).

---

## Rendering — CP437 Atlas

- 16x16 CP437 bitmap font PNG (256 glyphs, white on transparent)
- Pre-sliced into 256 sub-images
- `CellBuffer` — 2D array of `Cell { Glyph, FG, BG }`
- Screen layout: HUD (top 2-3 rows) | Viewport + Minimap (middle) | Message log (bottom 3-5 rows)

### Glyph Conventions

| Entity/Tile | Glyph | Color |
|-------------|-------|-------|
| Player | `@` | White |
| Crew (by role) | `c` | Gold / Blue / Red / Cyan / White |
| Prisoners | `p` | Gray |
| Floor | `.` | Dark gray |
| Wall | `#` | Light gray |
| Door (closed/open) | `+` / `/` | Brown |
| Turbolift | `>` | White |
| Console | `=` | Green |
| Pipe (ok/leaking) | `─` / `~` | Blue or Yellow / RED blink |
| Star | `*` | Yellow/White/Red |
| Ship | `^v<>` | White (player), various (NPC) |

---

## Persistence — Repository Pattern

```go
type SaveRepository interface {
    SaveGame(state *GameState) error
    LoadGame(slotID string) (*GameState, error)
    ListSaves() ([]SaveInfo, error)
    DeleteSave(slotID string) error
}
```

SQLite implementation. Interface allows swapping backends later.

---

## Package Management

- **npm**: Works normally, global packages install to `/home/claude/.npm-global`
- **pip**: Use `--break-system-packages` flag
- Virtual environments: Create if needed for complex Python projects

---

## Build & Distribution

- **Local dev**: `make run` → `go run cmd/spacehole/main.go`
- **Release**: `make release` → Docker cross-compiles for Linux/Windows
- **CI**: GitHub Actions → Docker build → test → artifacts
