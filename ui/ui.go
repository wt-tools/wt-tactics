package ui

import (
	"context"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"github.com/grafov/kiwi"
	"github.com/wt-tools/wtradar/config"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type gui struct {
	log *kiwi.Logger
	bl  *battleLog
}

func Init(_ context.Context, conf *config.Config, log *kiwi.Logger) *gui {
	return &gui{
		log: log,
		bl:  newBattleLog(conf, log),
	}
}

func (g *gui) Run(_ context.Context) {
	l := g.log.New()
	go func() {
		err := g.bl.panel()
		if err != nil {
			l.Log("fatal", "can't run battle log window", "error", err)
			os.Exit(0)
		}
	}()
	app.Main()
}
