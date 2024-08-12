package ui

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type basicDisplay struct {
	V      string
	title  string
	valw   material.LabelStyle
	theme  *material.Theme
	height int
	color  color.NRGBA
}

func newBasicDisplay(theme *material.Theme, title string, height int) *basicDisplay {
	return &basicDisplay{
		title:  title,
		valw:   material.Label(theme, unit.Sp(height), "0"),
		theme:  theme,
		height: height,
	}
}

func (b *basicDisplay) Display(gtx C, visible bool) func(C) D {
	switch {
	case !visible:
		b.color = color.NRGBA{0, 0, 0, 0}
	case b.V == "" || b.V == "0" || b.V == "0.0":
		b.color = color.NRGBA{100, 100, 100, 70}
	default:
		b.color = color.NRGBA{0, 0, 0, 255}
	}
	return func(gtx C) D {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceEnd,
		}.Layout(gtx,
			layout.Rigid(
				func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(b.theme, unit.Sp(26), b.title)
					lbl.Color = b.color
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}),
			layout.Rigid(
				func(gtx layout.Context) layout.Dimensions {
					b.valw.Text = b.V
					b.valw.Color = b.color
					b.valw.Alignment = text.Middle
					return b.valw.Layout(gtx)
				},
			),
		)
	}
}
