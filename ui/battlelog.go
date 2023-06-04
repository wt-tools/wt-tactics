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
	"github.com/grafov/kiwi"
	"github.com/wt-tools/wtscope/action"
	"github.com/wt-tools/wtscope/events"
	"github.com/wt-tools/wtscope/input/hudmsg"
)

var headings []string

type battleLog struct {
	w                   *app.Window
	th                  *material.Theme
	cfg                 configurator
	log                 *kiwi.Logger
	listAll, listPlayer widget.List
	rowsAll, rowsPlayer []events.Event
	tropes              map[string]int
	latestTime          time.Duration
}

func newBattleLog(cfg configurator, log *kiwi.Logger) *battleLog {
	return &battleLog{
		w:      app.NewWindow(app.Title("WT Scope: Battle Log")),
		th:     material.NewTheme(gofont.Collection()),
		tropes: make(map[string]int),
		cfg:    cfg,
		log:    log,
	}
}

// TODO move this logic to another package out of UI
func (g *gui) UpdateBattleLog(ctx context.Context, gamelog *hudmsg.Service) {
	l := g.log.New()
	go func() {
		for {
			select {
			case data := <-gamelog.Messages:
				if (len(g.bl.rowsAll)+len(g.bl.rowsPlayer) > 0) && g.bl.latestTime > data.At {
					// Reset log on a new battle session.
					l.Log("new battle has began")
					g.bl.rowsAll = nil
					g.bl.rowsPlayer = nil
					g.bl.tropes = make(map[string]int)
				}
				g.bl.latestTime = data.At
				switch {
				case data.Player.Name == g.bl.cfg.PlayerName():
					if (data.Action == action.Destroyed ||
						data.Action == action.ShotDown) && data.TargetVehicle.Name != "" {
						g.bl.tropes[data.TargetVehicle.Name]++
					}
					fallthrough
				case data.TargetPlayer.Name == g.bl.cfg.PlayerName():
					g.bl.rowsPlayer = append(g.bl.rowsPlayer, data)
				default:
					g.bl.rowsAll = append(g.bl.rowsAll, data)
				}

				g.bl.w.Invalidate()
				l.Log("battle log", data)
			}
		}
	}()
}

func (b *battleLog) panel() error {
	var ops op.Ops
	b.listPlayer.Axis = layout.Vertical
	b.listPlayer.ScrollToEnd = true
	b.listAll.Axis = layout.Vertical
	b.listAll.ScrollToEnd = true
	for {
		e := <-b.w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceEvenly,
			}.Layout(gtx,
				layout.Rigid(b.header("Team battle log")),
				layout.Flexed(0.6, b.battleLogLayout),
				layout.Rigid(b.header("Trophies")),
				layout.Flexed(0.1, b.myTrophies),
				layout.Rigid(b.header("Personal battle log")),
				layout.Flexed(0.3, b.myLogLayout),
			)
			e.Frame(gtx.Ops)
		}
	}
}

func (b *battleLog) myTrophies(gtx C) D {
	var tropes []layout.FlexChild
	for name, times := range b.tropes {
		val := name
		if times > 1 {
			val = fmt.Sprintf("%s x %d", val, times)
		}
		l := material.Label(b.th, unit.Sp(26), val)
		l.Color = color.NRGBA{192, 192, 0, 255} // yellow
		tropes = append(tropes, layout.Rigid(func(gtx C) D { return layout.UniformInset(10).Layout(gtx, l.Layout) }))
	}
	return layout.Flex{
		Alignment: layout.Start,
		Axis:      layout.Horizontal,
		Spacing:   layout.SpaceEvenly,
	}.Layout(gtx, tropes...)
}

func (b *battleLog) header(title string) func(C) D {
	return func(gtx C) D {
		return layout.UniformInset(10).Layout(gtx,
			material.Label(b.th, unit.Sp(28), title).Layout,
		)
	}
}

func (b *battleLog) myLogLayout(gtx C) D {
	return material.List(b.th, &b.listPlayer).Layout(gtx, len(b.rowsPlayer), func(gtx layout.Context, i int) layout.Dimensions {
		var (
			text string
			act  row
		)
		switch {
		case len(b.rowsPlayer) == 0:
			text = "no personal battle log yet"
			return material.Label(b.th, unit.Sp(26), text).Layout(gtx)
		case i > len(b.rowsPlayer): // TODO broken case, handle this in other way
			act = row(b.rowsPlayer[len(b.rowsPlayer)-1])
		default:
			act = row(b.rowsPlayer[i])
		}
		return act.rowDisplay(gtx, b.cfg.PlayerName(), b.th)
	})
}

func (b *battleLog) battleLogLayout(gtx C) D {
	return material.List(b.th, &b.listAll).Layout(gtx, len(b.rowsAll), func(gtx layout.Context, i int) layout.Dimensions {
		var (
			text string
			act  row
		)
		switch {
		case len(b.rowsAll) == 0:
			text = "no battle log yet"
			return material.Label(b.th, unit.Sp(26), text).Layout(gtx)
		case i > len(b.rowsAll): // TODO broken case, handle this in other way
			// text = fmtAction(b.rows[len(b.rows)-1])
			act = row(b.rowsAll[len(b.rowsAll)-1])
		default:
			// text = fmtAction(b.rows[i])
			act = row(b.rowsAll[i])
		}
		return act.rowDisplay(gtx, b.cfg.PlayerName(), b.th)
	})
}

type row events.Event

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
						playerInfo := material.Label(th, unit.Sp(20), fmt.Sprintf("%s %s", r.Player.Squad, r.Player.Name))
						playerInfo.Color = color.NRGBA{0, 0, 0, 255} // black
						if r.Player.Name == playerName {
							playerInfo.Color = color.NRGBA{0, 255, 0, 255} // green
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
													layout.Rigid(material.Label(th, unit.Sp(26), r.Vehicle.Name).Layout),
													layout.Rigid(playerInfo.Layout),
												)
											},
										),
										// Action
										//		layout.Inset{0, 0, 0, 0}.Layout(gtx,
										layout.Flexed(0.5,
											material.Label(th, unit.Sp(28), r.ActionText).Layout),
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
												case r.TargetVehicle.Name != "":
													return layout.Flex{
														Alignment: layout.Middle,
														Axis:      layout.Vertical,
														Spacing:   layout.SpaceStart,
													}.Layout(gtx,
														layout.Rigid(material.Label(th, unit.Sp(26), r.TargetVehicle.Name).Layout),
														layout.Rigid(material.Label(th, unit.Sp(20), fmt.Sprintf("%s %s", r.TargetPlayer.Squad, r.TargetPlayer.Name)).Layout),
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

func fmtRawEvent(e events.Event) string {
	return fmt.Sprintf("%16s  %s", e.At, e.Origin)
}
