package main

import (
	"context"
	"os"
	"time"

	"github.com/wt-tools/wtradar/ui"
	"github.com/wt-tools/wtscope/config"
	"github.com/wt-tools/wtscope/input/gamechat"
	"github.com/wt-tools/wtscope/input/hudmsg"
	"github.com/wt-tools/wtscope/net/dedup"
	"github.com/wt-tools/wtscope/net/poll"

	"github.com/grafov/kiwi"
)

func main() {
	ctx := context.Background()
	kiwi.SinkTo(os.Stdout, kiwi.AsLogfmt()).Start()
	l := kiwi.New()
	errch := make(chan error, 8) // XXX разделить по компонентам
	conf, err := config.Load(errch)
	if err != nil {
		l.Log("status", "can't load configuration", "path", config.ConfPath)
		if err := config.CreateIfAbsent(); err != nil {
			os.Exit(1)
		}
		l.Log("status", "default configuration created")
		l.Log("hint", "check the file and fill it with your real config values", "path", config.ConfPath)
		os.Exit(0)
	}
	l.Log("status", "turning on radar", "config", conf.Dump())
	go showErrors(l, errch)
	defaultPolling := poll.New(poll.SetLogger(errch),
		poll.SetLoopDelay(250*time.Millisecond), poll.SetProblemDelay(4*time.Second))
	go defaultPolling.Do()
	gui := ui.Init(ctx, conf, l)

	{
		battleSvc := hudmsg.New(conf, defaultPolling, dedup.New(), errch)
		go battleSvc.Grab(ctx)
		gui.UpdateBattleLog(ctx, battleSvc)
	}
	{
		chatSvc := gamechat.New(conf, defaultPolling, dedup.New(), errch)
		go chatSvc.Grab(ctx)
		gui.UpdateGameChat(ctx, chatSvc)
	}
	gui.Run(ctx)
}

func showErrors(log *kiwi.Logger, errs chan error) {
	l := log.New()
	for {
		err := <-errs
		l.Log("problem", "wtscope failure", "error", err)
	}
}
