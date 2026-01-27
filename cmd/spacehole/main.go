package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/spacehole-rogue/spacehole_rogue/assets"
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

type Game struct {
	ticks    uint64
	atlas    *render.FontAtlas
	renderer *render.GridRenderer
	buffer   *render.CellBuffer
	grid     *world.TileGrid
	layout   *world.ShipLayout
	playerX  int
	playerY  int
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
	tileGrid := layout.ToTileGrid()

	g := &Game{
		atlas:    atlas,
		renderer: renderer,
		buffer:   buffer,
		grid:     tileGrid,
		layout:   layout,
		playerX:  layout.SpawnX(),
		playerY:  layout.SpawnY(),
	}

	g.drawScreen()
	return g
}

const (
	// Viewport center — where the player @ always appears on screen.
	viewCenterX = gridCols / 2 // 40
	viewCenterY = 13

	// Fixed UI positions
	panelX  = 60 // right-side legend panel
	hudRow  = 27 // systems HUD
	commsRow = 33 // message log
)

// cameraX/Y returns the world-to-screen offset so the player is always centered.
func (g *Game) cameraX() int { return viewCenterX - g.playerX }
func (g *Game) cameraY() int { return viewCenterY - g.playerY }

func (g *Game) drawScreen() {
	buf := g.buffer
	buf.Clear()

	// --- World layer (camera-relative) ---
	ox := g.cameraX()
	oy := g.cameraY()
	render.RenderTileGrid(buf, g.grid, ox, oy)

	// Player — always at viewport center
	buf.Set(viewCenterX, viewCenterY, '@', render.ColorWhite, render.ColorBlack)

	// --- Fixed UI layer ---

	// Title bar
	buf.WriteString(2, 0, title, render.ColorWhite, render.ColorBlack)
	nameLabel := fmt.Sprintf("[ %s ]", g.layout.Name)
	buf.WriteString(20, 0, nameLabel, render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(gridCols-19, 0, "Phase 2: Shuttle", render.ColorDarkGray, render.ColorBlack)

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

	// HUD resource bars (fixed bottom area)
	buf.WriteString(2, hudRow, "--- Systems ---", render.ColorLightCyan, render.ColorBlack)
	drawBar(buf, 2, hudRow+1, "Water ", 78, render.ColorBlue)
	drawBar(buf, 2, hudRow+2, "Energy", 95, render.ColorYellow)
	drawBar(buf, 2, hudRow+3, "Food  ", 42, render.ColorGreen)
	drawBar(buf, 2, hudRow+4, "Hull  ", 100, render.ColorLightGray)

	// Message log (fixed)
	buf.WriteString(2, commsRow, "--- Comms ---", render.ColorLightCyan, render.ColorBlack)
	buf.WriteString(2, commsRow+1, "Systems nominal. Welcome aboard the Nomad.", render.ColorCyan, render.ColorBlack)
	buf.WriteString(2, commsRow+2, "Water recycler online. Food replicator online.", render.ColorCyan, render.ColorBlack)
	buf.WriteString(2, commsRow+3, "Power cell at 95%. All systems green.", render.ColorCyan, render.ColorBlack)

	// Instructions
	buf.WriteString(2, gridRows-1, "WASD/Arrows: Move  ESC: Quit", render.ColorDarkGray, render.ColorBlack)
}

func drawBar(buf *render.CellBuffer, x, y int, label string, pct int, clr uint8) {
	buf.WriteString(x, y, label, render.ColorLightGray, render.ColorBlack)
	barW := 20
	filled := barW * pct / 100
	for i := 0; i < barW; i++ {
		if i < filled {
			buf.Set(x+7+i, y, 219, clr, render.ColorBlack) // █
		} else {
			buf.Set(x+7+i, y, 176, render.ColorDarkGray, render.ColorBlack) // ░
		}
	}
	s := fmt.Sprintf("%3d%%", pct)
	buf.WriteString(x+28, y, s, render.ColorLightGray, render.ColorBlack)
}

func (g *Game) Update() error {
	g.ticks++

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
		newX := g.playerX + dx
		newY := g.playerY + dy
		if g.grid.IsWalkable(newX, newY) {
			g.playerX = newX
			g.playerY = newY
			g.drawScreen()
		}
	}

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
	// Info bar row (row 1, just under the title)
	infoY := 1
	// Clear the info bar
	for x := 0; x < gridCols; x++ {
		g.buffer.Set(x, infoY, ' ', render.ColorBlack, render.ColorBlack)
	}

	ox := g.cameraX()
	oy := g.cameraY()
	tileX := cellX - ox
	tileY := cellY - oy

	// Check if hovering over the player
	if tileX == g.playerX && tileY == g.playerY {
		g.buffer.WriteString(2, infoY, "@ You - a former redshirt, stranded in space", render.ColorWhite, render.ColorBlack)
		return
	}

	// Check if hovering over a ship tile
	if tileX >= 0 && tileX < g.grid.Width && tileY >= 0 && tileY < g.grid.Height {
		tile := g.grid.Get(tileX, tileY)
		if tile.Kind != world.TileVoid {
			desc := tile.Describe()
			coordInfo := fmt.Sprintf("  [%d,%d]", tileX, tileY)
			g.buffer.WriteString(2, infoY, desc+coordInfo, render.ColorYellow, render.ColorBlack)
			return
		}
	}

	// Not over anything interesting
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
