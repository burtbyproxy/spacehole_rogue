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
	panelX   = 60 // right-side legend panel
	hudRow   = 27 // systems HUD
	commsRow = 33 // message log
	commsMax = 8  // max visible messages
)

// Game is the Ebitengine game struct. It owns rendering and input.
// All gameplay state lives in sim.
type Game struct {
	atlas    *render.FontAtlas
	renderer *render.GridRenderer
	buffer   *render.CellBuffer
	sim      *game.Sim
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

func (g *Game) drawScreen() {
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
	buf.WriteString(gridCols-19, 0, "Phase 3: Systems", render.ColorDarkGray, render.ColorBlack)

	// Legend (fixed right panel)
	buf.WriteString(panelX, 3, "Legend:", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(panelX, 4, " # Hull Wall", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(panelX, 5, " + Door", render.ColorBrown, render.ColorBlack)
	buf.WriteString(panelX, 6, " = Console", render.ColorGreen, render.ColorBlack)
	buf.WriteString(panelX, 7, " b Berth", render.ColorBrown, render.ColorBlack)
	buf.WriteString(panelX, 8, " t Toilet", render.ColorLightGray, render.ColorBlack)
	buf.WriteString(panelX, 9, " s Shower", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(panelX, 10, " R Replicator", render.ColorLightGreen, render.ColorBlack)
	buf.WriteString(panelX, 11, " w Water Recycler", render.ColorLightBlue, render.ColorBlack)
	buf.WriteString(panelX, 12, " f Food Stores", render.ColorGreen, render.ColorBlack)
	buf.WriteString(panelX, 13, " E Engine", render.ColorLightRed, render.ColorBlack)
	buf.WriteString(panelX, 14, " P Power Cell", render.ColorYellow, render.ColorBlack)
	buf.WriteString(panelX, 15, " W Water Tank", render.ColorBlue, render.ColorBlack)

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
	buf.WriteString(panelX, 17, "--- Equipment ---", render.ColorLightCyan, render.ColorBlack)
	drawEquipStatus(buf, panelX, 18, "Engine", g.sim.EngineOn)
	drawEquipStatus(buf, panelX, 19, "Water Recycler", g.sim.WaterRecyclerOn)
	drawEquipStatus(buf, panelX, 20, "Replicator", g.sim.ReplicatorOn)

	// Player needs (right panel, below equipment)
	n := &g.sim.Needs
	buf.WriteString(panelX, 22, "--- Status ---", render.ColorLightCyan, render.ColorBlack)
	drawNeedBar(buf, panelX, 23, "Hunger ", n.Hunger)
	drawNeedBar(buf, panelX, 24, "Thirst ", n.Thirst)
	drawNeedBar(buf, panelX, 25, "Hygiene", n.Hygiene)

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

func (g *Game) Update() error {
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

	// Tick simulation (resource drain, etc.)
	g.sim.Tick()

	// Redraw every frame (resources change over time)
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
