package main

import (
	"fmt"
	"log"

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

	sim := game.NewSim(layout)

	g := &Game{
		atlas:    atlas,
		renderer: renderer,
		buffer:   buffer,
		sim:      sim,
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
	// Player body indicator
	fullness := r.BodyFullness()
	if fullness > 0 {
		bodyClr := uint8(render.ColorDarkGray)
		if r.TotalWaste() >= 20 {
			bodyClr = render.ColorYellow
		}
		buf.WriteString(2, hudRow+5, fmt.Sprintf("Body: %d/%d full (%d waste)",
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

	// Message log (live from sim)
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	msgs := g.sim.Log.Recent(commsMax)
	for i, msg := range msgs {
		clr := msgColor(msg.Priority)
		buf.WriteString(2, commsRow+1+i, msg.Text, clr, render.ColorBlack)
	}

	// Instructions
	buf.WriteString(2, gridRows-1, "WASD: Move  E: Interact  T: Toggle  ESC: Quit", render.ColorDarkGray, render.ColorBlack)
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
		case game.ObjStation:
			buf.WriteString(infoX, row, "Station (docking: future)", render.ColorCyan, render.ColorBlack)
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
	buf.WriteString(2, gridRows-1, "WASD: Fly  E: Interact  N: Nav Map  ESC: Ship Interior", render.ColorDarkGray, render.ColorBlack)
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
	default:
		return g.updateShip()
	}
}

func (g *Game) updateShip() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
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

	// Pilot console → system map (fly around the star system)
	if g.sim.PilotActivated {
		g.sim.PilotActivated = false
		g.sim.Sector.EnsureSystemMap(g.sim.Sector.CurrentSystem)
		g.viewMode = ViewSystemMap
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
				g.viewMode = ViewSystemMap
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
		}
	}

	// E near object → approach interaction
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		sm := g.sim.Sector.CurrentSystemMap()
		obj := sm.NearestObject(sm.Shuttle.TileX(), sm.Shuttle.TileY(), 3)
		if obj != nil {
			g.logApproachInfo(obj)
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
		g.sim.Log.Add(fmt.Sprintf("Hailing %s. Docking available (future).", obj.Name), game.MsgInfo)
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
