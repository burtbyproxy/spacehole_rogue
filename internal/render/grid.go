package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Cell represents a single character cell on screen.
type Cell struct {
	Glyph byte  // CP437 code (0-255)
	FG    uint8 // Foreground color index (0-15)
	BG    uint8 // Background color index (0-15)
}

// CellBuffer is a 2D grid of character cells.
type CellBuffer struct {
	Cols  int
	Rows  int
	Cells []Cell
}

// NewCellBuffer creates a new cell buffer filled with blank cells.
func NewCellBuffer(cols, rows int) *CellBuffer {
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i] = Cell{Glyph: ' ', FG: ColorWhite, BG: ColorBlack}
	}
	return &CellBuffer{Cols: cols, Rows: rows, Cells: cells}
}

// Set writes a single cell at (x, y). Out-of-bounds writes are ignored.
func (b *CellBuffer) Set(x, y int, glyph byte, fg, bg uint8) {
	if x >= 0 && x < b.Cols && y >= 0 && y < b.Rows {
		b.Cells[y*b.Cols+x] = Cell{Glyph: glyph, FG: fg, BG: bg}
	}
}

// Get reads a single cell at (x, y). Out-of-bounds reads return a blank cell.
func (b *CellBuffer) Get(x, y int) Cell {
	if x >= 0 && x < b.Cols && y >= 0 && y < b.Rows {
		return b.Cells[y*b.Cols+x]
	}
	return Cell{}
}

// Clear resets all cells to blank (space on black).
func (b *CellBuffer) Clear() {
	for i := range b.Cells {
		b.Cells[i] = Cell{Glyph: ' ', FG: ColorWhite, BG: ColorBlack}
	}
}

// WriteString writes a string starting at (x, y). Each rune occupies one cell.
func (b *CellBuffer) WriteString(x, y int, s string, fg, bg uint8) {
	offset := 0
	for _, ch := range s {
		if ch > 255 {
			ch = '?'
		}
		b.Set(x+offset, y, byte(ch), fg, bg)
		offset++
	}
}

// GridRenderer draws a CellBuffer to an Ebitengine screen.
type GridRenderer struct {
	Atlas   *FontAtlas
	CellW   int
	CellH   int
	bgPixel *ebiten.Image // 1x1 white pixel for drawing backgrounds
}

// NewGridRenderer creates a renderer with the given atlas and cell dimensions.
func NewGridRenderer(atlas *FontAtlas, cellW, cellH int) *GridRenderer {
	bgPixel := ebiten.NewImage(1, 1)
	bgPixel.Fill(color.White)
	return &GridRenderer{
		Atlas:   atlas,
		CellW:   cellW,
		CellH:   cellH,
		bgPixel: bgPixel,
	}
}

// Draw renders the entire CellBuffer to the screen.
func (r *GridRenderer) Draw(screen *ebiten.Image, buf *CellBuffer) {
	scaleX := float64(r.CellW) / float64(GlyphWidth)
	scaleY := float64(r.CellH) / float64(GlyphHeight)

	var op ebiten.DrawImageOptions

	for y := 0; y < buf.Rows; y++ {
		for x := 0; x < buf.Cols; x++ {
			cell := buf.Cells[y*buf.Cols+x]
			px := float64(x * r.CellW)
			py := float64(y * r.CellH)

			// Draw background color
			if cell.BG != ColorBlack {
				op = ebiten.DrawImageOptions{}
				op.GeoM.Scale(float64(r.CellW), float64(r.CellH))
				op.GeoM.Translate(px, py)
				op.ColorScale.ScaleWithColor(Palette[cell.BG])
				screen.DrawImage(r.bgPixel, &op)
			}

			// Draw foreground glyph
			if cell.Glyph != ' ' && cell.Glyph != 0 {
				glyph := r.Atlas.Glyph(cell.Glyph)
				op = ebiten.DrawImageOptions{}
				op.GeoM.Scale(scaleX, scaleY)
				op.GeoM.Translate(px, py)
				op.ColorScale.ScaleWithColor(Palette[cell.FG])
				screen.DrawImage(glyph, &op)
			}
		}
	}
}

// DrawFloating renders a single glyph at sub-pixel screen coordinates.
// Used for entities that move smoothly between tile cells (e.g. ships in system map).
func (r *GridRenderer) DrawFloating(screen *ebiten.Image, glyph byte, fg uint8, px, py float64) {
	if glyph == ' ' || glyph == 0 {
		return
	}
	g := r.Atlas.Glyph(glyph)
	scaleX := float64(r.CellW) / float64(GlyphWidth)
	scaleY := float64(r.CellH) / float64(GlyphHeight)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(px, py)
	op.ColorScale.ScaleWithColor(Palette[fg])
	screen.DrawImage(g, &op)
}
