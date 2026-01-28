# SpaceHole Camper

Space van life roguelike. Scrounge enough fuel for one more jump, camp on hostile worlds, and chase the ghost of a ship you were once crew on.

```
    *  .  *       .    *         .      *
  .    ____     .         *    .           *
    .-'    '-.      *           .    *
   /  SPACEHOLE  \    .    *         .
  |    CAMPER    |       .      *        .
   \  ________  /   *         .     *
    '-.____.--'        .   *      .
  .    *     .    *        .   *     .
```

You were a redshirt on the legendary USS Monkey Lion. Something happened. Now you're stranded with a busted shuttle, barely enough fuel for one jump, and a vague memory of where the ship was headed.

Your shuttle is your camper van. Each star system is a pit stop where you need to scrounge enough "juice" to reach the next one. Camp on planets, scavenge derelicts, dodge pirates, and piece together the trail back to the Monkey Lion.

## Requirements

- Go 1.24 or later
- A terminal that supports 16-color ANSI

### Platform-specific

**Windows**: Works out of the box.

**Linux**: You may need X11 development libraries:
```bash
# Debian/Ubuntu
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

**macOS**: Works out of the box (requires Xcode command line tools).

## Installation

Clone and build:

```bash
git clone https://github.com/spacehole-rogue/spacehole_rogue.git
cd spacehole_rogue
go build -o spacehole ./cmd/spacehole
```

Or install directly:

```bash
go install github.com/spacehole-rogue/spacehole_rogue/cmd/spacehole@latest
```

## Running

```bash
./spacehole
```

## Core Loop

1. **Jump** into a new system (uses most of your fuel)
2. **Scout** the system - scan planets, check for threats
3. **Find a camp** - planet surface, station, asteroid, derelict
4. **Survive** - gather resources, repair, rest, handle events
5. **Solve the episode** - complete the system's quest for fuel
6. **Upgrade** - better shuttle, better gear, better odds
7. **Jump** again - toward whatever goal you've chosen

## Goals (Pick Your Own)

- **Find the Monkey Lion** - piece together clues, retrace your steps
- **Find Home** - navigate back to Earth / Federation space
- **Become a Pirate King** - join them, lead them, raid ships
- **Build an Empire** - claim a station, establish trade routes
- **Just Survive** - no goal, pure sandbox, the journey is the point

## Controls

### Ship Interior
| Key | Action |
|-----|--------|
| WASD / Arrows | Move |
| E | Interact |
| T | Toggle equipment |
| Tab | Character sheet |
| ESC | Quit |

### System Map
| Key | Action |
|-----|--------|
| WASD / Arrows | Fly |
| E | Interact (orbit, dock, scan) |
| N | Sector navigation |
| Tab | Character sheet |
| ESC | Return to ship |

### Surface / Camps
| Key | Action |
|-----|--------|
| WASD / Arrows | Move |
| E | Interact |
| Tab | Character sheet |

## Survival

Your shuttle needs:
- **Jump Fuel** - can't leave the system without it
- **Energy** - powers systems and life support
- **Water** - drinking, hygiene, coolant
- **Organics** - food and medical supplies

You need:
- **Food & Water** - or you starve
- **Hygiene** - or morale tanks
- **Rest** - or fatigue catches up

Where you camp matters:
- **Planets** need the right gear (thermal for ice/volcanic)
- **Stations** are safe but cost credits
- **Derelicts** have salvage but... what killed the crew?
- **Open space** means pirates will find you

## Status

Early development. Working systems:
- Ship interior with resource management
- System map flight and scanning
- Planet surface exploration
- Episode/encounter framework
- Skill progression

See [planning/CAMPER_DESIGN.md](planning/CAMPER_DESIGN.md) for the full design doc.

---

*"Your thermal gear will last maybe an hour in this cold. The pirate frigate is still in orbit. You check the scanner one more time. The fuel depot is 3 klicks north, under the glacier. You're not dying on this frozen rock."*
