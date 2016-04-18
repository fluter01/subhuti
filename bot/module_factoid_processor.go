// Copyright 2016 Alex Fluter

package bot

import (
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
	return nil
}

func (f *FactoidProcessor) Stop() error {
	f.Logger.Println("FactoidProcessor stopped")
	f.State = Stopped
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
	f.Logger.Println(arr)
	f.Logger.Println(len(arr))

	if len(arr) < 3 {
		return "Usage: factadd <channel> <keyword> <description>", nil
	}

	channel, keyword, desc = arr[0], arr[1], arr[2]

	f.Logger.Println("add:", channel, keyword, desc)

	fact := &Factoid{
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
		return "", err
	}

	return "", nil
}

func (f *FactoidProcessor) factrem(req *MessageRequest, args string) (string, error) {
	return "", nil
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
