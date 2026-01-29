# SpaceHole Rogue

ASCII space survival roguelike in Go. You're a stranded redshirt with a busted shuttle, scrounging fuel one jump at a time.

Star Trek parody tone. The USS Monkey Lion is out there somewhere - you were crew once. Now you're just trying to survive.

---

## Quick Reference

| Document | Purpose |
|----------|---------|
| **[planning/TODO.md](planning/TODO.md)** | **Current status and roadmap** |
| [planning/DESIGN.md](planning/DESIGN.md) | Game vision, design pillars, story |
| [planning/ARCHITECTURE.md](planning/ARCHITECTURE.md) | Tech stack, directory structure, ECS |
| [planning/SYSTEMS.md](planning/SYSTEMS.md) | Matter flow, skills, alerts, episodes |
| [planning/PHASES.md](planning/PHASES.md) | Implementation phases |

---

## Current Status

**Working:** Prologue, shuttle systems, sector/system maps, planet exploration, stations, trading, discovery

**Next up:** Jump drive fuel system, equipment repair, ship encounters, missions, danger!

See [planning/TODO.md](planning/TODO.md) for full details.

---

## The Pitch

Space van life roguelike. Each jump costs nearly all your fuel - you're stranded until you solve the system's problem.

**Design Pillars:**
1. Jump fuel is the critical gate - can't leave without it
2. Power management creates interesting choices
3. "One more system" is the addiction
4. Meaningful scarcity - closed-loop matter system
5. Roguelike danger - death is real

**Core Loop:** Jump → Stranded → Explore → Gather Resources → Solve Problems → Upgrade → Jump

---

## Tech Stack

- **Go** + **Ebitengine** (rendering) + **Ark ECS**
- CP437 ASCII tileset, 16x16 glyphs
- SQLite persistence via repository pattern
- Single binary distribution

---

## Equipment Power Model

Equipment has three power modes:
- **PowerNone** - No power needed (beds, doors, tanks)
- **PowerConstant** - Reserves power while ON (generator, recycler, consoles)
- **PowerOnUse** - Draws power per interaction (replicators, shower)

Constant-draw equipment must be turned ON (T key) to use. Reserved power shows as dark gold in the energy bar.

---

## Key Commands (In-Game)

| Key | Action |
|-----|--------|
| WASD/Arrows | Move (ship interior) / Fly (system map) |
| E | Interact / Use |
| T | Toggle equipment |
| Tab | Cycle overlay (pipes) / Toggle character sheet |
| N | Open nav map (from system map) |
| ESC | Back / Menu |

---

## File Structure

```
cmd/spacehole/main.go          ← Entry point
internal/game/                 ← Simulation (ECS, systems, tick)
internal/render/               ← Ebitengine rendering
internal/world/                ← Data structures, loaders
internal/data/                 ← Persistence layer
assets/                        ← Fonts, ship layouts, game data
planning/                      ← Design docs (you are here)
```

---

## The Vibe

> "Your thermal gear will last maybe an hour in this cold. The scanner says there's a fuel depot 3 klicks north, under the glacier. The pirate frigate is still in orbit. You're not dying on this frozen rock."

That's the game.