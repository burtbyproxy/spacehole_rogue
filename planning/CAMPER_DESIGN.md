# SpaceHole Camper - Design Document

## The Pitch

**Space van life roguelike.** You're a stranded redshirt scraping together enough fuel to make one more jump, chasing the ghost of a ship you were once crew on. Camp on hostile worlds, scavenge derelicts, dodge pirates, and slowly piece together the path back to the USS Monkey Lion.

The game should FEEL like desperate survival but PLAY like a satisfying resource puzzle.

---

## Core Fantasy

You were nobody - a redshirt on the legendary USS Monkey Lion. Something happened. You woke up stranded on some backwater planet with nothing but a broken shuttle and a vague memory of where the ship was headed.

Now you're a space nomad. Your shuttle is your camper van. Each system is a pit stop where you need to scrounge enough "juice" to make it to the next one. Some stops are friendly. Most aren't.

The Monkey Lion is out there. You just need to survive long enough to find it.

---

## Core Loop

```
[JUMP] ──► [STRANDED] ──► [FIND CAMP] ──► [SURVIVE/GATHER] ──► [SOLVE EPISODE] ──► [JUMP]
   │                                              │
   │                                              ▼
   │                                    [UPGRADE SHUTTLE]
   │                                    [UPGRADE SKILLS]
   │                                              │
   └──────────────────────────────────────────────┘
```

### 1. JUMP
- Jumping costs nearly all your drive fuel (let's call it "plasma" or "jump juice")
- You arrive in a new system with ~5-10% fuel remaining
- Can't jump again until you refuel
- Each jump gets you closer to (or further from) the Monkey Lion

### 2. STRANDED
- You're stuck in this system until you solve its problem
- Fly around, scan planets, assess threats
- Find somewhere safe(ish) to set up camp

### 3. FIND CAMP
- Different locations have different pros/cons:
  - **Planet surface**: Safe from pirates, but environment hazards (need gear)
  - **Station**: Safe and comfortable, but costs credits or favors
  - **Asteroid/moon**: Hidden, good for scavenging, no atmosphere
  - **Derelict**: Free shelter, but what killed the crew?
  - **Open space**: Dangerous - pirates, radiation, no resources

### 4. SURVIVE/GATHER
- Time passes while camping
- Random events trigger (good and bad)
- Gather resources from environment
- Repair shuttle, craft supplies
- Rest to recover (hunger, thirst, fatigue)

### 5. SOLVE EPISODE
- Each system has a "problem" that gates your fuel
- Could be:
  - Derelict with fuel cells to salvage
  - Station that trades fuel for cargo/services
  - Mining operation you can work for fuel
  - Pirate blockade you need to sneak/fight past
  - Ancient ruins with power cores
- Completing the episode rewards enough fuel for next jump (plus extras)

### 6. UPGRADE
- Better shuttle = camp in harsher places
- Better skills = handle more situations
- Better gear = survive longer, gather more

---

## Resource Model

### Jump Fuel (Primary Gate)
- **Max capacity**: Starts at 100 units
- **Jump cost**: ~90-95 units per jump
- **Sources**: Salvage, trade, quest rewards, rare mining
- This is THE bottleneck - everything revolves around getting more

### Ship Resources (Daily Survival)
- **Energy**: Powers systems, shields, life support
- **Water**: Drinking, hygiene, coolant
- **Organics**: Food, medical supplies
- **Hull**: Damage from hazards, combat

### Personal Resources
- **Hunger/Thirst/Hygiene**: The body sim (already built)
- **Fatigue**: New - need to sleep, camping helps
- **Health**: Injuries from combat, hazards

### Trade Resources
- **Credits**: Universal currency
- **Cargo**: Stuff you find to sell or use
- **Reputation**: With factions (pirates, traders, etc.)

---

## Camping Mechanics

### Setting Up Camp
1. Land on planet / dock at station / anchor to asteroid
2. Choose "Make Camp" option
3. Time begins passing (can choose duration)
4. Events roll based on location type + duration + threat level

### Camp Activities (While Time Passes)
- **Rest**: Recover fatigue, heal minor injuries
- **Repair**: Fix hull, systems (costs materials)
- **Scavenge**: Gather local resources (depends on location)
- **Craft**: Make supplies from raw materials
- **Research**: Study scans, decrypt data, learn clues
- **Wait**: Just pass time (sometimes you need to wait out danger)

### Camp Events (Random)
- **Neutral**: Weather, wildlife, passing ships
- **Good**: Find cache, friendly traveler, resource node
- **Bad**: Equipment failure, creature attack, pirates find you
- **Story**: Clue about Monkey Lion, distress signal, mystery

### Location Modifiers
| Location | Safety | Resources | Events | Requirements |
|----------|--------|-----------|--------|--------------|
| Temperate Planet | High | Medium | Calm | None |
| Ice Planet | Medium | Low | Weather | Thermal Gear |
| Volcanic Planet | Low | High | Hazards | Heat Shielding |
| Station | Very High | None | Social | Credits |
| Asteroid | Medium | Mining | Sparse | EVA Suit |
| Derelict | Variable | Salvage | Creepy | Courage |
| Open Space | Dangerous | None | Pirates | Weapons |

---

## Upgrade Trees

### Shuttle Upgrades
- **Hull Plating**: Take more damage
- **Thermal Shielding**: Camp on volcanic/ice worlds
- **Cargo Expansion**: Carry more stuff
- **Fuel Tank**: Store more jump juice
- **Stealth Plating**: Avoid pirates
- **Weapons**: Fight back
- **Scanner Range**: See more of system
- **Life Support**: Longer trips, less resource drain

### Character Skills (Already Exists - Expand)
- **Piloting**: Better flight, fuel efficiency
- **Engineering**: Repair efficiency, crafting
- **Science**: Better scans, research speed
- **Combat**: Fighting ability
- **Survival**: Resource gathering, camping events
- **Social**: Better trades, NPC interactions

### Gear (Personal Equipment)
- **EVA Suit**: Required for asteroids, derelicts
- **Thermal Gear**: Ice/volcanic planets
- **Weapon**: Combat encounters
- **Scanner**: Find resources while camping
- **Medkit**: Heal injuries
- **Rations**: Emergency food/water

---

## Starting Scenario

### Prologue (Skippable after first run)
- Wake up on a planet/station
- Shuttle is broken, need to fix it
- Tutorial teaches: movement, interaction, resource management
- Mini-quest to gather parts and repair shuttle
- First jump is the "leaving home" moment

### Starting State
- Shuttle: Damaged but functional, minimal upgrades
- Fuel: Exactly enough for one jump
- Resources: Basic supplies for a few days
- Skills: Level 1 across the board
- Gear: Basic EVA suit, no thermal gear
- Credits: Almost none
- Clue: Vague memory of Monkey Lion's last heading

---

## Goals (Sandbox Style)

The game is a sandbox with optional long-term goals. Play your way.

### Goal 1: Find the Monkey Lion
The "main quest" for those who want one.
- **What You Know**: You were crew, something happened, you got stranded
- **How You Find It**: Clues in derelicts, stations, ruins. Star charts. Survivor accounts.
- **Endgame**: Piece together the ML's location, reach it, discover what happened

### Goal 2: Find Home
Get back to Earth / Federation space / wherever "home" is.
- **What You Know**: Rough direction, many jumps away
- **How You Get There**: Accumulate enough fuel/upgrades for the long haul
- **Endgame**: Finally reach safe harbor, retire from the nomad life

### Goal 3: Become the Pirate King
If you can't beat 'em...
- **How**: Build rep with pirates, take jobs, raid ships
- **Endgame**: Control a sector, have minions, live dangerously

### Goal 4: Build an Empire
Settle down, but in space.
- **How**: Claim a station or planet, build it up, trade routes
- **Endgame**: Become a merchant prince / station boss

### Goal 5: Just Survive
No goal. Pure sandbox.
- **How**: Keep camping, keep jumping, see what happens
- **Endgame**: There is none. The journey is the destination.

### Discovery System (Supports All Goals)
- **Clues**: Found everywhere - derelicts, stations, planets, NPCs
- **Star Charts**: Piece together the galaxy, mark points of interest
- **Logs/Data**: Captain's logs, sensor data, personal journals
- **Reputation**: Opens doors (or closes them) based on your choices

---

## What Exists vs What's Needed

### Already Built (Keep/Adapt)
- [x] Ship interior exploration
- [x] Resource management (water, organic, energy)
- [x] Body simulation (hunger, thirst, hygiene)
- [x] System map flight
- [x] Planet scanning
- [x] Surface exploration
- [x] Episode/encounter system
- [x] Skill system
- [x] Cargo system

### Needs Modification
- [ ] Jump fuel as separate critical resource
- [ ] Starting scenario (not in space)
- [ ] Energy → split into ship energy vs jump fuel
- [ ] Episode triggers (camp events, not just random)

### Needs Building
- [ ] Camping mechanic (time passage, event rolls)
- [ ] Camp activity menu
- [ ] Upgrade system (shuttle parts)
- [ ] Gear/equipment system
- [ ] Clue/discovery system for ML
- [ ] Threat level per location
- [ ] Fatigue system
- [ ] Starting planet/station map

---

## Implementation Phases

### Phase 1: Jump Fuel Gate
- Add JumpFuel resource separate from Energy
- Jumping costs JumpFuel
- Can't jump with insufficient fuel
- Basic fuel sources (salvage, quest rewards)

### Phase 2: Camping Core
- "Make Camp" action when landed/docked
- Time passage with basic events
- Resource gathering while camping
- Fatigue system + rest mechanic

### Phase 3: Location Threat Levels
- Different camp locations have different risks
- Environment requirements (thermal gear, etc.)
- Camp events based on location type

### Phase 4: Upgrade System
- Shuttle upgrade tree
- Gear/equipment slots
- Upgrade acquisition (buy, craft, find)

### Phase 5: Starting Scenario
- New game starts on planet/station
- Shuttle repair tutorial
- First jump narrative

### Phase 6: Monkey Lion Trail
- Clue system
- Star chart reconstruction
- Story beats and endings

---

## The Vibe

> Your shuttle's reactor hums in the darkness. Outside, ice crystals ping against the thermal shielding you jury-rigged from a dead miner's suit. The scanner says there's a fuel depot 3 klicks north, buried under the glacier. Your thermal gear will last maybe an hour in this cold.
>
> You could wait for the storm to pass. But that pirate frigate is still in orbit, and they're going to start scanning the surface eventually.
>
> You check the scanner one more time. 3 klicks north. Under the glacier.
>
> The Monkey Lion is out there somewhere. You're not dying on this frozen rock.

---

## Open Questions

1. **Permadeath or persistent?** Roguelike runs vs meta-progression?
2. **Galaxy size?** Fixed map or procedural infinite? How far is "far"?
3. **Combat depth?** Simple (resolve via stats) or tactical (real-time/turn-based)?
4. **NPC crew?** Solo forever or recruitable crew members?
5. **Ship progression?** Upgrade shuttle forever or eventually get a bigger ship?
6. **Faction depth?** Simple reputation or complex politics?
7. **Base building?** Can you actually settle somewhere permanently?
8. **Time pressure?** Is there a ticking clock or pure sandbox pace?
