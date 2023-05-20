package ui

import (
	"context"
	"fmt"
	"image/color"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/grafov/kiwi"
	"github.com/wt-tools/wt-tactics/config"
	"github.com/wt-tools/wtscope/action"
	"github.com/wt-tools/wtscope/input/hudmsg"
)

var headings []string

type battleLog struct {
	w          *app.Window
	th         *material.Theme
	cfg        *config.Config
	log        *kiwi.Logger
	grid       component.GridState
	list       widget.List
	rows       []action.GeneralAction
	latestTime time.Duration
}

func newBattleLog(cfg *config.Config, log *kiwi.Logger) *battleLog {
	return &battleLog{
		w:   app.NewWindow(app.Title("WT Scope: Battle Log")),
		th:  material.NewTheme(gofont.Collection()),
		cfg: cfg,
		log: log,
	}
}

func (g *gui) UpdateBattleLog(ctx context.Context, gamelog *hudmsg.Service) {
	l := g.log.New()
	go func() {
		for {
			select {
			case data := <-gamelog.Messages:
				if len(g.bl.rows) > 0 && g.bl.latestTime > data.At {
					// Reset log on a new battle session.
					g.bl.rows = nil
				}
				g.bl.latestTime = data.At
				g.bl.rows = append(g.bl.rows, data)
				g.bl.w.Invalidate()
				l.Log("battle log", data)
			}
		}
	}()
}

func (b *battleLog) panel() error {
	var ops op.Ops
	b.list.Axis = layout.Vertical
	b.list.ScrollToEnd = true
	for {
		e := <-b.w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			if len(b.rows) == 0 {
				continue
			}
			gtx := layout.NewContext(&ops, e)
			b.listLayout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (b *battleLog) listLayout(gtx C) D {
	return material.List(b.th, &b.list).Layout(gtx, len(b.rows), func(gtx layout.Context, i int) layout.Dimensions {
		var (
			text string
			act  row
		)
		switch {
		case len(b.rows) == 0:
			text = "no battle log yet"
			return material.Label(b.th, unit.Sp(26), text).Layout(gtx)
		case i > len(b.rows): // TODO broken case, handle this in other way
			// text = fmtAction(b.rows[len(b.rows)-1])
			act = row(b.rows[len(b.rows)-1])
		default:
			// text = fmtAction(b.rows[i])
			act = row(b.rows[i])
		}
		return act.rowDisplay(gtx, b.cfg.CurrentPlayer(), b.th)
	})
}

type row action.GeneralAction

func (r row) rowDisplay(gtx C, playerName string, th *material.Theme) D {
	return layout.UniformInset(10).Layout(gtx,
		func(gtx C) D {
			return layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Horizontal,
				Spacing:   layout.SpaceEvenly,
			}.Layout(gtx,
				// Timestamp
				layout.Rigid(
					func(gtx C) D {
						return layout.UniformInset(10).Layout(gtx,
							material.Label(th, unit.Sp(14), r.At.String()).Layout,
						)
					},
				),
				layout.Flexed(0.9,
					func(gtx C) D {
						playerInfo := material.Label(th, unit.Sp(20), fmt.Sprintf("%s %s", r.Damage.Player.Clan, r.Damage.Player.Name))
						playerInfo.Color = color.NRGBA{0, 0, 0, 255}
						if r.Damage.Player.Name == playerName {
							playerInfo.Color = color.NRGBA{0, 255, 0, 255}
						}
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(gtx,
							// Raw log row
							layout.Rigid(material.Label(th, unit.Sp(14), r.Origin).Layout),
							// Player - action - player info
							layout.Rigid(
								func(gtx C) D {
									return layout.Flex{
										Alignment: layout.Middle,
										Axis:      layout.Horizontal,
										Spacing:   layout.SpaceEvenly,
									}.Layout(gtx,
										// Initiator player
										layout.Flexed(0.2,
											func(gtx C) D {
												return layout.Flex{
													Alignment: layout.Middle,
													Axis:      layout.Vertical,
													Spacing:   layout.SpaceEnd,
												}.Layout(gtx,
													layout.Rigid(material.Label(th, unit.Sp(26), r.Damage.Vehicle.Name).Layout),
													layout.Rigid(playerInfo.Layout),
												)
											},
										),
										// Action
										//		layout.Inset{0, 0, 0, 0}.Layout(gtx,
										layout.Flexed(0.5,
											material.Label(th, unit.Sp(28), r.Damage.ActionRaw).Layout),
										// Target player
										layout.Flexed(0.2,
											func(gtx C) D {
												switch {
												case r.Achievement != nil && r.Achievement.Name != "":
													return layout.Flex{
														Alignment: layout.Middle,
														Axis:      layout.Vertical,
														Spacing:   layout.SpaceStart,
													}.Layout(gtx,
														layout.Rigid(material.Label(th, unit.Sp(26), r.Achievement.Name).Layout),
													)
												case r.Damage.TargetVehicle.Name != "":
													return layout.Flex{
														Alignment: layout.Middle,
														Axis:      layout.Vertical,
														Spacing:   layout.SpaceStart,
													}.Layout(gtx,
														layout.Rigid(material.Label(th, unit.Sp(26), r.Damage.TargetVehicle.Name).Layout),
														layout.Rigid(material.Label(th, unit.Sp(20), fmt.Sprintf("%s %s", r.Damage.TargetPlayer.Clan, r.Damage.TargetPlayer.Name)).Layout),
													)
												}
												return layout.Flex{}.Layout(gtx)
											},
										),
									)
								}))
					}))
		})
}

func fmtAction(a action.GeneralAction) string {
	return fmt.Sprintf("%16s  %s", a.At, a.Origin)
}
