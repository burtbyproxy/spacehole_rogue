package render

import "image/color"

// CGA 16-color palette indices.
const (
	ColorBlack        = 0
	ColorBlue         = 1
	ColorGreen        = 2
	ColorCyan         = 3
	ColorRed          = 4
	ColorMagenta      = 5
	ColorBrown        = 6
	ColorLightGray    = 7
	ColorDarkGray     = 8
	ColorLightBlue    = 9
	ColorLightGreen   = 10
	ColorLightCyan    = 11
	ColorLightRed     = 12
	ColorLightMagenta = 13
	ColorYellow       = 14
	ColorWhite        = 15
)

// Palette contains the classic CGA 16-color palette.
var Palette = [16]color.RGBA{
	{0, 0, 0, 255},       // 0: Black
	{0, 0, 170, 255},     // 1: Blue
	{0, 170, 0, 255},     // 2: Green
	{0, 170, 170, 255},   // 3: Cyan
	{170, 0, 0, 255},     // 4: Red
	{170, 0, 170, 255},   // 5: Magenta
	{170, 85, 0, 255},    // 6: Brown
	{170, 170, 170, 255}, // 7: Light Gray
	{85, 85, 85, 255},    // 8: Dark Gray
	{85, 85, 255, 255},   // 9: Light Blue
	{85, 255, 85, 255},   // 10: Light Green
	{85, 255, 255, 255},  // 11: Light Cyan
	{255, 85, 85, 255},   // 12: Light Red
	{255, 85, 255, 255},  // 13: Light Magenta
	{255, 255, 85, 255},  // 14: Yellow
	{255, 255, 255, 255}, // 15: White
}
