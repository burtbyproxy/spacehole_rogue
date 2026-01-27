package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	GlyphWidth  = 16
	GlyphHeight = 16
	AtlasCols   = 16
	AtlasRows   = 16
)

// FontAtlas holds the CP437 glyph atlas and cached sub-images.
type FontAtlas struct {
	image  *ebiten.Image
	glyphs [256]*ebiten.Image
}

// NewFontAtlas generates a CP437 font atlas at startup.
// ASCII characters (32-126) are rendered with basicfont.Face7x13.
// Box-drawing and block characters are drawn manually.
func NewFontAtlas() *FontAtlas {
	atlasW := AtlasCols * GlyphWidth  // 256
	atlasH := AtlasRows * GlyphHeight // 256

	img := image.NewNRGBA(image.Rect(0, 0, atlasW, atlasH))
	face := basicfont.Face7x13

	for code := 0; code < 256; code++ {
		col := code % AtlasCols
		row := code / AtlasCols
		cx := col * GlyphWidth
		cy := row * GlyphHeight

		r := CP437ToUnicode[code]

		// ASCII printable range: render with basicfont
		if r >= 32 && r <= 126 {
			drawFontGlyph(img, face, cx, cy, r)
			continue
		}

		// Box-drawing characters
		if bc, ok := boxChars[byte(code)]; ok {
			drawBoxGlyph(img, cx, cy, bc[0], bc[1], bc[2], bc[3])
			continue
		}

		// Block elements and shading
		drawBlockGlyph(img, cx, cy, byte(code))
	}

	eimg := ebiten.NewImageFromImage(img)
	a := &FontAtlas{image: eimg}

	// Cache sub-images for each glyph
	for code := 0; code < 256; code++ {
		col := code % AtlasCols
		row := code / AtlasCols
		x := col * GlyphWidth
		y := row * GlyphHeight
		rect := image.Rect(x, y, x+GlyphWidth, y+GlyphHeight)
		a.glyphs[code] = eimg.SubImage(rect).(*ebiten.Image)
	}

	return a
}

// Glyph returns the cached sub-image for a CP437 character code.
func (a *FontAtlas) Glyph(code byte) *ebiten.Image {
	return a.glyphs[code]
}

// drawFontGlyph renders a single ASCII character into the atlas.
// basicfont.Face7x13 glyphs are 7x13, centered in a 16x16 cell.
func drawFontGlyph(img *image.NRGBA, face font.Face, cellX, cellY int, r rune) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot:  fixed.P(cellX+4, cellY+13), // centered horizontally, baseline at y+13
	}
	d.DrawString(string(r))
}

// boxChars maps CP437 codes to single-line box connection flags: {left, right, top, bottom}.
var boxChars = map[byte][4]bool{
	179: {false, false, true, true},  // │
	180: {true, false, true, true},   // ┤
	191: {true, false, false, true},  // ┐
	192: {false, true, true, false},  // └
	193: {true, true, true, false},   // ┴
	194: {true, true, false, true},   // ┬
	195: {false, true, true, true},   // ├
	196: {true, true, false, false},  // ─
	197: {true, true, true, true},    // ┼
	217: {true, false, true, false},  // ┘
	218: {false, true, false, true},  // ┌
}

// drawBoxGlyph draws a single-line box-drawing character.
// Lines are 2 pixels wide, centered in the 16x16 cell.
func drawBoxGlyph(img *image.NRGBA, cellX, cellY int, left, right, top, bottom bool) {
	w := color.NRGBA{255, 255, 255, 255}
	cx := cellX + 7
	cy := cellY + 7

	if left {
		for x := cellX; x < cx+2; x++ {
			img.SetNRGBA(x, cy, w)
			img.SetNRGBA(x, cy+1, w)
		}
	}
	if right {
		for x := cx; x < cellX+GlyphWidth; x++ {
			img.SetNRGBA(x, cy, w)
			img.SetNRGBA(x, cy+1, w)
		}
	}
	if top {
		for y := cellY; y < cy+2; y++ {
			img.SetNRGBA(cx, y, w)
			img.SetNRGBA(cx+1, y, w)
		}
	}
	if bottom {
		for y := cy; y < cellY+GlyphHeight; y++ {
			img.SetNRGBA(cx, y, w)
			img.SetNRGBA(cx+1, y, w)
		}
	}
}

// drawBlockGlyph draws block elements and shading characters.
func drawBlockGlyph(img *image.NRGBA, cellX, cellY int, code byte) {
	w := color.NRGBA{255, 255, 255, 255}

	switch code {
	case 176: // ░ Light shade
		for y := 0; y < GlyphHeight; y++ {
			for x := 0; x < GlyphWidth; x++ {
				if (x+y)%4 == 0 {
					img.SetNRGBA(cellX+x, cellY+y, w)
				}
			}
		}
	case 177: // ▒ Medium shade
		for y := 0; y < GlyphHeight; y++ {
			for x := 0; x < GlyphWidth; x++ {
				if (x+y)%2 == 0 {
					img.SetNRGBA(cellX+x, cellY+y, w)
				}
			}
		}
	case 178: // ▓ Dark shade
		for y := 0; y < GlyphHeight; y++ {
			for x := 0; x < GlyphWidth; x++ {
				if (x+y)%4 != 0 {
					img.SetNRGBA(cellX+x, cellY+y, w)
				}
			}
		}
	case 219: // █ Full block
		for y := 0; y < GlyphHeight; y++ {
			for x := 0; x < GlyphWidth; x++ {
				img.SetNRGBA(cellX+x, cellY+y, w)
			}
		}
	case 220: // ▄ Lower half
		for y := GlyphHeight / 2; y < GlyphHeight; y++ {
			for x := 0; x < GlyphWidth; x++ {
				img.SetNRGBA(cellX+x, cellY+y, w)
			}
		}
	case 221: // ▌ Left half
		for y := 0; y < GlyphHeight; y++ {
			for x := 0; x < GlyphWidth/2; x++ {
				img.SetNRGBA(cellX+x, cellY+y, w)
			}
		}
	case 222: // ▐ Right half
		for y := 0; y < GlyphHeight; y++ {
			for x := GlyphWidth / 2; x < GlyphWidth; x++ {
				img.SetNRGBA(cellX+x, cellY+y, w)
			}
		}
	case 223: // ▀ Upper half
		for y := 0; y < GlyphHeight / 2; y++ {
			for x := 0; x < GlyphWidth; x++ {
				img.SetNRGBA(cellX+x, cellY+y, w)
			}
		}
	case 254: // ■ Small square
		for y := 4; y < 12; y++ {
			for x := 4; x < 12; x++ {
				img.SetNRGBA(cellX+x, cellY+y, w)
			}
		}
	case 127: // ⌂ Open box / house (bed glyph)
		// Draw a simple open-top box: U-shape with a hat
		// Top line (roof)
		for x := 3; x < 13; x++ {
			img.SetNRGBA(cellX+x, cellY+3, w)
			img.SetNRGBA(cellX+x, cellY+4, w)
		}
		// Left wall
		for y := 3; y < 13; y++ {
			img.SetNRGBA(cellX+3, cellY+y, w)
			img.SetNRGBA(cellX+4, cellY+y, w)
		}
		// Right wall
		for y := 3; y < 13; y++ {
			img.SetNRGBA(cellX+11, cellY+y, w)
			img.SetNRGBA(cellX+12, cellY+y, w)
		}
		// Bottom line
		for x := 3; x < 13; x++ {
			img.SetNRGBA(cellX+x, cellY+11, w)
			img.SetNRGBA(cellX+x, cellY+12, w)
		}
	}
}
