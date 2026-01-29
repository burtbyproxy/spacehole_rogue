# SpaceHole Rogue - Current TODO

## Current State (as of Phase 7b+)

### WORKING
- [x] ASCII rendering with CP437 tileset
- [x] Shuttle layout with rooms and equipment
- [x] Player movement and collision
- [x] Equipment interaction (E key) and toggle (T key)
- [x] Matter system: water/organic tanks, clean/dirty cycles
- [x] Recycler: dirty -> clean conversion
- [x] Generator: produces power over time
- [x] Player needs: hunger, thirst, hygiene (slow background drain)
- [x] Food/drink stations dispense from clean matter
- [x] Toilet flushes waste back to dirty pools
- [x] Shower uses clean water for hygiene
- [x] Power reservation system (constant-draw equipment)
- [x] HUD with resource bars (dirty/clean split, reserved/free power)
- [x] Message log with priority colors
- [x] Sector map with procedural star systems
- [x] System map with 2D flight (WASD)
- [x] Planets, stations, derelicts, NPC ships
- [x] Planet scanning from orbit
- [x] Surface exploration with fog of war
- [x] Prologue: stranded start, find parts/power to escape
- [x] Discovery log and XP system
- [x] Docking at stations (auto-refill, buy/sell cargo)
- [x] Cargo system with pads
- [x] Personal inventory for packs
- [x] Episodes (random events on system entry)
- [x] Consoles require power to use

### IN PROGRESS
- [ ] Jump drive system (have equipment, need fuel consumption + UI)
- [ ] Equipment condition/degradation system (have Condition field, unused)

---

## Priority Roadmap

### 1. Jump Drive System
**Goal:** Travel between star systems costs jump fuel, creates tension

- [ ] Nav console shows fuel cost to each system
- [ ] Jump consumes fuel from JumpFuel pool
- [ ] Can't jump without enough fuel
- [ ] Incinerator converts cargo -> fuel (already works!)
- [ ] Fuel cells found on surfaces add to fuel tank

### 2. Equipment Repair System
**Goal:** Equipment breaks, you fix it with spare parts

- [ ] Equipment degrades on use or randomly
- [ ] Low condition = reduced efficiency
- [ ] 0 condition = broken, can't turn on
- [ ] Spare parts item repairs equipment (E on broken equipment)
- [ ] Find spare parts on planets, buy at stations
- [ ] Jump stress: small chance equipment degrades on FTL jump

### 3. Component Install/Uninstall
**Goal:** Customize your ship at shipyards

- [ ] Shipyard station type
- [ ] Remove equipment from tile (get component item)
- [ ] Install equipment from inventory to empty tile
- [ ] Equipment has different tiers/quality
- [ ] Cost credits for install/uninstall service

### 4. Hail/Encounter System
**Goal:** NPC ships are interesting, not just obstacles

- [x] Ships hail when nearby (have HailState)
- [x] Viewscreen answers hails
- [ ] Encounter dialogue with options
- [ ] Trade with merchants
- [ ] Bribe/fight/flee pirates
- [ ] Patrol checks (cargo inspection)
- [ ] Distress calls you can answer

### 5. Mission System
**Goal:** Stations give you things to do

- [ ] Mission board at stations
- [ ] Delivery missions (cargo A to station B)
- [ ] Retrieval missions (get item from planet)
- [ ] Escort missions (follow ship to destination)
- [ ] Time limits on some missions
- [ ] Reputation with factions

### 6. Danger Systems
**Goal:** This is a roguelike - death is real

- [ ] Hostile ships attack on sight
- [ ] Hull damage from combat
- [ ] Planet hazards (radiation, hostile creatures)
- [ ] Equipment malfunction events
- [ ] Permadeath with score/stats
- [ ] Random events during travel

---

## Lower Priority (Future)

### Combat System
- Ship weapons and shields
- Targeting subsystems
- Personal weapons for away missions
- (Shuttle has no weapons - need to upgrade first!)

### Crew System
- Hire crew at stations
- Crew has skills and needs
- Larger ships need crew to function

### Ship Progression
- Find/buy larger ships
- More equipment slots
- The USS Monkey Lion as endgame goal

---

## Technical Debt

- [ ] Consolidate duplicate bar drawing code
- [ ] Equipment templates could use inheritance/composition
- [ ] Surface generation could be more varied
- [ ] Save/load system not implemented yet
