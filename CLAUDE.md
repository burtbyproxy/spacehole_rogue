# SpaceHole Camper

ASCII space survival roguelike in Go. You're a stranded redshirt with a busted shuttle, scrounging fuel one jump at a time.

Star Trek parody tone. The USS Monkey Lion is out there somewhere - you were crew once. Now you're just trying to survive.

---

## Quick Reference

| Document | Purpose |
|----------|---------|
| [planning/CAMPER_DESIGN.md](planning/CAMPER_DESIGN.md) | **Current direction** - space camper survival loop, goals, camping mechanics |
| [planning/DESIGN.md](planning/DESIGN.md) | Original vision (ML lore, races, episodes) |
| [planning/ARCHITECTURE.md](planning/ARCHITECTURE.md) | Tech stack, directory structure, ECS components |
| [planning/SYSTEMS.md](planning/SYSTEMS.md) | Matter flow, skills, alerts, episodes, races |
| [planning/PHASES.md](planning/PHASES.md) | Implementation roadmap |

### Implementation Specs

| Phase | Document |
|-------|----------|
| Sector Navigation | [planning/phase-07-sector-nav.md](planning/phase-07-sector-nav.md) |
| System Map + Flight | [planning/phase-07b-system-map.md](planning/phase-07b-system-map.md) |
| Matter Refactor | [planning/phase-matter-refactor.md](planning/phase-matter-refactor.md) |

---

## The Pitch

Space van life roguelike. Each jump costs nearly all your fuel - you're stranded until you solve the system's problem.

**Design Pillars:**
1. Jump fuel is the critical gate - can't leave without it
2. Where you camp matters - planet type, threat level, gear requirements
3. "One more jump" is the addiction
4. Sandbox goals - find the ML, find home, become a pirate, or just survive
5. Meaningful scarcity

**Core Loop:** Jump → Stranded → Find Camp → Survive/Gather → Solve Episode → Upgrade → Jump

---

## Tech Stack

- **Go** + **Ebitengine** (rendering) + **Ark ECS**
- CP437 ASCII tileset, 16x16 glyphs
- SQLite persistence via repository pattern
- Single binary distribution

---

## Current Phase

Check [planning/PHASES.md](planning/PHASES.md) for implementation status and next steps.

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