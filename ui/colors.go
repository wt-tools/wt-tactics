package ui

import "image/color"

var (
	bgColor       = color.NRGBA{R: 192, G: 192 + 32, B: 192, A: 255}
	evenRowColor  = color.NRGBA{R: 192, G: 192 + 16, B: 192, A: 255}
	oddRowColor   = color.NRGBA{R: 127, G: 127 + 16, B: 127, A: 255}
	thinLineColor = color.NRGBA{A: 255}
)
