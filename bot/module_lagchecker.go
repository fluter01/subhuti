// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const lagCheckMarker = ":LAGCHECK"

type LagChecker struct {
	BaseModule
	bot *Bot
}

func init() {
	RegisterInitModuleFunc(NewLagChecker)
}

func NewLagChecker(bot *Bot) Module {
	lc := new(LagChecker)
	lc.bot = bot
	lc.Name = "LagChecker"
	lc.Logger = bot.Logger

	return lc
}

func (lc *LagChecker) Init() error {
	lc.Logger.Println("Initializing LagChecker")
	lc.State = Initialized
	return nil
}

func (lc *LagChecker) Start() error {
	lc.Logger.Println("Starting LagChecker")
	lc.bot.RegisterEventHandler(Pong, lc.handlePong)
	lc.bot.foreachIRC(func(irc *IRC) {
		irc.interpreter.RegisterCommand("lagcheck", lc.run)
	})
	lc.State = Running
	return nil
}

func (lc *LagChecker) Stop() error {
	lc.Logger.Println("LagChecker stopped")
	lc.State = Stopped
	return nil
}

func (lc *LagChecker) String() string {
	return lc.Name
}

func (lc *LagChecker) Status() string {
	return lc.State.String()
}

func (lc *LagChecker) Run() {
}

func (lc *LagChecker) handlePong(data interface{}) {

	pong, ok := data.(*PongData)
	if !ok {
		lc.Logger.Println("corrupted pong data")
		return
	}
	if strings.HasPrefix(pong.origin, lagCheckMarker[1:]) {
		now := time.Now().UnixNano()
		then, err := strconv.ParseInt(pong.origin[len(lagCheckMarker):], 10, 64)
		if err != nil {
			lc.Logger.Println("invalid timestamp")
			return
		}
		d := now - then
		lag := time.Duration(d) * time.Nanosecond
		lc.Logger.Printf("%s Lag %s", pong.irc, lag)
		pong.irc.lag = lag
	}
}

func (lc *LagChecker) run(irc *IRC, args string) (string, error) {
	err := irc.Ping(fmt.Sprintf("%s %d", lagCheckMarker, time.Now().UnixNano()))
	return "", err
}
