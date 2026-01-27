# SpaceHole Rogue — Game Systems Reference

## Matter System — Finite Resource Management

The core survival mechanic. All matter on a ship/outpost flows through a directed graph of nodes. The complexity scales with ship size — the RULES are the same, the SCALE is what changes.

### Matter Types

| Type | Examples | Shuttle version | Big ship version |
|------|----------|-----------------|------------------|
| Water | Drinking, hygiene, coolant | 1 tank, 1 pipe | 20+ tanks, recyclers, treatment plants |
| Organic | Food, biological waste | 1 food locker | Mess hall replicators, hydroponic bays |
| Energy | Power generation/consumption | 1 power cell | Reactor → power grid → every system |
| Structural | Hull plating, repair material | Patch kit | Fabricators, salvage processing, hull armor |

Expandable enum — future types: dark matter (SH Drive), coolant, plasma.

### Matter Primitives (ECS Components)

```go
type MatterType uint8
const (
    MatterWater MatterType = iota
    MatterOrganic
    MatterEnergy
    MatterStructural
)

type MatterVolume struct {
    Type        MatterType
    AmountClean int
    AmountDirty int
}

type MatterInput struct {
    Capacity int
    Volume   MatterVolume
    Active   bool
}

type MatterStorage struct {
    Capacity int
    Volume   MatterVolume
}

type MatterOutput struct {
    Capacity   int
    Efficiency int
    Active     bool
}

type MatterTreatment struct {
    Type       MatterType
    Capacity   int
    Efficiency int
    Duration   int
    Active     bool
    Progress   int
}

type MatterLeak struct {
    Rate       int
    MatterType MatterType
}
```

### Composed Nodes

| Node | Components | Example |
|------|------------|---------|
| Pipe | MatterInput + MatterStorage (small) + MatterOutput | Water pipe, power conduit |
| Tank | MatterInput + MatterStorage (large) + MatterOutput | Water tank, fuel tank, battery |
| Treatment | MatterInput + MatterStorage + MatterTreatment + MatterOutput | Water recycler, waste processor |
| Sink | MatterInput + MatterStorage (always open) + MatterTreatment + MatterOutput | Kitchen sink |
| Generator | MatterStorage + MatterOutput | Reactor, replicator |
| Consumer | MatterInput | Life support, consoles |
| Matter Bank | MatterInput + MatterStorage (huge) + MatterOutput | Central outpost hub |

### MatterFlowSystem (push-based, per tick)

1. Outputs push matter to connected inputs (up to capacity)
2. Inputs accept and move to storage
3. Treatment converts dirty → clean
4. Storage overflow = leak event
5. Leaks subtract matter + emit alert
6. Update ship-wide totals for HUD bars

### Connections

Every connection is a pipe entity with a position on the tile map. Pipes are physical — visible in Johnny Tubes, damageable, repairable. Engineering crew (or the player) physically walks to broken pipes to fix them.

---

## Skill Progression (Learn by Doing)

No skill points. No allocation screens. Skills improve by USING them — Skyrim/Kenshi style.

| Skill | Improved by | Effect |
|-------|-------------|--------|
| Engineering | Repairing pipes, fixing subsystems, upgrading equipment | Faster repairs, diagnose problems sooner, craft upgrades |
| Combat | Fighting (phaser, melee) | Better accuracy, damage, dodge chance |
| Piloting | Flying the ship, maneuvering in combat | Fuel efficiency, evasion, docking speed |
| Science | Scanning, research, using science equipment | Better scan results, identify anomalies, research speed |
| Diplomacy | Trading, talking to NPCs, negotiating | Better trade prices, crew morale influence, new dialogue options |
| Leadership | Giving orders, managing crew, making command decisions | Crew follows orders faster, morale bonus, unlock command abilities |
| Survival | Eating, managing resources, resisting hazards | Slower hunger/thirst, resist infections, environmental resilience |

Skills have visible levels (1-10) shown on the character sheet. Each level unlocks a specific perk or ability that the player can read in plain language: "Engineering 3: You can now repair shield emitters." No stat tables, no math — just "you can now do this thing you couldn't before."

---

## Discovery System (No Man's Sky / Starfield Vibes)

Exploration should feel rewarding even when there's no combat or episode.

### Discovery Log Categories

- **Star Systems** — First visit to each system type (binary star, red dwarf, nebula, etc.) gets logged with a description and XP bonus
- **Races** — First contact with each race unlocks their lore page. "Deborahs: Literally just zebras. Scientists remain baffled by their career success."
- **Technology** — Find unique tech on derelicts, in outposts, from trade. Each discovery logged.
- **Anomalies** — Weird stuff: a planet that orbits backward, a singing asteroid, a black hole with a space station in its accretion disk
- **Crew Stories** — Key moments with crew members get logged as "personal log entries."

### Discovery Bonuses

- First time visiting a star type: XP + credits bonus
- First contact with a race: Unique encounter + lore
- Scanning a planet: Reveals resources, points of interest, potential hazards
- Derelict exploration: Random loot tables (weapons, tech, matter, crew in cryo to rescue)
- Rare finds: Ancient technology, abandoned colonies with backstory, unique named items

The "just one more" pull: The discovery log has visible gaps. "Systems discovered: 12/47". "Races encountered: 3/7". Completionism drives exploration.

---

## Alert System — How the Ship Talks to You

On bigger ships, you can't watch everything. The game communicates through an alert priority system in the message log:

| Priority | Color | Example |
|----------|-------|---------|
| CRITICAL | Red, blinking | "HULL BREACH Deck 3 Section 7 — Matter venting to space!" |
| Warning | Yellow | "Water pressure dropping in Section 4. Engineering dispatched." |
| Info | Cyan | "Crew shift change. Alpha shift reporting to stations." |
| Social | White | "Lt. Torres and Ensign Kim are arguing about replicator privileges." |
| Discovery | Green | "Sensors detect an uncharted asteroid field in System J-7." |

**Key design rule:** The game never gives you a problem without also telling you what's being done about it (on bigger ships). "Water pressure dropping — Engineering dispatched." You can intervene or trust the crew. On the shuttle, YOU are the alert system — you see the red pipe yourself.

---

## Episode Generator — Procedural Encounters

Each star system entry rolls one from each table.

### Mission Tables

**Investigate:** Crash Site, Derelict Ship, Distress Signal, Everyone Disappeared, Missing Scientist, Vanished Ship

**Military:** Inspection, Pursuit, Shore Leave

**Research:** Blackhole, Civilization, Nebula, Planet, Quasar, Star, Unexplored System, Wormhole

**Support:** Defend, Deliver, Medical, Rescue, Transport

### Twists
Assassination Attempt, Crew Infected, Equipment Malfunction, Marooned, Officer Goes Insane, Series of Murders, Ship Captured, Ship Damaged, Surprise Attack, Taken Prisoner, Thoughts Manifested, Time Travel, Unjustly Court Martialed

### Locations
Alternate Dimension, Civilian Colony, Paradise Garden, False Utopia, Military Outpost, Mining Base, Pleasure District, Prehistoric Planet, Prison, Research Base, Starknight Command

### Characters
Alien Ambassador, Ambitious Officer, Brainwashed Colonists, Creepy Children, Deranged Scientist, Eccentric Trader, Evil Twin, Genetic Superhuman, Giant Cube, Historical Figure, Honorable Enemy Captain, Hotshot Pilot, Lonely Godling, Love Interest, Massive Single Celled Organisms, Molten Stone Creature, Old Rival, Powerful Psychic, Primitive Monster, Reclusive Dictator, Robot Overlord, Rogue Satellite, Secret Weapon, Sentient Cloud, Shady Diplomat, Shakespearean Acting Troupe, Space Hippies, Supercomputer, War Criminal

Single roll per table. All seeded. Certain episodes can also drop ML clues (Chapter 1).

---

## Races

Player is always human. Other races are NPCs.

| Race | Description | Gameplay Traits |
|------|-------------|-----------------|
| Humans | "The most boring race" | Baseline |
| Bampeetos | Raccoons with amazing buttrock hair | Nocturnal, steal cargo/trash, destructive when bored |
| Tonies | Unicorn heads on stick bodies | Advanced (bad) genetic engineering, always angry |
| Deborahs | Literally just zebras | Hold important jobs they can't do, everyone thinks they're genius |
| Purgons | Humanoid cats | Steal technology, purring when nearby |
| Rasheans | Humanoid turtles | Instant bulletproof shell, panic-fire when scared (often) |
| Species 7482 | Assimilation claws | Primitive tech, scratching converts you slowly |

Expandable — new races are data, not code changes.

---

## Ship & Location System

Ships, outposts, planets share a base tile/room/equipment system with key differences.

- **Ships**: Multi-deck, mobile, subsystems, power grid, crew hierarchy, can dock
- **Outposts**: Single-level (usually), fixed position, type (Military/Extraction/Research), matter bank
- **Planet Surfaces**: Outdoor tile maps, landing party context, terrain types

### Crew Hierarchy

| Tier | Roles | Color |
|------|-------|-------|
| Captain | Captain | Gold star |
| Commanders | Science Officer, Chief Medical Officer, Chief Engineer, Tactical Officer | Gold |
| Specialists | Communications, Navigator, Nurse, Transporter Tech, Commando | Blue |
| Support | Science, Technical, Security, Trainees | Red |
| Civilians | Ambassadors, Diplomats, Love Interests, Scientists, Entertainers | White |
| Prisoners | Bad guys in the brig | Gray |

### Rooms (Ship)
Bridge, Captain's Quarters, Officer's Quarters, Quarters, Ready Room, Armory, Engineering, Mess Hall, Cargo Bay, Barracks, Johnny Tubes, Auditorium, Lounge, Holodeck, Science Center, Shuttle Bay, Sickbay, Turbolift, Transporter Room, Brig

### Rooms (Outpost)
Bathrooms, Mess Hall, Transporter Room, Cargo Bay, Barracks, Quarters, Meeting Room, Lounge (with Jukebox, Dance Floor, Game Cabinets, Bar), Offices, Holodeck, Landing Pad, Garage, Security, Brig, Workshop, Lab

### Equipment

Both fixed (Warp Core, Viewscreen, Captain's Chair) and portable (Phasers, Knives, Tools). Player has an inventory with slots.

---

## Matter System Refactor (Shuttle Scale)

### New Matter Flow

```
CLEAN TANKS ──→ DISPENSERS ──→ PLAYER BODY ──→ WASTE
  ↑                              ↓
  └── MATTER RECYCLER ←── DIRTY POOLS ←── TOILET/SHOWER
```

- **Food Station**: vending machine for food. E to eat. 5 clean organic → body, -35 hunger.
- **Drink Station**: vending machine for water. E to drink. 3 clean water → body, -25 thirst.
- **Matter Recycler**: combined unit. Toggleable (T). Has internal buffer per type (capacity 5 each on shuttle). Pulls dirty matter in, processes over time, spits clean matter out. Costs 1 energy per unit processed.
- **Tanks (Water, Organic)**: info only. No eating/drinking at tanks.

### Equipment Types

| Equipment | Glyph | Purpose |
|-----------|-------|---------|
| FoodStation | F | Food vending machine (dispenses clean organic) |
| DrinkStation | D | Water vending machine (dispenses clean water) |
| MatterRecycler | ▒ | Combined dirty→clean for water + organic, with buffer tanks |
| Generator | g | Power generator (generates energy at fixed rate) |

Engine no longer generates power — Generator does. Engine provides thrust only.

### Pipe Infrastructure

Pipes route through tiles using a bitmask system:

```go
type PipeFlags uint8
const (
    PipeWater   PipeFlags = 1 << iota // blue
    PipeOrganic                        // green
    PipePower                          // yellow
)
```

Press Tab to cycle overlay modes (Normal → Water → Organic → Power). In overlay mode, pipes glow through walls.

### Breakage + Leaks

Pipes and equipment have condition (0-100). Degrades over time. At 0, breaks.

Visual deterioration in overlay:
- 100-75: bright pipe color (healthy)
- 74-50: slightly dimmer
- 49-25: yellow tint (warning)
- 24-1: red tint (critical)
- 0: broken glyph (~), red, blinking

When broken: matter leaks, alert fires. Walk to break, press E to fix.
