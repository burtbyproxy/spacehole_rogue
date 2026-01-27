# SpaceHole Rogue — Game Design Document

## Overview

ASCII space roguelike in Go. Single-player. You start as nobody — crash landed, waking from cryo, surviving a razed colony. Work your way onto a shuttle, then a ship. Explore a sector, gather clues about the legendary lost USS Monkey Lion and its SpaceHole Drive. Finding the ML is Chapter 1. The main game is life aboard the ML as it warps sector to sector, triggering episodic encounters. Goal: bring it home.

The game should FEEL like a hardcore simulation but PLAY like a fun roguelike. The goofy Star Trek parody tone is the spoonful of sugar. A zebra named Deborah is your Chief Science Officer — she has no idea what she's doing, but the ship's sensors somehow still work. That's the vibe.

---

## Design Pillars

### 1. The Game Teaches Itself Through Ship Scale
No tutorials, no manual, no community college degree required. The shuttle has 4 rooms and 3 pipes — you learn water/energy/organic by keeping a space van alive. The big ship has 50 crew handling 200 pipes — you just get alerts when things break.

### 2. Automate the Routine, Dramatize the Exceptions
A healthy ship is boring (on purpose). Crew walks around, pipes flow, lights stay on. When something goes wrong, the game SHOWS you — red blinking tiles, panicked crew chatter, alarms. Your job is to react to crises, not babysit systems.

### 3. "One More System" is the Addiction
You can always see the next star on the map. What's there? A 30-second jump to find out. Maybe it's a derelict with loot. Maybe it's a Lonely Godling who manifests your crew's thoughts as bad theater. You won't know until you go.

### 4. Crew You Care About
You rescued a Rashean from a prison planet. She panics in combat and accidentally shoots the ceiling, but her bulletproof shell saved the away team. The Deborah in engineering keeps the warp core running despite being a zebra. When a redshirt dies, it hurts because you saw them eating lunch in the mess hall yesterday.

### 5. Meaningful Scarcity
Water at 30%. No stations nearby. An ice asteroid on sensors, but it's in hostile space. That tension — do I risk it? — is the engine that drives every decision.

---

## Progressive Complexity — The Game Teaches Itself

The same simulation systems run at every scale. The difference is how much the PLAYER has to manage vs. how much the CREW handles.

### Phase 1: The Shuttle (Solo Survival)
You are alone in a space van.

- 4 rooms: cockpit, bunk, storage, engine room
- 1 deck, ~15x10 tiles. You can see the whole thing without scrolling.
- Matter system is TINY: 1 water tank, 1 power cell, 1 food locker, 1 hull
- 3 pipes total. If one breaks, you walk 3 tiles to fix it.
- YOU do everything: pilot, repair, trade, fight, scavenge
- The HUD shows 4 bars: Water, Energy, Organic, Hull. That's it.

The player learns: "Oh, water goes down. I need to get more. That station sells it. I need credits. That derelict has salvage." The entire resource loop clicks in 10 minutes.

**Feels like:** early Subnautica, FTL's resource tension, Space Trucker Simulator

### Phase 2: Small Ship (Crew of 5-10)
You join a ship's crew. You have a role.

- The captain assigns you a station (engineering, science, security, etc.)
- 2-3 decks, ~30x20 tiles per deck. Turbolifts between them.
- 15-30 pipes. You CAN'T check them all. But the other engineers can.
- The game teaches delegation: you see crew doing their jobs.
- You do YOUR job, plus help when there's a crisis
- Crew relationships form. The pilot is your buddy. The security chief is an ass.
- You get PROMOTED based on skill and performance.

The player learns: "I don't have to do everything. The crew handles routine stuff. I focus on my role and help with emergencies."

**Feels like:** being a junior officer in Star Trek, FTL crew management

### Phase 3: The USS Monkey Lion (Crew of 50+)
You're a senior officer. Maybe eventually captain.

- 8-12 decks. Huge ship. Minimap essential.
- 200+ pipes, 50+ crew, full ship subsystems
- You manage through ALERTS, not direct observation.
- You make COMMAND decisions: where to go, which episodes to engage, resource allocation
- You CAN still walk around, fix pipes yourself, fight on away teams. But you don't HAVE to.
- The SH Drive warps to new sectors — each jump is a new frontier

The player learns: "I understand all these systems from the shuttle and small ship. Now I'm orchestrating them."

**Feels like:** being Captain Kirk, FTL at epic scale, the payoff for everything you learned

---

## Core Gameplay Loops

### Micro Loop (every 5-10 seconds)
See something → Walk to it → Interact → Get feedback

- See a blinking red pipe → Walk to it → Press E to repair → Pipe fixed, water restored
- See a crew member → Walk to them → Press E to talk → Learn gossip, get a quest
- See loot on a derelict → Walk to it → Press E to pick up → Inventory updated

### Session Loop (15-30 minutes)
Pick destination → Travel → Arrive → Episode → Reward → Maintain → Pick next destination

1. **Pick destination** — Look at sector map. That unexplored system has a question mark.
2. **Travel** — En route, stuff happens. A pipe breaks. Crew gossip. Random encounter?
3. **Arrive** — Episode triggers: "Investigate + Derelict Ship + Crew Infected + Deranged Scientist"
4. **Episode** — Board the derelict. Find the scientist. Discover the crew is infected. Make choices.
5. **Reward** — Salvaged tech, credits, a new crew member, an ML clue
6. **Maintain** — Dock at nearest station. Repair, resupply, trade, recruit.
7. **Repeat** — What's in the next system?

The "one more system" hook: Each system is a ~10-20 minute episode. Quick enough to always say "just one more."

### Campaign Loop (across many sessions)
Explore sector → Gather clues → Upgrade ship/crew → Find the ML → Warp to new sectors → Get home

- **Chapter 1:** ~5-10 hours of shuttle → small ship → find the ML
- **Chapter 2+:** Open-ended episodic gameplay aboard the ML. Each SH Drive jump is a new sector.
- **Long-term goals:** bring the ML home, build the best crew, discover all races, max skills

---

## What Makes It Feel "Deep" Without Being Complex

### The Illusion of Depth
The matter system has 4 types, push-based flow, and pipe entities. That's actually simple. But it LOOKS complex because:

- You can walk through the ship and SEE pipes visually in the tile map
- The HUD shows 4 resource bars fluctuating in real-time
- Crew members walk to repair things and you can follow them
- The message log narrates what's happening

The player thinks "wow, this is a deep simulation" but the actual rules are: stuff flows from A to B, treatment converts dirty to clean, broken pipes leak. Three rules.

### The Humor Does Heavy Lifting
When the game says "The Deborah in Cargo Bay B is eating your supplies. She is a zebra. She doesn't understand inventory management." — that's:

- **Funny** — keeps the tone light
- **Communicates game state** — your organic matter is being consumed
- **Suggests action** — maybe assign someone to cargo bay
- **Builds character** — you now know Deborah #3 and her deal

The parody tone makes complex systems approachable because you're laughing while you learn.

### Visual Language
A single glance at the screen should tell you the ship's health:

```
GOOD ship:                      BAD ship:
  . . . . . .                    . . ~ . . .      ← red blinking pipe
  . c . . c .                    . c . . ! .      ← crew with problem marker
  ─ ─ ─ ─ ─ ─  (blue pipes)     ─ ~ ─ ─ ─ ~     ← leaking pipes
  . . = . . .                    . . = . . .
  Water: ████████ 89%            Water: ███░░░░░ 31%  ← yellow/red
  Energy: ███████ 95%            Energy: ██░░░░░░ 22%  ← RED
```

No numbers to read. No menus to check. Colors tell the story.

---

## Fun Hooks — Why Players Keep Playing

### 1. The Loot Chase
Derelicts, debris fields, defeated ships — all have randomized loot. Better weapons, rare tech, exotic matter, crew members frozen in cryo pods.

### 2. Crew Attachment Engine
Crew members have names, personalities, relationships, and visible behavior. Over time, you form opinions about them. When the game kills one in an episode, it should sting.

### 3. The Upgrade Treadmill
- Shuttle: basic → improved → advanced shuttle
- Small ship: join as crew → earn better quarters → get your own shuttle in the bay
- ML: restore damaged systems → unlock new decks → fully operational legendary vessel
- Equipment: always something better to find/buy/craft
- Skills: always another level to reach, another perk to unlock

### 4. The Clue Trail (Chapter 1)
ML clues are scattered across the sector. Some from NPCs, some from derelicts, some from episodes. Each clue narrows the search. The player pieces together the mystery like a detective.

### 5. The Episode Machine
The 4-table episode generator creates thousands of unique combinations. Most players will never see the same episode twice. The twist system means even routine missions go sideways.

### 6. Risk/Reward Decisions
- "That derelict might have amazing loot, but it's in Species 7482 territory."
- "We can trade water for credits, but we only have enough for 3 more days."
- "The shortcut through the nebula saves fuel but might damage sensors."

Every decision should have a clear upside AND a visible cost.

---

## Story Structure

### The Setup
You were a redshirt on the USS Monkey Lion — the legendary ship with the SpaceHole Drive. You were on an away mission when everything went wrong. The team got separated. The ML jumped. You're stranded with nothing.

### Chapter 1: The Prologue — Assemble the Crew, Find the Ship

**Phase 1: Survival (Shuttle gameplay)**
Wake up stranded. Find or repair a shuttle. Learn to survive solo.

**Phase 2: Find Your People (Quest line)**
You know key ML crew members were scattered across the sector. Each one is in a different system, in trouble, needing rescue:

- The Chief Engineer is trapped in a mining base, keeping the reactor from melting down
- The Science Officer (a Deborah) is being worshipped as a god by a primitive civilization
- The Tactical Officer is in a prison, because of course she is
- The Navigator is working for an Eccentric Trader, plotting routes for credits
- The Medical Officer is treating plague victims at a colony, refusing to leave

Each rescue is its own mini-episode. As you find crew, they join you on the shuttle.

**Phase 3: Find the Monkey Lion**
Once key crew are assembled, they help piece together where the ML jumped. Final quest: locate and board the ML. The ML is damaged, understaffed, drifting — it needs you and your assembled crew.

### Chapter 2+: The Main Game — Aboard the Monkey Lion
- Restore the ship's systems, recruit more crew at stations
- The SH Drive warps between sectors — each jump generates a new sector
- Episodic gameplay: explore, trade, fight, manage resources, survive
- Goal: bring the USS Monkey Lion home

### The USS Monkey Lion
- Only ship in the galaxy with the SpaceHole Drive
- SH Drive components: Pattern Buffer → Buserd Injector → Obscuration Chamber → Pattern Buffer 2 → Finalization Thruster
- The drive projects dark matter as a "space hole" in front of the ship
- When you find it, it's damaged and understaffed — restoring it IS the gameplay of early Chapter 2