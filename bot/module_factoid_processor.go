// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type FactoidProcessor struct {
	BaseModule
	bot      *Bot
	factoids *Factoids
}

func init() {
	RegisterInitModuleFunc(NewFactoidProcessor)
}

func NewFactoidProcessor(bot *Bot) Module {
	f := new(FactoidProcessor)
	f.bot = bot
	f.Name = "FactoidProcessor"
	f.Logger = bot.Logger
	f.factoids = NewFactoids(bot.config.DataDir + "/" + bot.config.DB)
	return f
}

func (f *FactoidProcessor) Init() error {
	f.Logger.Println("Initializing FactoidProcessor")
	f.State = Initialized
	return nil
}

func (f *FactoidProcessor) Start() error {
	f.Logger.Println("Starting FactoidProcessor")
	f.bot.foreachIRC(func(irc *IRC) {
		irc.interpreter.RegisterCommand("factadd", f.factadd)
		irc.interpreter.RegisterCommand("factrem", f.factrem)
		irc.interpreter.RegisterCommand("factchange", f.factchange)
		irc.interpreter.RegisterCommand("factfind", f.factfind)
		irc.interpreter.RegisterCommand("factinfo", f.factinfo)
		irc.interpreter.RegisterCommand("factshow", f.factshow)
		irc.interpreter.RegisterCommand("factset", f.factset)
		irc.interpreter.RegisterCommand("fact", f.factcall)
	})
	f.State = Running
	f.factoids.Dump(os.Stderr)
	return nil
}

func (f *FactoidProcessor) Stop() error {
	f.Logger.Println("FactoidProcessor stopped")
	f.State = Stopped
	f.factoids.Dump(os.Stderr)
	return nil
}

func (f *FactoidProcessor) String() string {
	return f.Name
}

func (f *FactoidProcessor) Status() string {
	return f.State.String()
}

func (f *FactoidProcessor) Run() {
}

// factadd <channel> <keyword> <factoid...>
func (f *FactoidProcessor) factadd(req *MessageRequest, args string) (string, error) {
	var channel, keyword, desc string

	arr := strings.SplitN(args, " ", 3)

	if len(arr) < 3 {
		return "Usage: factadd <channel> <keyword> <description>", nil
	}

	channel, keyword, desc = arr[0], arr[1], arr[2]

	f.Logger.Println("add:", channel, keyword, desc)

	fact := &Factoid{
		Network:  req.irc.config.Name,
		Owner:    req.from,
		Nick:     req.nick,
		Channel:  channel,
		Keyword:  keyword,
		Desc:     desc,
		Created:  time.Now(),
		RefCount: 0,
		RefUser:  "none",
		Enabled:  true,
	}

	if err := f.factoids.Add(fact); err != nil {
		f.Logger.Println("add error:", err)
		return err.Error(), nil
	}

	return fmt.Sprintf("factoid %s added to %s",
		keyword, channel), nil
}

func (f *FactoidProcessor) factrem(req *MessageRequest, args string) (string, error) {
	var channel, keyword string

	arr := strings.SplitN(args, " ", 2)

	if len(arr) < 2 {
		return "Usage: factadd <channel> <keyword>", nil
	}

	channel, keyword = arr[0], arr[1]

	f.Logger.Println("remove:", channel, keyword)

	fact := &Factoid{
		Network: req.irc.config.Name,
		Owner:   req.from,
		Nick:    req.nick,
		Channel: channel,
		Keyword: keyword,
	}

	if err := f.factoids.Remove(fact); err != nil {
		f.Logger.Println("remove error:", err)
		return err.Error(), nil
	}

	return fmt.Sprintf("factoid %s removed from %s",
		keyword, channel), nil
}

func (f *FactoidProcessor) factchange(req *MessageRequest, args string) (string, error) {
	return "", nil
}

func (f *FactoidProcessor) factfind(req *MessageRequest, args string) (string, error) {
	return "", nil
}

func (f *FactoidProcessor) factinfo(req *MessageRequest, args string) (string, error) {
	return "", nil
}

func (f *FactoidProcessor) factshow(req *MessageRequest, args string) (string, error) {
	return "", nil
}

func (f *FactoidProcessor) factset(req *MessageRequest, args string) (string, error) {
	return "", nil
}

func (f *FactoidProcessor) factcall(req *MessageRequest, args string) (string, error) {
	return "", nil
}
