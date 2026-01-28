package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/spacehole-rogue/spacehole_rogue/assets"
	"github.com/spacehole-rogue/spacehole_rogue/internal/game"
	"github.com/spacehole-rogue/spacehole_rogue/internal/render"
	"github.com/spacehole-rogue/spacehole_rogue/internal/world"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	title        = "SpaceHole Rogue"

	cellWidth  = 16
	cellHeight = 16
	gridCols   = screenWidth / cellWidth   // 80
	gridRows   = screenHeight / cellHeight // 45
)

const (
	// Viewport center — where the player @ always appears on screen.
	viewCenterX = gridCols / 2 // 40
	viewCenterY = 13

	// Fixed UI positions
	panelX   = 60 // right-side legend/info panel
	hudRow   = 27 // systems HUD
	commsRow = 33 // message log
	commsMax = 8  // max visible messages
)

// ViewMode controls which screen is displayed.
type ViewMode int

const (
	ViewShip ViewMode = iota
	ViewSectorMap
	ViewSystemMap
	ViewStation
	ViewCargo
	ViewCharSheet
	ViewEncounter
	ViewEpisode
	ViewSurface
)

// Station submenu states.
const (
	stMenuMain     = 0
	stMenuRepairs  = 1
	stMenuTrade    = 2
	stMenuBuy      = 3
	stMenuSell     = 4
	stMenuBar      = 5
	stMenuFaction  = 6
)

// floatingSprite is a glyph drawn at sub-pixel screen coordinates,
// for entities that move smoothly between tile cells.
type floatingSprite struct {
	Glyph  byte
	FG     uint8
	PX, PY float64
}

// Game is the Ebitengine game struct. It owns rendering and input.
// All gameplay state lives in sim.
type Game struct {
	atlas    *render.FontAtlas
	renderer *render.GridRenderer
	buffer   *render.CellBuffer
	sim      *game.Sim
	viewMode ViewMode
	sprites  []floatingSprite // sub-tile sprites drawn on top of CellBuffer

	// Character sheet
	prevViewMode ViewMode // view to return to from char sheet

	// Station docking state
	stationMenu int               // current station submenu (stMenu* constants)
	stationData *game.StationData // current docked station (nil when not docked)
}

func NewGame() *Game {
	atlas := render.NewFontAtlas()
	renderer := render.NewGridRenderer(atlas, cellWidth, cellHeight)
	buffer := render.NewCellBuffer(gridCols, gridRows)

	// Load shuttle layout from embedded assets
	data, err := assets.Ships.ReadFile("ships/shuttle.json")
	if err != nil {
		log.Fatalf("load shuttle: %v", err)
	}
	layout, err := world.LoadShipLayout(data)
	if err != nil {
		log.Fatalf("parse shuttle: %v", err)
	}

	// Generate a seed from current time for this run
	seed := time.Now().UnixNano()
	sim := game.NewSimWithPrologue(layout, seed)

	g := &Game{
		atlas:    atlas,
		renderer: renderer,
		buffer:   buffer,
		sim:      sim,
		viewMode: ViewSurface, // Start on surface during prologue
	}

	g.drawScreen()
	return g
}

// cameraX/Y returns the world-to-screen offset so the player is always centered.
func (g *Game) cameraX() int {
	px, _ := g.sim.PlayerPos()
	return viewCenterX - px
}
func (g *Game) cameraY() int {
	_, py := g.sim.PlayerPos()
	return viewCenterY - py
}

// --- Draw dispatch ---

func (g *Game) drawScreen() {
	g.sprites = g.sprites[:0] // clear floating sprites; only system map populates them
	switch g.viewMode {
	case ViewSectorMap:
		g.drawSectorMapView()
	case ViewSystemMap:
		g.drawSystemMapView()
	case ViewStation:
		g.drawStationView()
	case ViewCargo:
		g.drawCargoView()
	case ViewCharSheet:
		g.drawCharSheetView()
	case ViewEncounter:
		g.drawEncounterView()
	case ViewEpisode:
		g.drawEpisodeView()
	case ViewSurface:
		g.drawSurfaceView()
	default:
		g.drawShipView()
	}
}

func (g *Game) drawShipView() {
	buf := g.buffer
	buf.Clear()

	// --- World layer (camera-relative) ---
	ox := g.cameraX()
	oy := g.cameraY()
	render.RenderTileGrid(buf, g.sim.Grid, ox, oy)

	// Player — always at viewport center
	buf.Set(viewCenterX, viewCenterY, '@', render.ColorWhite, render.ColorBlack)

	// --- Fixed UI layer ---

	// Title bar
	buf.WriteString(2, 0, title, render.ColorWhite, render.ColorBlack)
	nameLabel := fmt.Sprintf("[ %s ]", g.sim.Layout.Name)
	buf.WriteString(20, 0, nameLabel, render.ColorLightCyan, render.ColorBlack)
	// Show current star system
	curStar := g.sim.Sector.Systems[g.sim.Sector.CurrentSystem]
	buf.WriteString(gridCols-len(curStar.Name)-4, 0, curStar.Name, render.ColorYellow, render.ColorBlack)

	// Legend (fixed right panel)
	buf.WriteString(panelX, 3, "Legend:", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(panelX, 4, " # Hull  + Door", render.ColorLightGray, render.ColorBlack)
	row := 5
	legendItem := func(glyph byte, fg uint8, label string) {
		buf.Set(panelX+1, row, glyph, fg, render.ColorBlack)
		buf.WriteString(panelX+2, row, " "+label, fg, render.ColorBlack)
		row++
	}
	legendItem('=', render.ColorLightCyan, "Nav Station")
	legendItem('=', render.ColorLightGreen, "Pilot Station")
	legendItem('=', render.ColorLightMagenta, "Science Station")
	legendItem('=', render.ColorBrown, "Cargo Console")
	legendItem('-', render.ColorLightCyan, "Viewscreen")
	legendItem('$', render.ColorLightGreen, "Food Replicator")
	legendItem('$', render.ColorLightBlue, "Drink Replicator")
	legendItem('$', render.ColorLightCyan, "Medical")
	legendItem('*', render.ColorRed, "Incinerator")
	legendItem(254, render.ColorGreen, "Organic Tank")
	legendItem(254, render.ColorBlue, "Water Tank")
	legendItem(254, render.ColorBrown, "Battery")
	legendItem(177, render.ColorLightMagenta, "Recycler")
	legendItem(177, render.ColorBrown, "Generator")
	legendItem('%', render.ColorBrown, "Engine")

	// HUD matter bars (live from sim)
	r := &g.sim.Resources
	buf.WriteString(2, hudRow, "--- Matter ---", render.ColorLightCyan, render.ColorBlack)
	drawMatterBar(buf, 2, hudRow+1, "Water  ", &r.Water, render.ColorBlue, render.ColorDarkGray)
	drawMatterBar(buf, 2, hudRow+2, "Organic", &r.Organic, render.ColorGreen, render.ColorDarkGray)
	drawEnergyBar(buf, 2, hudRow+3, "Energy ", r.Energy, r.MaxEnergy, render.ColorYellow)
	drawSimpleBar(buf, 2, hudRow+4, "Hull   ", r.Hull, r.MaxHull, render.ColorLightGray)
	// Credits and cargo
	buf.WriteString(2, hudRow+5, fmt.Sprintf("Credits: %d  Cargo: %d/%d pads",
		r.Credits, r.PadsUsed(), len(r.CargoPads)), render.ColorLightCyan, render.ColorBlack)
	// Player body indicator
	fullness := r.BodyFullness()
	if fullness > 0 {
		bodyClr := uint8(render.ColorDarkGray)
		if r.TotalWaste() >= 20 {
			bodyClr = render.ColorYellow
		}
		buf.WriteString(2, hudRow+6, fmt.Sprintf("Body: %d/%d full (%d waste)",
			fullness, game.MaxBodyFullness, r.TotalWaste()), bodyClr, render.ColorBlack)
	}

	// Equipment status (right panel, below legend)
	row++ // gap
	buf.WriteString(panelX, row, "--- Equipment ---", render.ColorLightCyan, render.ColorBlack)
	row++
	drawEquipStatus(buf, panelX, row, "Engine", g.sim.EngineOn)
	row++
	drawEquipStatus(buf, panelX, row, "Generator", g.sim.GeneratorOn)
	row++
	drawEquipStatus(buf, panelX, row, "Recycler", g.sim.RecyclerOn)
	row++

	// Player needs (right panel, below equipment)
	n := &g.sim.Needs
	row++ // gap
	buf.WriteString(panelX, row, "--- Status ---", render.ColorLightCyan, render.ColorBlack)
	row++
	drawNeedBar(buf, panelX, row, "Hunger ", n.Hunger)
	row++
	drawNeedBar(buf, panelX, row, "Thirst ", n.Thirst)
	row++
	drawNeedBar(buf, panelX, row, "Hygiene", n.Hygiene)
	row += 2

	// Standing on indicator
	px, py := g.sim.PlayerPos()
	tile := g.sim.Grid.Get(px, py)
	buf.WriteString(panelX, row, "Standing on:", render.ColorDarkGray, render.ColorBlack)
	row++
	buf.WriteString(panelX+1, row, tile.Describe(), render.ColorLightGray, render.ColorBlack)

	// Message log (live from sim)
	// Blinking hail alert when pending hail exists
	if g.sim.PendingHail != nil && (g.sim.Ticks/30)%2 == 0 {
		buf.WriteString(2, commsRow, ">>> INCOMING HAIL <<<", render.ColorYellow, render.ColorBlack)
	} else {
		buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	}
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}

	// Instructions
	buf.WriteString(2, gridRows-1, "WASD: Move  E: Interact  T: Toggle  Tab: Status  ESC: Quit", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawSectorMapView() {
	buf := g.buffer
	buf.Clear()

	sec := g.sim.Sector
	cur := sec.Systems[sec.CurrentSystem]
	sel := sec.Systems[sec.CursorSystem]

	// Title bar
	buf.WriteString(2, 0, "--- Sector Map ---", render.ColorLightCyan, render.ColorBlack)
	curLabel := fmt.Sprintf("Location: %s", cur.Name)
	buf.WriteString(gridCols-len(curLabel)-2, 0, curLabel, render.ColorYellow, render.ColorBlack)

	// Draw connection line from current to cursor (if different)
	if sec.CursorSystem != sec.CurrentSystem {
		drawJumpLine(buf, cur.X, cur.Y, sel.X, sel.Y)
	}

	// Draw stars
	for i, sys := range sec.Systems {
		clr := starColor(sys.Type)

		glyph := byte('.')
		if sys.Visited {
			glyph = '*'
		}

		if i == sec.CurrentSystem {
			// Player's current system — show shuttle marker
			buf.Set(sys.X, sys.Y, '^', render.ColorWhite, render.ColorBlack)
		} else {
			buf.Set(sys.X, sys.Y, glyph, clr, render.ColorBlack)
		}

		// Cursor brackets around selected system
		if i == sec.CursorSystem {
			buf.Set(sys.X-1, sys.Y, '[', render.ColorYellow, render.ColorBlack)
			buf.Set(sys.X+1, sys.Y, ']', render.ColorYellow, render.ColorBlack)
			// Show name near selected star
			nameX := sys.X + 3
			if nameX+len(sys.Name) > panelX-1 {
				nameX = sys.X - len(sys.Name) - 2
			}
			buf.WriteString(nameX, sys.Y, sys.Name, render.ColorWhite, render.ColorBlack)
		}
	}

	// Info panel (right side)
	infoX := panelX
	buf.WriteString(infoX, 3, "--- Selected ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(infoX, 4, sel.Name, render.ColorWhite, render.ColorBlack)
	buf.WriteString(infoX, 5, game.StarTypeName(sel.Type), starColor(sel.Type), render.ColorBlack)

	if sel.Visited {
		buf.WriteString(infoX, 6, "Visited", render.ColorDarkGray, render.ColorBlack)
	} else {
		buf.WriteString(infoX, 6, "Unexplored", render.ColorLightGreen, render.ColorBlack)
	}

	if sec.CursorSystem != sec.CurrentSystem {
		cost := sec.EnergyCostTo(sec.CursorSystem)
		costClr := uint8(render.ColorLightGray)
		if cost > g.sim.Resources.Energy {
			costClr = render.ColorLightRed
		}
		buf.WriteString(infoX, 8, fmt.Sprintf("Jump cost: %d energy", cost), costClr, render.ColorBlack)
		buf.WriteString(infoX, 9, fmt.Sprintf("Available: %d", g.sim.Resources.Energy), render.ColorYellow, render.ColorBlack)
	} else {
		buf.WriteString(infoX, 8, "You are here.", render.ColorYellow, render.ColorBlack)
	}

	// Ship status summary
	r := &g.sim.Resources
	buf.WriteString(infoX, 11, "--- Ship ---", render.ColorLightCyan, render.ColorBlack)
	drawEnergyBar(buf, infoX, 12, "Energy ", r.Energy, r.MaxEnergy, render.ColorYellow)
	drawSimpleBar(buf, infoX, 13, "Hull   ", r.Hull, r.MaxHull, render.ColorLightGray)

	// Explored counter
	visited := 0
	for _, sys := range sec.Systems {
		if sys.Visited {
			visited++
		}
	}
	buf.WriteString(infoX, 15, fmt.Sprintf("Explored: %d/%d", visited, len(sec.Systems)), render.ColorDarkGray, render.ColorBlack)

	// Comms log
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}

	// Instructions
	buf.WriteString(2, gridRows-1, "WASD: Select star  E: Jump  ESC: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawSystemMapView() {
	buf := g.buffer
	buf.Clear()

	sm := g.sim.Sector.CurrentSystemMap()
	curStar := g.sim.Sector.Systems[g.sim.Sector.CurrentSystem]

	// Viewport dimensions (leave room for title bar, info panel, comms log)
	const (
		vpX = 0
		vpY = 2
		vpW = 58
		vpH = 30
	)

	// Float camera centered exactly on shuttle — map scrolls, shuttle stays at viewport center.
	camFX := sm.Shuttle.X - float64(vpW)/2.0
	camFY := sm.Shuttle.Y - float64(vpH)/2.0
	maxCamX := float64(sm.Width - vpW)
	maxCamY := float64(sm.Height - vpH)
	if camFX < 0 {
		camFX = 0
	}
	if camFY < 0 {
		camFY = 0
	}
	if camFX > maxCamX {
		camFX = maxCamX
	}
	if camFY > maxCamY {
		camFY = maxCamY
	}

	// Title bar
	sysTitle := fmt.Sprintf("--- %s System ---", curStar.Name)
	buf.WriteString(2, 0, sysTitle, render.ColorLightCyan, render.ColorBlack)
	typeLabel := game.StarTypeName(curStar.Type)
	buf.WriteString(2+len(sysTitle)+2, 0, typeLabel, starColor(curStar.Type), render.ColorBlack)

	// Viewport pixel bounds (for clipping floating sprites to the viewport area)
	vpPxLeft := float64(vpX * cellWidth)
	vpPxTop := float64(vpY * cellHeight)
	vpPxRight := float64((vpX + vpW) * cellWidth)
	vpPxBottom := float64((vpY + vpH) * cellHeight)
	cw := float64(cellWidth)
	ch := float64(cellHeight)

	// addSprite converts a world-space position to screen pixels and clips to viewport.
	addSprite := func(glyph byte, fg uint8, worldX, worldY float64) {
		px := (float64(vpX) + worldX - camFX) * cw
		py := (float64(vpY) + worldY - camFY) * ch
		if px >= vpPxLeft-cw && px < vpPxRight && py >= vpPxTop-ch && py < vpPxBottom {
			g.sprites = append(g.sprites, floatingSprite{Glyph: glyph, FG: fg, PX: px, PY: py})
		}
	}

	// Background scatter stars (deterministic hash — all drawn as floating sprites)
	startWX := int(camFX)
	startWY := int(camFY)
	endWX := startWX + vpW + 2
	endWY := startWY + vpH + 2
	if startWX < 0 {
		startWX = 0
	}
	if startWY < 0 {
		startWY = 0
	}
	if endWX > sm.Width {
		endWX = sm.Width
	}
	if endWY > sm.Height {
		endWY = sm.Height
	}
	for wy := startWY; wy < endWY; wy++ {
		for wx := startWX; wx < endWX; wx++ {
			hash := wx*31 + wy*17 + 7
			if hash < 0 {
				hash = -hash
			}
			if hash%23 == 0 {
				addSprite(250, render.ColorDarkGray, float64(wx), float64(wy))
			}
		}
	}

	// Draw space objects as floating sprites
	for i := range sm.Objects {
		obj := &sm.Objects[i]
		glyph, fg := spaceObjectAppearance(obj)
		if obj.Kind == game.ObjStar {
			fg = starColor(curStar.Type)
		}
		addSprite(glyph, fg, float64(obj.X), float64(obj.Y))
	}

	// Shuttle — drawn at its exact float position (viewport center unless camera clamped at edge)
	addSprite(shuttleGlyph(sm.Shuttle.FaceDX, sm.Shuttle.FaceDY), render.ColorWhite, sm.Shuttle.X, sm.Shuttle.Y)

	// --- Info panel (right side) ---
	infoX := panelX
	buf.WriteString(infoX, 2, "--- System ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(infoX, 3, curStar.Name, render.ColorWhite, render.ColorBlack)
	buf.WriteString(infoX, 4, game.StarTypeName(curStar.Type), starColor(curStar.Type), render.ColorBlack)

	// Object counts
	nPlanets, nStations, nShips, nDerelicts := 0, 0, 0, 0
	for _, obj := range sm.Objects {
		switch obj.Kind {
		case game.ObjPlanet:
			nPlanets++
		case game.ObjStation:
			nStations++
		case game.ObjShip:
			nShips++
		case game.ObjDerelict:
			nDerelicts++
		}
	}
	row := 6
	buf.WriteString(infoX, row, fmt.Sprintf("Planets:   %d", nPlanets), render.ColorLightGray, render.ColorBlack)
	row++
	if nStations > 0 {
		buf.WriteString(infoX, row, fmt.Sprintf("Stations:  %d", nStations), render.ColorCyan, render.ColorBlack)
		row++
	}
	if nShips > 0 {
		buf.WriteString(infoX, row, fmt.Sprintf("Ships:     %d", nShips), render.ColorLightGray, render.ColorBlack)
		row++
	}
	if nDerelicts > 0 {
		buf.WriteString(infoX, row, fmt.Sprintf("Derelicts: %d", nDerelicts), render.ColorDarkGray, render.ColorBlack)
		row++
	}
	row++

	// Nearby object info
	nearObj := sm.NearestObject(sm.Shuttle.TileX(), sm.Shuttle.TileY(), 5)
	if nearObj != nil {
		buf.WriteString(infoX, row, "--- Nearby ---", render.ColorLightCyan, render.ColorBlack)
		row++
		buf.WriteString(infoX, row, nearObj.Name, render.ColorWhite, render.ColorBlack)
		row++
		switch nearObj.Kind {
		case game.ObjStar:
			buf.WriteString(infoX, row, "Star - don't fly into it", render.ColorYellow, render.ColorBlack)
		case game.ObjPlanet:
			buf.WriteString(infoX, row, game.PlanetKindName(nearObj.PlanetType), render.ColorLightGray, render.ColorBlack)
			row++
			objIdx := g.findObjectIndex(sm, nearObj)
			if objIdx >= 0 {
				key := game.ScanKey(g.sim.Sector.CurrentSystem, objIdx)
				if scan, ok := g.sim.Discovery.PlanetsScanned[key]; ok {
					buf.WriteString(infoX, row, "SCANNED", render.ColorLightGreen, render.ColorBlack)
					row++
					buf.WriteString(infoX, row, scan.Resources, render.ColorLightGray, render.ColorBlack)
				} else {
					buf.WriteString(infoX, row, "Unscanned - press E", render.ColorYellow, render.ColorBlack)
				}
			}
		case game.ObjStation:
			buf.WriteString(infoX, row, "Station - press E to dock", render.ColorCyan, render.ColorBlack)
		case game.ObjDerelict:
			buf.WriteString(infoX, row, "Derelict - salvageable?", render.ColorDarkGray, render.ColorBlack)
		case game.ObjShip:
			kind := game.ShipAIKindName(nearObj.AIKind)
			clr := uint8(render.ColorLightGray)
			if nearObj.AIKind == game.AIPirate {
				clr = render.ColorLightRed
			}
			buf.WriteString(infoX, row, kind+" vessel", clr, render.ColorBlack)
		}
		row++
		// Distance label
		dx := nearObj.X - sm.Shuttle.TileX()
		dy := nearObj.Y - sm.Shuttle.TileY()
		d2 := dx*dx + dy*dy
		distLabel := "Approaching"
		if d2 <= 4 {
			distLabel = "Very close"
		} else if d2 <= 9 {
			distLabel = "Close"
		}
		buf.WriteString(infoX, row, distLabel, render.ColorDarkGray, render.ColorBlack)
		row++
	}
	row++

	// Ship status
	r := &g.sim.Resources
	buf.WriteString(infoX, row, "--- Ship ---", render.ColorLightCyan, render.ColorBlack)
	row++
	drawEnergyBar(buf, infoX, row, "Energy ", r.Energy, r.MaxEnergy, render.ColorYellow)
	row++
	drawSimpleBar(buf, infoX, row, "Hull   ", r.Hull, r.MaxHull, render.ColorLightGray)
	row += 2
	buf.WriteString(infoX, row, fmt.Sprintf("Pos: %d, %d", sm.Shuttle.TileX(), sm.Shuttle.TileY()), render.ColorDarkGray, render.ColorBlack)
	row++
	spdPct := sm.Shuttle.SpeedPct()
	spdClr := uint8(render.ColorDarkGray)
	if spdPct > 75 {
		spdClr = render.ColorLightGreen
	} else if spdPct > 25 {
		spdClr = render.ColorLightGray
	}
	buf.WriteString(infoX, row, fmt.Sprintf("Speed: %d%%", spdPct), spdClr, render.ColorBlack)

	// --- Radar minimap ---
	g.drawRadar(buf, sm, curStar)

	// Comms log
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}

	// Instructions
	buf.WriteString(2, gridRows-1, "WASD: Fly  E: Interact/Scan  N: Nav  Tab: Status  ESC: Ship", render.ColorDarkGray, render.ColorBlack)
}

// drawRadar renders a shuttle-centered minimap of the star system.
// The shuttle is always at the center; objects scroll around it.
func (g *Game) drawRadar(buf *render.CellBuffer, sm *game.SystemMap, star game.StarSystem) {
	const (
		radarX     = panelX     // left edge of radar
		radarY     = 24         // header row
		radarW     = 18         // minimap width in cells
		radarH     = 8          // minimap height in cells
		bodyY      = radarY + 1 // first row of minimap body
		radarScale = 18.0       // world tiles per radar cell
	)

	centerRX := radarW / 2 // shuttle always at center cell
	centerRY := radarH / 2

	// Header
	buf.WriteString(radarX, radarY, "--- Radar ---", render.ColorLightCyan, render.ColorBlack)

	// Background — dark speckled field
	for y := 0; y < radarH; y++ {
		for x := 0; x < radarW; x++ {
			buf.Set(radarX+1+x, bodyY+y, 176, render.ColorDarkGray, render.ColorBlack) // ░
		}
	}

	// Map objects relative to shuttle position
	for i := range sm.Objects {
		obj := &sm.Objects[i]
		dx := float64(obj.X) - sm.Shuttle.X
		dy := float64(obj.Y) - sm.Shuttle.Y
		rx := centerRX + int(dx/radarScale+0.5)
		ry := centerRY + int(dy/radarScale+0.5)

		// Skip if off radar
		if rx < 0 || rx >= radarW || ry < 0 || ry >= radarH {
			continue
		}

		var glyph byte
		var fg uint8
		switch obj.Kind {
		case game.ObjStar:
			glyph = '*'
			fg = starColor(star.Type)
		case game.ObjPlanet:
			glyph = 'o'
			fg = planetColor(obj.PlanetType)
		case game.ObjStation:
			glyph = 'H'
			fg = render.ColorCyan
		case game.ObjDerelict:
			glyph = '%'
			fg = render.ColorDarkGray
		case game.ObjShip:
			glyph = '.'
			fg = shipColor(obj.AIKind)
		default:
			continue
		}
		buf.Set(radarX+1+rx, bodyY+ry, glyph, fg, render.ColorBlack)
	}

	// Shuttle marker — always at center
	buf.Set(radarX+1+centerRX, bodyY+centerRY, '+', render.ColorWhite, render.ColorBlack)
}

// drawJumpLine draws a dotted line between two points on the sector map.
func drawJumpLine(buf *render.CellBuffer, x1, y1, x2, y2 int) {
	dx := x2 - x1
	dy := y2 - y1
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		return
	}
	for i := 1; i < steps; i++ {
		x := x1 + dx*i/steps
		y := y1 + dy*i/steps
		if i%2 == 0 {
			buf.Set(x, y, 250, render.ColorDarkGray, render.ColorBlack) // · middle dot
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func starColor(t game.StarType) uint8 {
	switch t {
	case game.StarYellow:
		return render.ColorYellow
	case game.StarRed:
		return render.ColorLightRed
	case game.StarBlue:
		return render.ColorLightBlue
	case game.StarWhite:
		return render.ColorWhite
	case game.StarOrange:
		return render.ColorBrown
	default:
		return render.ColorWhite
	}
}

func spaceObjectAppearance(obj *game.SpaceObject) (glyph byte, fg uint8) {
	switch obj.Kind {
	case game.ObjStar:
		return '*', render.ColorYellow
	case game.ObjPlanet:
		return 'O', planetColor(obj.PlanetType)
	case game.ObjStation:
		return 'H', render.ColorCyan
	case game.ObjDerelict:
		return '%', render.ColorDarkGray
	case game.ObjAsteroid:
		return '.', render.ColorLightGray
	case game.ObjShip:
		return shipGlyph(obj.AIKind), shipColor(obj.AIKind)
	default:
		return '?', render.ColorWhite
	}
}

func planetColor(k game.PlanetKind) uint8 {
	switch k {
	case game.PlanetRocky:
		return render.ColorBrown
	case game.PlanetGas:
		return render.ColorLightMagenta
	case game.PlanetIce:
		return render.ColorLightCyan
	case game.PlanetVolcanic:
		return render.ColorLightRed
	default:
		return render.ColorLightGray
	}
}

func shipGlyph(k game.ShipAIKind) byte {
	switch k {
	case game.AITrader:
		return 'T'
	case game.AIPatrol:
		return 'P'
	case game.AIPirate:
		return '!'
	default:
		return '?'
	}
}

func shuttleGlyph(dx, dy int) byte {
	if dy < 0 {
		return '^'
	}
	if dy > 0 {
		return 'v'
	}
	if dx < 0 {
		return '<'
	}
	if dx > 0 {
		return '>'
	}
	return '^' // default facing up
}

func shipColor(k game.ShipAIKind) uint8 {
	switch k {
	case game.AITrader:
		return render.ColorLightGreen
	case game.AIPatrol:
		return render.ColorLightBlue
	case game.AIPirate:
		return render.ColorLightRed
	default:
		return render.ColorWhite
	}
}

// --- UI helper functions ---

func msgColor(p game.MsgPriority) uint8 {
	switch p {
	case game.MsgCritical:
		return render.ColorLightRed
	case game.MsgWarning:
		return render.ColorYellow
	case game.MsgDiscovery:
		return render.ColorLightGreen
	case game.MsgSocial:
		return render.ColorWhite
	default:
		return render.ColorCyan
	}
}

// drawMatterBar shows a bar split into clean (solid) and dirty (shaded) segments.
func drawMatterBar(buf *render.CellBuffer, x, y int, label string, pool *game.MatterPool, cleanClr, dirtyClr uint8) {
	barW := 20
	cap := pool.Capacity
	if cap == 0 {
		cap = 1
	}
	cleanFill := barW * pool.Clean / cap
	dirtyFill := barW * pool.Dirty / cap

	// Label color based on clean level
	labelClr := uint8(render.ColorLightGray)
	cleanPct := pool.Clean * 100 / cap
	if cleanPct <= 15 {
		labelClr = render.ColorLightRed
	} else if cleanPct <= 30 {
		labelClr = render.ColorYellow
	}
	buf.WriteString(x, y, label, labelClr, render.ColorBlack)

	for i := 0; i < barW; i++ {
		if i < cleanFill {
			buf.Set(x+8+i, y, 219, cleanClr, render.ColorBlack) // █ clean
		} else if i < cleanFill+dirtyFill {
			buf.Set(x+8+i, y, 178, dirtyClr, render.ColorBlack) // ▓ dirty
		} else {
			buf.Set(x+8+i, y, 176, render.ColorBlack, render.ColorBlack) // ░ lost/held
		}
	}
	info := fmt.Sprintf("%2dc/%2dd", pool.Clean, pool.Dirty)
	buf.WriteString(x+29, y, info, labelClr, render.ColorBlack)
}

// drawEnergyBar shows a single-value bar (no clean/dirty split).
func drawEnergyBar(buf *render.CellBuffer, x, y int, label string, val, max int, clr uint8) {
	barW := 20
	if max == 0 {
		max = 1
	}
	filled := barW * val / max

	labelClr := uint8(render.ColorLightGray)
	pct := val * 100 / max
	if pct <= 15 {
		labelClr = render.ColorLightRed
	} else if pct <= 30 {
		labelClr = render.ColorYellow
	}
	buf.WriteString(x, y, label, labelClr, render.ColorBlack)

	for i := 0; i < barW; i++ {
		if i < filled {
			buf.Set(x+8+i, y, 219, clr, render.ColorBlack) // █
		} else {
			buf.Set(x+8+i, y, 176, render.ColorDarkGray, render.ColorBlack) // ░
		}
	}
	info := fmt.Sprintf("%3d/%d", val, max)
	buf.WriteString(x+29, y, info, labelClr, render.ColorBlack)
}

// drawSimpleBar is a basic percentage bar (for hull, etc.)
func drawSimpleBar(buf *render.CellBuffer, x, y int, label string, val, max int, clr uint8) {
	drawEnergyBar(buf, x, y, label, val, max, clr)
}

// drawEquipStatus renders an ON/OFF indicator for toggleable equipment.
func drawEquipStatus(buf *render.CellBuffer, x, y int, name string, on bool) {
	if on {
		buf.WriteString(x, y, fmt.Sprintf(" %s", name), render.ColorLightGreen, render.ColorBlack)
		buf.WriteString(x+len(name)+1, y, " ON", render.ColorLightGreen, render.ColorBlack)
	} else {
		buf.WriteString(x, y, fmt.Sprintf(" %s", name), render.ColorDarkGray, render.ColorBlack)
		buf.WriteString(x+len(name)+1, y, " OFF", render.ColorLightRed, render.ColorBlack)
	}
}

// drawNeedBar shows a player need (hunger/thirst/hygiene) as a 10-char bar with label.
func drawNeedBar(buf *render.CellBuffer, x, y int, label string, val int) {
	barW := 10
	filled := barW * val / 100

	// Color based on severity
	clr := uint8(render.ColorLightGreen)
	switch {
	case val >= 80:
		clr = render.ColorLightRed
	case val >= 60:
		clr = render.ColorYellow
	case val >= 30:
		clr = render.ColorLightGray
	}

	buf.WriteString(x, y, label, clr, render.ColorBlack)
	for i := 0; i < barW; i++ {
		if i < filled {
			buf.Set(x+8+i, y, 219, clr, render.ColorBlack) // █
		} else {
			buf.Set(x+8+i, y, 176, render.ColorDarkGray, render.ColorBlack) // ░
		}
	}
	lvl := game.NeedLevel(val)
	buf.WriteString(x+19, y, lvl, clr, render.ColorBlack)
}

// --- Update dispatch ---

func (g *Game) Update() error {
	// Always tick simulation — resources drain even while looking at the map
	g.sim.Tick()

	switch g.viewMode {
	case ViewSectorMap:
		return g.updateSectorMap()
	case ViewSystemMap:
		return g.updateSystemMap()
	case ViewStation:
		return g.updateStation()
	case ViewCargo:
		return g.updateCargo()
	case ViewCharSheet:
		return g.updateCharSheet()
	case ViewEncounter:
		return g.updateEncounter()
	case ViewEpisode:
		return g.updateEpisode()
	case ViewSurface:
		return g.updateSurface()
	default:
		return g.updateShip()
	}
}

func (g *Game) updateShip() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.sim.IsOrbiting() {
			// ESC while orbiting → leave orbit, return to system map
			g.sim.LeaveOrbit()
			g.viewMode = ViewSystemMap
			return nil
		}
		return ebiten.Termination
	}

	// Player movement
	dx, dy := 0, 0
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		dy = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		dy = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		dx = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		dx = 1
	}
	if dx != 0 || dy != 0 {
		g.sim.TryMovePlayer(dx, dy)
	}

	// Interact / Toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.sim.Interact()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.sim.ToggleEquipment()
	}

	// Nav console → sector map (pick a star to jump to)
	if g.sim.NavActivated {
		g.sim.NavActivated = false
		g.sim.Sector.CursorSystem = g.sim.Sector.CurrentSystem
		g.viewMode = ViewSectorMap
	}

	// Pilot console → surface (if landed), leave orbit (if orbiting), or system map
	if g.sim.IsOnSurface() {
		// Just landed from pilot console — go to surface view
		g.viewMode = ViewSurface
	} else if g.sim.PilotActivated {
		g.sim.PilotActivated = false
		if g.sim.IsOrbiting() {
			g.sim.LeaveOrbit()
		}
		g.sim.Sector.EnsureSystemMap(g.sim.Sector.CurrentSystem)
		g.viewMode = ViewSystemMap
	}

	// Cargo console → cargo view (inspect and jettison)
	if g.sim.CargoActivated {
		g.sim.CargoActivated = false
		g.viewMode = ViewCargo
	}

	// Science console → scan orbited planet, or character sheet if not orbiting
	if g.sim.ScanActivated {
		g.sim.ScanActivated = false
		if g.sim.IsOrbiting() {
			g.sim.ScanPlanet(g.sim.OrbitPlanetIdx)
		} else {
			g.prevViewMode = ViewShip
			g.viewMode = ViewCharSheet
		}
	}

	// Comms station (viewscreen) → encounter
	if g.sim.CommsActivated {
		g.sim.CommsActivated = false
		g.sim.StartEncounter()
		if g.sim.ActiveEncounter != nil {
			g.prevViewMode = ViewShip
			g.viewMode = ViewEncounter
		}
	}

	// Tab → character sheet
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.prevViewMode = ViewShip
		g.viewMode = ViewCharSheet
	}

	// Redraw
	g.drawScreen()

	// Mouse hover: convert screen pixel to grid cell to ship tile
	mx, my := ebiten.CursorPosition()
	cellX := mx / cellWidth
	cellY := my / cellHeight
	g.updateHoverInfo(cellX, cellY)

	// Update FPS counter
	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

func (g *Game) updateSectorMap() error {
	// ESC → back to system map (not ship interior)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.viewMode = ViewSystemMap
		g.drawScreen()
		return nil
	}

	// Cursor movement — snap to nearest star in direction
	dx, dy := 0, 0
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		dy = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		dy = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		dx = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		dx = 1
	}
	if dx != 0 || dy != 0 {
		next := g.sim.Sector.NearestInDirection(dx, dy)
		if next >= 0 {
			g.sim.Sector.CursorSystem = next
		}
	}

	// Jump to selected system
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		target := g.sim.Sector.CursorSystem
		if target != g.sim.Sector.CurrentSystem {
			if g.sim.NavigateTo(target) {
				if g.sim.ActiveEpisode != nil {
					g.viewMode = ViewEpisode
				} else {
					g.viewMode = ViewSystemMap
				}
			}
		}
	}

	// Redraw
	g.drawScreen()

	// Update FPS counter
	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

func (g *Game) updateSystemMap() error {
	// ESC → back to ship interior
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.viewMode = ViewShip
		return nil
	}

	// N → open sector nav map for interstellar jumps
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.sim.Sector.CursorSystem = g.sim.Sector.CurrentSystem
		g.viewMode = ViewSectorMap
		return nil
	}

	// Tab → character sheet
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.prevViewMode = ViewSystemMap
		g.viewMode = ViewCharSheet
		return nil
	}

	// Thrust-based WASD flight (physics handles acceleration, drag, speed cap)
	{
		dx, dy := 0, 0
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			dy = -1
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			dy = 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
			dx = -1
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
			dx = 1
		}
		if dx != 0 || dy != 0 {
			sm := g.sim.Sector.CurrentSystemMap()
			sm.Shuttle.ApplyThrust(dx, dy)
			g.sim.Skills.AddXP(game.SkillPiloting, 0.1)
		}
	}

	// E near object → approach interaction, dock, or scan
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		sm := g.sim.Sector.CurrentSystemMap()
		obj := sm.NearestObject(sm.Shuttle.TileX(), sm.Shuttle.TileY(), 3)
		if obj != nil {
			switch obj.Kind {
			case game.ObjStation:
				// Dock at station
				sd := g.sim.DockAtStation()
				if sd != nil {
					g.stationData = sd
					g.stationMenu = stMenuMain
					g.viewMode = ViewStation
				}
			case game.ObjPlanet:
				// Enter orbit around planet → transitions to ship interior
				objIdx := g.findObjectIndex(sm, obj)
				if objIdx >= 0 {
					g.sim.EnterOrbit(objIdx)
					g.viewMode = ViewShip
				}
			default:
				g.logApproachInfo(obj)
			}
		} else {
			g.sim.Log.Add("Nothing nearby to interact with.", game.MsgSocial)
		}
	}

	// Redraw
	g.drawScreen()

	// Update FPS counter
	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

// logApproachInfo logs a message when the player interacts near a space object.
func (g *Game) logApproachInfo(obj *game.SpaceObject) {
	switch obj.Kind {
	case game.ObjStar:
		g.sim.Log.Add(fmt.Sprintf("Approaching %s. Careful of the corona.", obj.Name), game.MsgWarning)
	case game.ObjPlanet:
		g.sim.Log.Add(fmt.Sprintf("Orbiting %s. %s.", obj.Name, game.PlanetKindName(obj.PlanetType)), game.MsgDiscovery)
	case game.ObjStation:
		g.sim.Log.Add(fmt.Sprintf("Hailing %s. Fly closer and press E to dock.", obj.Name), game.MsgInfo)
	case game.ObjDerelict:
		g.sim.Log.Add("Derelict detected on sensors. Salvage potential (future).", game.MsgDiscovery)
	case game.ObjShip:
		switch obj.AIKind {
		case game.AITrader:
			g.sim.Log.Add(fmt.Sprintf("A %s drifts nearby. Hailing frequencies open.", obj.Name), game.MsgInfo)
		case game.AIPatrol:
			g.sim.Log.Add(fmt.Sprintf("%s scanning the area. Papers in order.", obj.Name), game.MsgInfo)
		case game.AIPirate:
			g.sim.Log.Add(fmt.Sprintf("WARNING: %s on intercept course!", obj.Name), game.MsgWarning)
		}
	}
}

// --- Station view ---

func (g *Game) drawStationView() {
	buf := g.buffer
	buf.Clear()

	switch g.stationMenu {
	case stMenuRepairs:
		g.drawStationRepairs(buf)
	case stMenuTrade:
		g.drawStationTrade(buf)
	case stMenuBuy:
		g.drawStationBuy(buf)
	case stMenuSell:
		g.drawStationSell(buf)
	case stMenuBar:
		g.drawStationBar(buf)
	case stMenuFaction:
		g.drawStationFaction(buf)
	default:
		g.drawStationMain(buf)
	}

	// Comms log (always visible)
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}
}

func (g *Game) drawStationMain(buf *render.CellBuffer) {
	sd := g.stationData
	cx := 4 // content left margin

	// Banner
	buf.WriteString(cx, 2, "=========================================", render.ColorCyan, render.ColorBlack)
	buf.WriteString(cx+1, 3, fmt.Sprintf("WELCOME TO %s", sd.Name), render.ColorWhite, render.ColorBlack)
	buf.WriteString(cx+1, 4, fmt.Sprintf("\"%s\"", sd.Tagline), render.ColorDarkGray, render.ColorBlack)
	buf.WriteString(cx, 5, "=========================================", render.ColorCyan, render.ColorBlack)

	// Menu
	buf.WriteString(cx+2, 7, "1. Repairs & Maintenance", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx+2, 8, "2. Trade Goods", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx+2, 9, "3. Bar", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx+2, 10, fmt.Sprintf("4. %s Office", sd.Faction), render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx+2, 12, "5. Undock", render.ColorYellow, render.ColorBlack)

	// Footer
	r := &g.sim.Resources
	buf.WriteString(cx+2, 14, fmt.Sprintf("Credits: %d    Cargo: %d/%d pads",
		r.Credits, r.PadsUsed(), len(r.CargoPads)), render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(cx, 15, "=========================================", render.ColorCyan, render.ColorBlack)

	// Ship status (right panel)
	infoX := panelX
	buf.WriteString(infoX, 2, "--- Ship Status ---", render.ColorLightCyan, render.ColorBlack)
	drawEnergyBar(buf, infoX, 3, "Energy ", r.Energy, r.MaxEnergy, render.ColorYellow)
	drawSimpleBar(buf, infoX, 4, "Hull   ", r.Hull, r.MaxHull, render.ColorLightGray)
	drawMatterBar(buf, infoX, 6, "Water  ", &r.Water, render.ColorBlue, render.ColorDarkGray)
	drawMatterBar(buf, infoX, 7, "Organic", &r.Organic, render.ColorGreen, render.ColorDarkGray)

	buf.WriteString(2, gridRows-1, "1-5: Select  ESC: Undock", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawStationRepairs(buf *render.CellBuffer) {
	cx := 4
	r := &g.sim.Resources
	damage := r.MaxHull - r.Hull

	buf.WriteString(cx, 2, "--- REPAIRS & MAINTENANCE ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(cx, 4, fmt.Sprintf("Hull: %d / %d", r.Hull, r.MaxHull), render.ColorWhite, render.ColorBlack)
	drawSimpleBar(buf, cx, 5, "       ", r.Hull, r.MaxHull, render.ColorLightGray)

	if damage == 0 {
		buf.WriteString(cx, 7, "Hull integrity at 100%. No repairs needed.", render.ColorLightGreen, render.ColorBlack)
	} else {
		fullCost := damage * 2
		buf.WriteString(cx, 7, fmt.Sprintf("Damage: %d pts   Full repair: %dcr", damage, fullCost), render.ColorYellow, render.ColorBlack)
		buf.WriteString(cx, 9, fmt.Sprintf("1. Full repair (%dcr)", fullCost), render.ColorLightGray, render.ColorBlack)
		tenCost := 10 * 2
		if damage < 10 {
			tenCost = damage * 2
		}
		buf.WriteString(cx, 10, fmt.Sprintf("2. Repair 10 pts (%dcr)", tenCost), render.ColorLightGray, render.ColorBlack)
	}

	buf.WriteString(cx, 12, fmt.Sprintf("Credits: %d", r.Credits), render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(cx, 14, "0. Back", render.ColorYellow, render.ColorBlack)
	buf.WriteString(2, gridRows-1, "1-2: Repair  0: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawStationTrade(buf *render.CellBuffer) {
	cx := 4

	buf.WriteString(cx, 2, "--- TRADE GOODS ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(cx, 4, "1. Buy from station", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx, 5, "2. Sell to station", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx, 7, "0. Back", render.ColorYellow, render.ColorBlack)

	r := &g.sim.Resources
	buf.WriteString(cx, 9, fmt.Sprintf("Credits: %d    Cargo: %d/%d pads",
		r.Credits, r.PadsUsed(), len(r.CargoPads)), render.ColorLightCyan, render.ColorBlack)

	buf.WriteString(2, gridRows-1, "1-2: Select  0: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawStationBuy(buf *render.CellBuffer) {
	sd := g.stationData
	cx := 4
	r := &g.sim.Resources

	buf.WriteString(cx, 2, "--- BUY FROM STATION ---", render.ColorLightCyan, render.ColorBlack)

	stocked := sd.StockedList()
	row := 4
	for i, k := range stocked {
		price := sd.SellPrices[k]
		stock := sd.Stock[k]
		clr := uint8(render.ColorLightGray)
		if stock == 0 {
			clr = render.ColorDarkGray
		} else if price > r.Credits {
			clr = render.ColorDarkGray
		}
		label := fmt.Sprintf("%d. %-18s %3dcr  (x%d)", i+1, game.CargoName(k), price, stock)
		buf.WriteString(cx, row, label, clr, render.ColorBlack)
		row++
	}

	row += 1
	buf.WriteString(cx, row, fmt.Sprintf("Credits: %d    Cargo: %d/%d pads",
		r.Credits, r.PadsUsed(), len(r.CargoPads)), render.ColorLightCyan, render.ColorBlack)
	row += 2
	buf.WriteString(cx, row, "0. Back", render.ColorYellow, render.ColorBlack)

	buf.WriteString(2, gridRows-1, "1-8: Buy item  0: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawStationSell(buf *render.CellBuffer) {
	sd := g.stationData
	cx := 4
	r := &g.sim.Resources

	buf.WriteString(cx, 2, "--- SELL TO STATION ---", render.ColorLightCyan, render.ColorBlack)

	row := 4
	anyItems := false
	for i, pad := range r.CargoPads {
		if pad.Kind == game.CargoNone {
			continue
		}
		anyItems = true
		price := sd.BuyPrices[pad.Kind]
		label := fmt.Sprintf("%d. %-18s %3dcr  (x%d)", i+1, game.CargoName(pad.Kind), price, pad.Count)
		buf.WriteString(cx, row, label, render.ColorLightGray, render.ColorBlack)
		row++
	}

	if !anyItems {
		buf.WriteString(cx, row, "Cargo bay empty. Nothing to sell.", render.ColorDarkGray, render.ColorBlack)
		row++
	}

	row += 1
	buf.WriteString(cx, row, fmt.Sprintf("Credits: %d    Cargo: %d/%d pads",
		r.Credits, r.PadsUsed(), len(r.CargoPads)), render.ColorLightCyan, render.ColorBlack)
	row += 2
	buf.WriteString(cx, row, "0. Back", render.ColorYellow, render.ColorBlack)

	buf.WriteString(2, gridRows-1, "1-9: Sell from pad  0: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawStationBar(buf *render.CellBuffer) {
	sd := g.stationData
	cx := 4

	buf.WriteString(cx, 2, "--- BAR ---", render.ColorLightCyan, render.ColorBlack)

	// Render bar scene text (may contain \n for multi-line)
	row := 4
	line := ""
	for _, ch := range sd.BarScene {
		if ch == '\n' {
			buf.WriteString(cx, row, line, render.ColorLightGray, render.ColorBlack)
			row++
			line = ""
		} else {
			line += string(ch)
		}
	}
	if line != "" {
		buf.WriteString(cx, row, line, render.ColorLightGray, render.ColorBlack)
		row++
	}

	row += 2
	buf.WriteString(cx, row, "0. Back", render.ColorYellow, render.ColorBlack)
	buf.WriteString(2, gridRows-1, "0: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawStationFaction(buf *render.CellBuffer) {
	sd := g.stationData
	cx := 4

	buf.WriteString(cx, 2, fmt.Sprintf("--- %s OFFICE ---", sd.Faction), render.ColorLightCyan, render.ColorBlack)

	buf.WriteString(cx, 4, "A recruiter sits behind a battered desk covered in", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx, 5, "pamphlets. A poster reads:", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(cx, 7, fmt.Sprintf("\"JOIN THE %s.", sd.Faction), render.ColorWhite, render.ColorBlack)
	buf.WriteString(cx, 8, " See the galaxy. Die heroically.", render.ColorWhite, render.ColorBlack)
	buf.WriteString(cx, 9, " Pension not included.\"", render.ColorWhite, render.ColorBlack)

	buf.WriteString(cx, 11, "The USS Monkey Lion emblem hangs on the wall.", render.ColorDarkGray, render.ColorBlack)
	buf.WriteString(cx, 12, "You get the feeling this is where legends begin.", render.ColorDarkGray, render.ColorBlack)
	buf.WriteString(cx, 13, "Or at least where the paperwork does.", render.ColorDarkGray, render.ColorBlack)

	buf.WriteString(cx, 15, "\"Nothing available right now, but check back.\"", render.ColorLightGray, render.ColorBlack)

	buf.WriteString(cx, 17, "0. Back", render.ColorYellow, render.ColorBlack)
	buf.WriteString(2, gridRows-1, "0: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) updateStation() error {
	// ESC always undocks
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.sim.Log.Add("Undocked.", game.MsgInfo)
		g.stationData = nil
		g.viewMode = ViewSystemMap
		g.drawScreen()
		return nil
	}

	switch g.stationMenu {
	case stMenuMain:
		g.updateStationMain()
	case stMenuRepairs:
		g.updateStationRepairs()
	case stMenuTrade:
		g.updateStationTrade()
	case stMenuBuy:
		g.updateStationBuy()
	case stMenuSell:
		g.updateStationSell()
	case stMenuBar, stMenuFaction:
		if pressedDigit(0) {
			g.stationMenu = stMenuMain
		}
	}

	g.drawScreen()

	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

func (g *Game) updateStationMain() {
	if pressedDigit(1) {
		g.stationMenu = stMenuRepairs
	} else if pressedDigit(2) {
		g.stationMenu = stMenuTrade
	} else if pressedDigit(3) {
		g.stationMenu = stMenuBar
		g.sim.Skills.AddXP(game.SkillDiplomacy, 1.0)
	} else if pressedDigit(4) {
		g.stationMenu = stMenuFaction
	} else if pressedDigit(5) {
		g.sim.Log.Add("Undocked.", game.MsgInfo)
		g.stationData = nil
		g.viewMode = ViewSystemMap
	}
}

func (g *Game) updateStationRepairs() {
	if pressedDigit(1) {
		g.sim.RepairHull(0) // 0 = full repair
	} else if pressedDigit(2) {
		g.sim.RepairHull(10)
	} else if pressedDigit(0) {
		g.stationMenu = stMenuMain
	}
}

func (g *Game) updateStationTrade() {
	if pressedDigit(1) {
		g.stationMenu = stMenuBuy
	} else if pressedDigit(2) {
		g.stationMenu = stMenuSell
	} else if pressedDigit(0) {
		g.stationMenu = stMenuMain
	}
}

func (g *Game) updateStationBuy() {
	sd := g.stationData
	stocked := sd.StockedList()
	for i, k := range stocked {
		if pressedDigit(i + 1) {
			g.sim.BuyCargo(sd, k)
			break
		}
	}
	if pressedDigit(0) {
		g.stationMenu = stMenuTrade
	}
}

func (g *Game) updateStationSell() {
	sd := g.stationData
	r := &g.sim.Resources
	for i := range r.CargoPads {
		if pressedDigit(i + 1) {
			g.sim.SellCargo(sd, i)
			break
		}
	}
	if pressedDigit(0) {
		g.stationMenu = stMenuTrade
	}
}

// pressedDigit returns true if the number key (0-9) was just pressed.
func pressedDigit(n int) bool {
	switch n {
	case 0:
		return inpututil.IsKeyJustPressed(ebiten.Key0)
	case 1:
		return inpututil.IsKeyJustPressed(ebiten.Key1)
	case 2:
		return inpututil.IsKeyJustPressed(ebiten.Key2)
	case 3:
		return inpututil.IsKeyJustPressed(ebiten.Key3)
	case 4:
		return inpututil.IsKeyJustPressed(ebiten.Key4)
	case 5:
		return inpututil.IsKeyJustPressed(ebiten.Key5)
	case 6:
		return inpututil.IsKeyJustPressed(ebiten.Key6)
	case 7:
		return inpututil.IsKeyJustPressed(ebiten.Key7)
	case 8:
		return inpututil.IsKeyJustPressed(ebiten.Key8)
	case 9:
		return inpututil.IsKeyJustPressed(ebiten.Key9)
	}
	return false
}

// --- Cargo view ---

func (g *Game) drawCargoView() {
	buf := g.buffer
	buf.Clear()

	cx := 4
	r := &g.sim.Resources

	buf.WriteString(cx, 2, "--- CARGO MANIFEST ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(cx, 3, fmt.Sprintf("Pads: %d/%d used    Total units: %d",
		r.PadsUsed(), len(r.CargoPads), r.CargoCount()), render.ColorLightGray, render.ColorBlack)

	row := 5
	anyItems := false
	for i, pad := range r.CargoPads {
		if pad.Kind == game.CargoNone {
			continue
		}
		anyItems = true
		label := fmt.Sprintf("%d. %-18s x%d", i+1, game.CargoName(pad.Kind), pad.Count)
		buf.WriteString(cx, row, label, render.ColorLightGray, render.ColorBlack)
		row++
	}

	if !anyItems {
		buf.WriteString(cx, row, "Cargo bay empty.", render.ColorDarkGray, render.ColorBlack)
		row++
	}

	row += 2
	buf.WriteString(cx, row, fmt.Sprintf("Credits: %d", r.Credits), render.ColorLightCyan, render.ColorBlack)
	row++
	// Fuel tank status
	fuelClr := uint8(render.ColorRed)
	if r.JumpFuel > r.MaxJumpFuel/2 {
		fuelClr = render.ColorYellow
	}
	if r.JumpFuel >= r.MaxJumpFuel*9/10 {
		fuelClr = render.ColorLightGreen
	}
	buf.WriteString(cx, row, fmt.Sprintf("Jump Fuel: %d/%d", r.JumpFuel, r.MaxJumpFuel), fuelClr, render.ColorBlack)
	row += 2
	buf.WriteString(cx, row, "1-9: Incinerate (cargo -> fuel)", render.ColorYellow, render.ColorBlack)
	row++
	buf.WriteString(cx, row, "Shift+1-9: Jettison (throw away)", render.ColorDarkGray, render.ColorBlack)

	// Comms log
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}

	buf.WriteString(2, gridRows-1, "1-9: Incinerate  Shift+1-9: Jettison  ESC: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) updateCargo() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.viewMode = ViewShip
		g.drawScreen()
		return nil
	}

	r := &g.sim.Resources
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	for i := range r.CargoPads {
		if pressedDigit(i + 1) {
			if shift {
				// Jettison (throw away)
				g.sim.JettisonCargo(i)
			} else {
				// Incinerate (convert to fuel)
				g.sim.IncinerateCargo(i)
			}
			break
		}
	}

	g.drawScreen()

	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

// --- Encounter view ---

func (g *Game) drawEncounterView() {
	buf := g.buffer
	buf.Clear()

	enc := g.sim.ActiveEncounter
	if enc == nil {
		return
	}

	cx := 4

	// Title
	buf.WriteString(cx, 1, "=== ENCOUNTER ===", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(gridCols-12, 1, "ESC: End", render.ColorDarkGray, render.ColorBlack)

	// Ship info
	buf.WriteString(cx, 2, "=========================================", render.ColorCyan, render.ColorBlack)
	buf.WriteString(cx+1, 3, "INCOMING TRANSMISSION", render.ColorWhite, render.ColorBlack)
	kindLabel := game.EncounterKindLabel(enc.Kind)
	shipLine := fmt.Sprintf("%s \"%s\"", kindLabel, enc.ShipName)
	// Use a shorter display: just the full Name from the SpaceObject
	shipLine = enc.ShipName
	clr := uint8(render.ColorLightGray)
	switch enc.Kind {
	case game.EncounterTrader:
		clr = render.ColorLightGreen
	case game.EncounterPatrol:
		clr = render.ColorLightBlue
	case game.EncounterPirate:
		clr = render.ColorLightRed
	}
	buf.WriteString(cx+1, 4, shipLine, clr, render.ColorBlack)
	buf.WriteString(cx, 5, "-----------------------------------------", render.ColorCyan, render.ColorBlack)

	// Greeting
	buf.WriteString(cx+1, 7, fmt.Sprintf("\"%s\"", enc.Greeting), render.ColorWhite, render.ColorBlack)
	buf.WriteString(cx, 8, "=========================================", render.ColorCyan, render.ColorBlack)

	// Options
	row := 10
	for i, opt := range enc.Options {
		label := fmt.Sprintf(" %d. %s", i+1, opt.Label)
		optClr := uint8(render.ColorLightGray)
		if !opt.Enabled {
			optClr = render.ColorDarkGray
			label += " [" + opt.DisableText + "]"
		}
		if enc.Resolved {
			optClr = render.ColorDarkGray // dim all options after resolution
		}
		buf.WriteString(cx, row, label, optClr, render.ColorBlack)
		row++
	}

	// Result text (after resolution)
	if enc.ResultText != "" {
		row++
		buf.WriteString(cx, row, "-----------------------------------------", render.ColorCyan, render.ColorBlack)
		row++
		// Handle multi-line result text
		line := ""
		for _, ch := range enc.ResultText {
			if ch == '\n' {
				buf.WriteString(cx+1, row, line, render.ColorWhite, render.ColorBlack)
				row++
				line = ""
			} else {
				line += string(ch)
			}
		}
		if line != "" {
			buf.WriteString(cx+1, row, line, render.ColorWhite, render.ColorBlack)
			row++
		}
		buf.WriteString(cx, row, "-----------------------------------------", render.ColorCyan, render.ColorBlack)
	}

	// Comms log
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		msgClr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, msgClr, render.ColorBlack)
	}

	if enc.Resolved {
		buf.WriteString(2, gridRows-1, "ESC: End transmission", render.ColorDarkGray, render.ColorBlack)
	} else {
		buf.WriteString(2, gridRows-1, "1-9: Choose option  ESC: End transmission", render.ColorDarkGray, render.ColorBlack)
	}
}

func (g *Game) updateEncounter() error {
	enc := g.sim.ActiveEncounter

	// ESC ends the encounter
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.sim.EndEncounter()
		g.viewMode = g.prevViewMode
		g.drawScreen()
		return nil
	}

	if enc != nil && !enc.Resolved {
		// Check digit key presses for option selection
		for i := range enc.Options {
			if pressedDigit(i + 1) {
				opt := enc.Options[i]
				if !opt.Enabled {
					g.sim.Log.Add(opt.DisableText, game.MsgWarning)
					break
				}
				result := g.sim.ResolveEncounterOption(i)
				if result == "TRADE" {
					// Special case: trader wants to trade — open station trade
					// For now, just log it since we need a proper trade interface
					enc.ResultText = "The trader opens a trade channel.\n(Space trading not yet available outside stations.)"
					enc.Resolved = true
				} else {
					enc.ResultText = result
				}
				break
			}
		}
	}

	g.drawScreen()

	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

// --- Episode view ---

func (g *Game) drawEpisodeView() {
	buf := g.buffer
	buf.Clear()

	ep := g.sim.ActiveEpisode
	if ep == nil {
		return
	}

	cx := 4

	// Title
	buf.WriteString(cx, 1, "=== SYSTEM EVENT ===", render.ColorLightCyan, render.ColorBlack)
	if ep.Resolved {
		buf.WriteString(gridCols-16, 1, "ESC: Continue", render.ColorDarkGray, render.ColorBlack)
	} else {
		buf.WriteString(gridCols-16, 1, "1-4: Choose", render.ColorDarkGray, render.ColorBlack)
	}

	buf.WriteString(cx, 2, "=========================================", render.ColorCyan, render.ColorBlack)

	// Episode title
	buf.WriteString(cx+1, 3, "*** "+ep.Title+" ***", render.ColorYellow, render.ColorBlack)

	// Briefing — word-wrapped
	wrapWidth := gridCols - cx - 6 // ~70 chars
	row := 5
	for _, line := range wrapText(ep.Briefing, wrapWidth) {
		if row >= commsRow-2 {
			break
		}
		buf.WriteString(cx+1, row, line, render.ColorWhite, render.ColorBlack)
		row++
	}
	row++
	buf.WriteString(cx, row, "=========================================", render.ColorCyan, render.ColorBlack)
	row += 2

	// Options
	for i, opt := range ep.Options {
		if row >= commsRow-2 {
			break
		}
		label := fmt.Sprintf(" %d. %s", i+1, opt.Label)
		optClr := uint8(render.ColorLightGray)
		if !opt.Enabled {
			optClr = render.ColorDarkGray
			label += " [" + opt.DisableText + "]"
		}
		if ep.Resolved {
			optClr = render.ColorDarkGray
		}
		buf.WriteString(cx, row, label, optClr, render.ColorBlack)
		row++
	}

	// Result text (after resolution)
	if ep.ResultText != "" {
		row++
		if row < commsRow-1 {
			buf.WriteString(cx, row, "-----------------------------------------", render.ColorCyan, render.ColorBlack)
			row++
		}
		for _, line := range wrapText(ep.ResultText, wrapWidth) {
			if row >= commsRow-1 {
				break
			}
			clr := uint8(render.ColorWhite)
			if ep.MLClue && (len(line) > 0 && (line[0] == '"' || line[0] == '\'')) {
				clr = render.ColorLightGreen
			}
			buf.WriteString(cx+1, row, line, clr, render.ColorBlack)
			row++
		}
		if row < commsRow-1 {
			buf.WriteString(cx, row, "-----------------------------------------", render.ColorCyan, render.ColorBlack)
		}
	}

	// Comms log
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		msgClr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, msgClr, render.ColorBlack)
	}

	if ep.Resolved {
		buf.WriteString(2, gridRows-1, "ESC: Continue to system map", render.ColorDarkGray, render.ColorBlack)
	} else {
		buf.WriteString(2, gridRows-1, "1-4: Choose option", render.ColorDarkGray, render.ColorBlack)
	}
}

func (g *Game) updateEpisode() error {
	ep := g.sim.ActiveEpisode

	// ESC exits after resolution (or skips the episode)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if ep != nil && ep.Resolved {
			g.sim.EndEpisode()
			g.viewMode = ViewSystemMap
			g.drawScreen()
			return nil
		}
		// If not resolved, ESC still lets you leave (choosing to ignore)
		g.sim.EndEpisode()
		g.viewMode = ViewSystemMap
		g.drawScreen()
		return nil
	}

	if ep != nil && !ep.Resolved {
		for i := range ep.Options {
			if pressedDigit(i + 1) {
				opt := ep.Options[i]
				if !opt.Enabled {
					g.sim.Log.Add(opt.DisableText, game.MsgWarning)
					break
				}
				result := g.sim.ResolveEpisodeOption(i)
				ep.ResultText = result
				break
			}
		}
	}

	g.drawScreen()

	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

// --- Surface View ---

func (g *Game) drawSurfaceView() {
	buf := g.buffer
	buf.Clear()

	surf := g.sim.ActiveSurface
	if surf == nil {
		return
	}

	// Camera centers on player (same pattern as ship view)
	ox := viewCenterX - surf.PlayerX
	oy := viewCenterY - surf.PlayerY

	// Render terrain grid
	render.RenderSurfaceGrid(buf, surf.Grid, surf.TerrainType, ox, oy)

	// Draw player
	buf.Set(viewCenterX, viewCenterY, '@', render.ColorWhite, render.ColorBlack)

	// Draw shuttle marker if visible
	shuttleScreenX := surf.ShuttleX + ox
	shuttleScreenY := surf.ShuttleY + oy
	if shuttleScreenX >= 0 && shuttleScreenX < gridCols && shuttleScreenY >= 0 && shuttleScreenY < gridRows {
		// Don't overwrite if player is standing on it
		if surf.PlayerX != surf.ShuttleX || surf.PlayerY != surf.ShuttleY {
			buf.Set(shuttleScreenX, shuttleScreenY, 'H', render.ColorWhite, render.ColorDarkGray)
		}
	}

	// --- Right panel: Objective and controls ---
	cx := panelX
	cy := 2

	// Title changes based on prologue or regular surface
	if g.sim.InPrologue() {
		buf.WriteString(cx, cy, "=== PROLOGUE ===", render.ColorYellow, render.ColorBlack)
		cy += 2
		// Show location
		locName := game.PrologueLocationName(g.sim.Prologue.Location)
		buf.WriteString(cx, cy, locName, render.ColorLightCyan, render.ColorBlack)
		cy += 2
		// Show prologue objectives
		buf.WriteString(cx, cy, "Shuttle Needs:", render.ColorYellow, render.ColorBlack)
		cy++
		ps := g.sim.PrologueSurface
		for _, obj := range g.sim.Prologue.GetObjectives() {
			found := false
			name := ""
			switch obj {
			case game.PrologueObjFuel:
				found = ps.FuelFound
				name = "Fuel Cells"
			case game.PrologueObjParts:
				found = ps.PartsFound
				name = "Spare Parts"
			case game.PrologueObjPower:
				found = ps.PowerFound
				name = "Power Pack"
			}
			status := "[ ]"
			clr := uint8(render.ColorWhite)
			if found {
				status = "[X]"
				clr = render.ColorGreen
			}
			buf.WriteString(cx, cy, fmt.Sprintf("%s %s", status, name), clr, render.ColorBlack)
			cy++
		}
		cy++
	} else {
		buf.WriteString(cx, cy, "=== SURFACE ===", render.ColorYellow, render.ColorBlack)
		cy += 2
		// Location info
		buf.WriteString(cx, cy, fmt.Sprintf("POI: %s", surf.POI), render.ColorLightCyan, render.ColorBlack)
		cy += 2
		// Objective
		if surf.Objective != nil {
			obj := surf.Objective
			objClr := uint8(render.ColorWhite)
			status := "[ ]"
			if obj.Complete {
				objClr = uint8(render.ColorGreen)
				status = "[X]"
			}
			buf.WriteString(cx, cy, "Objective:", render.ColorYellow, render.ColorBlack)
			cy++
			buf.WriteString(cx, cy, fmt.Sprintf("%s %s", status, obj.Description), objClr, render.ColorBlack)
			cy += 2
		}
	}

	// Loot collected
	if surf.LootCollected > 0 {
		buf.WriteString(cx, cy, fmt.Sprintf("Crates searched: %d", surf.LootCollected), render.ColorBrown, render.ColorBlack)
		cy += 2
	}

	// Standing on indicator
	tile := surf.GetTile(surf.PlayerX, surf.PlayerY)
	buf.WriteString(cx, cy, "Standing on:", render.ColorDarkGray, render.ColorBlack)
	cy++
	buf.WriteString(cx+1, cy, tile.Describe(), render.ColorLightGray, render.ColorBlack)

	// Controls
	cy = hudRow
	buf.WriteString(cx, cy, "--- Controls ---", render.ColorDarkGray, render.ColorBlack)
	cy++
	buf.WriteString(cx, cy, "WASD: Move", render.ColorDarkGray, render.ColorBlack)
	cy++
	buf.WriteString(cx, cy, "E: Interact", render.ColorDarkGray, render.ColorBlack)
	cy++
	buf.WriteString(cx, cy, "At shuttle: Lift off", render.ColorDarkGray, render.ColorBlack)
	cy += 2

	// Show if at shuttle
	if surf.AtShuttle() {
		if g.sim.InPrologue() {
			if g.sim.PrologueSurface.CheckPrologueComplete() {
				buf.WriteString(cx, cy, ">>> SHUTTLE READY <<<", render.ColorLightGreen, render.ColorBlack)
				cy++
				buf.WriteString(cx, cy, "Press E to LAUNCH!", render.ColorLightGreen, render.ColorBlack)
			} else {
				buf.WriteString(cx, cy, ">>> AT SHUTTLE <<<", render.ColorYellow, render.ColorBlack)
				cy++
				buf.WriteString(cx, cy, "Not ready yet...", render.ColorYellow, render.ColorBlack)
			}
		} else {
			buf.WriteString(cx, cy, ">>> AT SHUTTLE <<<", render.ColorLightGreen, render.ColorBlack)
			cy++
			buf.WriteString(cx, cy, "Press E to lift off", render.ColorLightGreen, render.ColorBlack)
		}
	}

	// --- Comms log ---
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}
}

func (g *Game) updateSurface() error {
	surf := g.sim.ActiveSurface
	if surf == nil {
		// Shouldn't happen, but safety fallback
		g.viewMode = ViewShip
		return nil
	}

	// Movement
	dx, dy := 0, 0
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		dy = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		dy = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		dx = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		dx = 1
	}
	if dx != 0 || dy != 0 {
		g.sim.TrySurfaceMove(dx, dy)
	}

	// Interact
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		if surf.AtShuttle() {
			// Check if we're in prologue
			if g.sim.InPrologue() {
				if g.sim.PrologueSurface.CheckPrologueComplete() {
					// Complete prologue and launch
					g.sim.CompletePrologue()
					g.viewMode = ViewShip
					return nil
				} else {
					// Can't launch yet
					g.sim.Log.Add("Shuttle not ready. "+g.sim.PrologueSurface.ObjectiveStatus(), game.MsgWarning)
				}
			} else {
				// Normal surface - lift off
				g.sim.LiftOff()
				g.viewMode = ViewShip
				return nil
			}
		} else {
			// Interact with tile
			if g.sim.InPrologue() {
				g.sim.PrologueInteract()
			} else {
				g.sim.SurfaceInteract()
			}
		}
	}

	// ESC shows reminder
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.sim.Log.Add("Return to the shuttle (H) to lift off.", game.MsgInfo)
	}

	// Tab → character sheet
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.prevViewMode = ViewSurface
		g.viewMode = ViewCharSheet
		return nil
	}

	g.drawScreen()

	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)

	return nil
}

// wrapText splits text by newlines and then word-wraps each line to maxWidth.
func wrapText(s string, maxWidth int) []string {
	var result []string
	for _, paragraph := range splitLines(s) {
		if len(paragraph) <= maxWidth {
			result = append(result, paragraph)
			continue
		}
		// Word-wrap this paragraph
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if len(line)+1+len(w) > maxWidth {
				result = append(result, line)
				line = w
			} else {
				line += " " + w
			}
		}
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// splitLines splits a string by newlines for multi-line rendering.
func splitLines(s string) []string {
	var lines []string
	line := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, line)
			line = ""
		} else {
			line += string(ch)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// --- Character Sheet view ---

func (g *Game) drawCharSheetView() {
	buf := g.buffer
	buf.Clear()

	cx := 2
	skills := &g.sim.Skills
	disc := g.sim.Discovery
	r := &g.sim.Resources

	// Title
	buf.WriteString(cx, 0, "--- CHARACTER SHEET ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(gridCols-16, 0, "Tab/ESC: Back", render.ColorDarkGray, render.ColorBlack)

	// Banner
	buf.WriteString(cx, 2, "=========================================", render.ColorCyan, render.ColorBlack)
	buf.WriteString(cx+1, 3, "COMMANDER'S LOG", render.ColorWhite, render.ColorBlack)
	buf.WriteString(cx+30, 3, fmt.Sprintf("Credits: %d", r.Credits), render.ColorYellow, render.ColorBlack)
	buf.WriteString(cx, 4, "=========================================", render.ColorCyan, render.ColorBlack)

	// Skills section
	buf.WriteString(cx, 6, "--- Skills ---", render.ColorLightCyan, render.ColorBlack)
	for i := game.SkillID(0); i < game.SkillCount; i++ {
		row := 7 + int(i)
		lvl := skills.Level(i)
		cur, needed := skills.XPProgress(i)
		g.drawSkillBar(buf, cx+1, row, game.SkillName(i), lvl, cur, needed)
	}

	// Discovery section (left panel)
	dRow := 15
	buf.WriteString(cx, dRow, "--- Discovery ---", render.ColorLightCyan, render.ColorBlack)
	dRow++

	// Star types with colored indicators
	starLine := fmt.Sprintf(" Star Types:  %d/5  ", disc.TotalStarTypesSeen)
	buf.WriteString(cx, dRow, starLine, render.ColorLightGray, render.ColorBlack)
	starLabels := []string{"Y", "R", "B", "W", "O"}
	starColors := []uint8{render.ColorYellow, render.ColorLightRed, render.ColorLightBlue, render.ColorWhite, render.ColorBrown}
	offset := cx + len(starLine)
	for j := 0; j < 5; j++ {
		if disc.StarTypesSeen[j] {
			buf.WriteString(offset, dRow, "["+starLabels[j]+"]", starColors[j], render.ColorBlack)
		} else {
			buf.WriteString(offset, dRow, "[ ]", render.ColorDarkGray, render.ColorBlack)
		}
		offset += 4
	}
	dRow++

	totalSystems := len(g.sim.Sector.Systems)
	buf.WriteString(cx, dRow, fmt.Sprintf(" Systems:     %d/%d explored",
		disc.TotalSystemsVisited, totalSystems), render.ColorLightGray, render.ColorBlack)
	dRow++
	buf.WriteString(cx, dRow, fmt.Sprintf(" Planets:     %d scanned", disc.TotalScans), render.ColorLightGray, render.ColorBlack)
	dRow++
	buf.WriteString(cx, dRow, fmt.Sprintf(" Stations:    %d docked", disc.TotalStationsDocked), render.ColorLightGray, render.ColorBlack)
	dRow++

	// Perks (right panel)
	perkX := 44
	buf.WriteString(perkX, 6, "--- Perks ---", render.ColorLightCyan, render.ColorBlack)
	perkRow := 7
	for i := game.SkillID(0); i < game.SkillCount; i++ {
		lvl := skills.Level(i)
		if lvl >= 2 {
			perk := game.SkillPerk(i, lvl)
			label := fmt.Sprintf("%s %d:", game.SkillName(i), lvl)
			buf.WriteString(perkX, perkRow, label, render.ColorWhite, render.ColorBlack)
			perkRow++
			buf.WriteString(perkX+1, perkRow, perk, render.ColorLightGray, render.ColorBlack)
			perkRow++
		}
	}
	if perkRow == 7 {
		buf.WriteString(perkX, perkRow, "Level up skills to unlock perks!", render.ColorDarkGray, render.ColorBlack)
	}

	// Recent scans
	dRow++
	buf.WriteString(cx, dRow, "--- Recent Scans ---", render.ColorLightCyan, render.ColorBlack)
	dRow++
	if len(disc.RecentScans) == 0 {
		buf.WriteString(cx+1, dRow, "No planets scanned yet.", render.ColorDarkGray, render.ColorBlack)
	} else {
		// Show up to 3 recent scans
		limit := 3
		if len(disc.RecentScans) < limit {
			limit = len(disc.RecentScans)
		}
		for i := 0; i < limit; i++ {
			scan := disc.RecentScans[i]
			buf.WriteString(cx+1, dRow, fmt.Sprintf("%s (%s)", scan.Name, game.PlanetKindName(scan.PlanetType)),
				render.ColorWhite, render.ColorBlack)
			dRow++
			buf.WriteString(cx+2, dRow, scan.Resources, render.ColorLightGray, render.ColorBlack)
			dRow++
			if scan.POI != "" {
				buf.WriteString(cx+2, dRow, "POI: "+scan.POI, render.ColorLightGreen, render.ColorBlack)
				dRow++
			}
		}
	}

	// Comms log
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}

	buf.WriteString(2, gridRows-1, "Tab/ESC: Back", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) drawSkillBar(buf *render.CellBuffer, x, y int, name string, level int, curXP, neededXP float64) {
	barW := 10
	filled := 0
	if neededXP > 0 {
		filled = int(float64(barW) * curXP / neededXP)
		if filled > barW {
			filled = barW
		}
	}
	if level >= 10 {
		filled = barW
	}

	// Skill name (padded to 12 chars)
	label := fmt.Sprintf("%-12s", name)
	clr := uint8(render.ColorLightGray)
	if level >= 5 {
		clr = render.ColorLightGreen
	}
	buf.WriteString(x, y, label, clr, render.ColorBlack)

	// Bar
	buf.Set(x+13, y, '[', render.ColorDarkGray, render.ColorBlack)
	for i := 0; i < barW; i++ {
		if i < filled {
			buf.Set(x+14+i, y, '#', clr, render.ColorBlack)
		} else {
			buf.Set(x+14+i, y, '.', render.ColorDarkGray, render.ColorBlack)
		}
	}
	buf.Set(x+14+barW, y, ']', render.ColorDarkGray, render.ColorBlack)

	// Level and XP
	if level >= 10 {
		buf.WriteString(x+26, y, "Lv 10 MAX", render.ColorYellow, render.ColorBlack)
	} else {
		info := fmt.Sprintf("Lv %d  (%.0f/%.0f)", level, curXP, neededXP)
		buf.WriteString(x+26, y, info, render.ColorDarkGray, render.ColorBlack)
	}
}

func (g *Game) updateCharSheet() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.viewMode = g.prevViewMode
		g.drawScreen()
		return nil
	}

	g.drawScreen()

	fps := fmt.Sprintf("FPS: %.0f  TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	g.buffer.WriteString(gridCols-20, gridRows-1, fps, render.ColorDarkGray, render.ColorBlack)
	return nil
}

// findObjectIndex returns the index of a SpaceObject in the system map's Objects slice.
func (g *Game) findObjectIndex(sm *game.SystemMap, target *game.SpaceObject) int {
	for i := range sm.Objects {
		if &sm.Objects[i] == target {
			return i
		}
	}
	return -1
}

// updateHoverInfo shows a description of whatever the mouse is hovering over.
func (g *Game) updateHoverInfo(cellX, cellY int) {
	infoY := 1
	for x := 0; x < gridCols; x++ {
		g.buffer.Set(x, infoY, ' ', render.ColorBlack, render.ColorBlack)
	}

	ox := g.cameraX()
	oy := g.cameraY()
	tileX := cellX - ox
	tileY := cellY - oy

	px, py := g.sim.PlayerPos()
	if tileX == px && tileY == py {
		g.buffer.WriteString(2, infoY, "@ You - a former redshirt, stranded in space", render.ColorWhite, render.ColorBlack)
		return
	}

	grid := g.sim.Grid
	if tileX >= 0 && tileX < grid.Width && tileY >= 0 && tileY < grid.Height {
		tile := grid.Get(tileX, tileY)
		if tile.Kind != world.TileVoid {
			desc := tile.Describe()
			coordInfo := fmt.Sprintf("  [%d,%d]", tileX, tileY)
			g.buffer.WriteString(2, infoY, desc+coordInfo, render.ColorYellow, render.ColorBlack)
			return
		}
	}

	g.buffer.WriteString(2, infoY, "Hover over the ship to inspect", render.ColorDarkGray, render.ColorBlack)
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen, g.buffer)
	// Render floating sprites (sub-pixel precision) on top of the cell grid
	for _, s := range g.sprites {
		g.renderer.DrawFloating(screen, s.Glyph, s.FG, s.PX, s.PY)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
