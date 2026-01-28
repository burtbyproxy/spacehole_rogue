# SpaceHole Rogue

ASCII space roguelike in Go. Start as nobody, find the legendary USS Monkey Lion, bring it home.

```
    *  .  *       .    *         .      *
  .    ____     .         *    .           *
    .-'    '-.      *           .    *
   /  SPACEHOLE  \    .    *         .
  |    ROGUE     |       .      *        .
   \  ________  /   *         .     *
    '-.____.--'        .   *      .
  .    *     .    *        .   *     .
```

Star Trek parody tone. A zebra named Deborah is your Chief Science Officer. She has no idea what she's doing.

## Requirements

- Go 1.24 or later
- A terminal that supports 16-color ANSI (most modern terminals)

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

Clone the repository and build:

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

Or on Windows:
```
spacehole.exe
```

## Controls

### Ship Interior
| Key | Action |
|-----|--------|
| WASD / Arrows | Move |
| E | Interact with equipment |
| T | Toggle equipment on/off |
| Tab | Character sheet |
| ESC | Quit |

### System Map (Flying)
| Key | Action |
|-----|--------|
| WASD / Arrows | Fly shuttle |
| E | Interact (orbit planet, dock station, scan) |
| N | Open sector navigation map |
| Tab | Character sheet |
| ESC | Return to ship |

### Sector Map (Navigation)
| Key | Action |
|-----|--------|
| WASD / Arrows | Move cursor |
| E / Enter | Jump to selected system |
| ESC | Cancel |

### Surface Exploration
| Key | Action |
|-----|--------|
| WASD / Arrows | Move |
| E | Interact (terminals, crates, shuttle) |
| Tab | Character sheet |
| ESC | Info |

## Gameplay Loop

1. Explore your ship, manage resources (water, organics, energy)
2. Use the pilot station to fly around the star system
3. Scan planets from orbit using the science station
4. Land on planets with points of interest
5. Complete objectives, loot crates, return to shuttle
6. Jump to new systems via the nav console
7. Find clues about the legendary USS Monkey Lion

## Status

Early development. Core systems working:
- Ship interior exploration with resource management
- System map flight with planets, stations, derelicts
- Planet scanning and orbital mechanics
- Surface exploration with procedural terrain and structures
- Encounter system with hails and episodes
- Skill progression

---

*"The Deborah in Cargo Bay B is eating your supplies. She is a zebra. She doesn't understand inventory management."*
