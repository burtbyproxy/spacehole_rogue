# SpaceHole Rogue

ASCII space roguelike in Go. Start as nobody, find the legendary USS Monkey Lion, bring it home.

Star Trek parody tone. A zebra named Deborah is your Chief Science Officer. She has no idea what she's doing.

---

## Quick Reference

| Document | Purpose |
|----------|---------|
| [planning/DESIGN.md](planning/DESIGN.md) | Game vision, design pillars, gameplay loops, story structure |
| [planning/ARCHITECTURE.md](planning/ARCHITECTURE.md) | Tech stack, directory structure, ECS components, system execution order |
| [planning/SYSTEMS.md](planning/SYSTEMS.md) | Matter flow, skills, alerts, episodes, races — the mechanics reference |
| [planning/PHASES.md](planning/PHASES.md) | Implementation roadmap with verification milestones |

### Implementation Specs

| Phase | Document |
|-------|----------|
| Sector Navigation | [planning/phase-07-sector-nav.md](planning/phase-07-sector-nav.md) |
| System Map + Flight | [planning/phase-07b-system-map.md](planning/phase-07b-system-map.md) |
| Matter Refactor | [planning/phase-matter-refactor.md](planning/phase-matter-refactor.md) |

---

## The Pitch

The game should FEEL like a hardcore simulation but PLAY like a fun roguelike.

**Design Pillars:**
1. The game teaches itself through ship scale (shuttle → small ship → ML)
2. Automate the routine, dramatize the exceptions
3. "One more system" is the addiction
4. Crew you care about
5. Meaningful scarcity

**Core Loop:** Pick destination → Travel → Episode → Reward → Maintain → Repeat

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

> "The Deborah in Cargo Bay B is eating your supplies. She is a zebra. She doesn't understand inventory management."

That's the game.