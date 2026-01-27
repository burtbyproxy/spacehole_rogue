# SpaceHole Rogue — Implementation Phases

## Overview

This document outlines the MVP implementation phases. Each phase has a "Is it fun yet?" verification milestone. Detailed implementation specs for complex phases live in separate files.

---

## Phase 0: Scaffolding
- `go mod init`, directory structure, Ark + Ebitengine deps
- Makefile, Dockerfile
- Minimal window: "SpaceHole Rogue"

**Verify:** Window opens

---

## Phase 1: ASCII Grid Rendering
- CP437 atlas loading, CellBuffer, color palette
- Hardcoded test grid to verify rendering

**Verify:** ASCII renders correctly

---

## Phase 2: The Shuttle
- Design shuttle layout (15x10, 4 rooms, 3 pipes)
- JSON loader, TileGrid
- Render shuttle statically

**Verify:** Shuttle appears on screen, looks like a tiny ship

---

## Phase 3: ECS + Game Loop
- Ark world, core components, TickPolicy (real-time)
- Spawn player @ on the shuttle

**Verify:** `@` exists, game ticks

---

## Phase 4: Movement + Collision
- WASD/arrows, MovementSystem, SpatialIndex
- Walk around the shuttle, blocked by walls

**Verify:** Walk around the shuttle, walls block you

---

## Phase 5: Matter System (Shuttle Scale)
- 3 pipes, 1 water tank, 1 power cell, 1 food locker
- MatterFlowSystem, HUD resource bars
- Break a pipe manually (debug key) → see water drop → walk to pipe → press E to fix

**Verify:** Fix a broken pipe, watch water bar recover. FIRST FUN.

---

## Phase 6: Shuttle HUD + Message Log
- Resource bars (Water/Energy/Organic/Hull)
- Message log with colored priority alerts
- Alert system: "Water low!" in yellow/red

**Verify:** HUD tells you what's happening without confusion

---

## Phase 7: Sector Map + Navigation
→ See [phase-07-sector-nav.md](phases/phase-07-sector-nav.md)

- Procedural sector from seed
- Sector map view, star system view
- Fly the shuttle between systems
- Context-driven view transitions

**Verify:** Fly to a new system and see what's there. ONE MORE SYSTEM.

---

## Phase 7b: System Map + Local Flight + NPC Ships
→ See [phase-07b-system-map.md](phases/phase-07b-system-map.md)

- 2D scrolling space view with free WASD flight
- Star, planets, stations, derelicts, NPC ships
- NPC ships move autonomously (traders, patrols, pirates)
- Approach detection for interaction

**Verify:** Fly around a system, see NPC ships moving, approach objects for info

---

## Phase 8: Stations + Trading
- Dock at station → menu (buy/sell/repair/refuel)
- Credits, trade goods, price variation

**Verify:** Buy low, sell high, upgrade your shuttle. TRADE LOOP.

---

## Phase 9: Discovery System
- Discovery log, first-visit bonuses
- Scanning planets, logging anomalies
- Character sheet with skills (Engineering starts leveling from pipe repairs)

**Verify:** "I discovered a new star type!" feels rewarding

---

## Phase 10: Ship Layout + Crew (Small Ship)
- Design 5-deck scout ship, full room set
- Join as crew member (scripted transition from shuttle gameplay)
- 15-25 NPC crew with roles, pathfinding, basic AI
- Turbolifts, doors, minimap

**Verify:** Walk around a living ship with crew. THE SHIP IS ALIVE.

---

## Phase 11: Deep Crew Simulation
- Needs (fatigue, hunger, social), duty schedules, morale
- Personality traits, relationships, crew chatter in message log
- Engineering crew autonomously finds and repairs leaks

**Verify:** Crew drama makes you laugh. You care about NPCs.

---

## Phase 12: Ship Subsystems + Matter at Scale
- Full power grid, subsystems (shields, weapons, engines, etc.)
- Matter network at ship scale (20+ tanks, recyclers, treatment)
- Alert system for ship problems
- Damage → subsystem degradation → crew responds

**Verify:** Subsystem damage creates real tension

---

## Phase 13: Combat
- Ship-to-ship: weapons, shields, subsystem targeting
- Personal: phaser/melee on away teams
- Combat views

**Verify:** Combat is tense and tactical.

---

## Phase 14: Episode Generator
- Episode tables from JSON, roll on system entry
- Episode state management
- ML clue system (Chapter 1)

**Verify:** An episode surprises you. "I didn't expect THAT."

---

## Phase 15: Races
- 7 races with unique behaviors
- Race-specific encounters, crew recruitment

**Verify:** A Deborah is your science officer. You love her. She's a zebra.

---

## Phase 16: Save/Load + Polish
- SQLite persistence, main menu
- Turn-based mode option
- UI polish, help tooltips (contextual, not tutorial dumps)

**Verify:** Save, quit, come back, everything is exactly as you left it

---

## Phase 17: USS Monkey Lion
- ML ship layout, SH Drive rooms
- Chapter 1 → Chapter 2 transition
- SH Drive jump mechanic → new sector generation

**Verify:** The SH Drive fires. New sector. The adventure continues.

---

## Matter Refactor Phase
→ See [phase-matter-refactor.md](phases/phase-matter-refactor.md)

This can be done after Phase 5-6 or integrated into later phases:

1. **Step 1:** Split Replicator into dispensers + combined recycler with buffer tanks
2. **Step 2:** Add pipe/wire data to tiles, SimCity-style overlay toggle
3. **Step 3:** Breakage + leaks — pipes/equipment can break, leaking matter onto floors
4. **Step 4:** Visual polish — tank brightness by fullness, pipe brightness by flow
