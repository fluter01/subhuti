// Copyright 2016 Alex Fluter

package bot

import (
	"flag"
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
	f.bot.RegisterEventHandler(MessageParseEvent, f.handleMessage)
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

// factchange <channel> <keyword> s/<pattern>/<change to>/
// factchange <channel> <keyword> <new desc>
func (f *FactoidProcessor) factchange(req *MessageRequest, args string) (string, error) {
	var channel, keyword, newdesc string

	arr := strings.SplitN(args, " ", 3)

	if len(arr) < 3 {
		return "Usage: factadd <channel> <keyword> s/<pattern>/<change to>/", nil
	}

	channel, keyword, newdesc = arr[0], arr[1], arr[2]

	f.Logger.Println("change:", channel, keyword)

	fact := &Factoid{
		Network: req.irc.config.Name,
		Owner:   req.from,
		Nick:    req.nick,
		Channel: channel,
		Keyword: keyword,
		Desc:    newdesc,
	}

	if err := f.factoids.Change(fact); err != nil {
		f.Logger.Println("change error:", err)
		return err.Error(), nil
	}

	return "", nil
}

// factfind [-channel channel] [-owner nick] [-by nick] [text]
func (f *FactoidProcessor) factfind(req *MessageRequest, args string) (string, error) {
	var (
		channel string
		owner   string
		by      string
		text    string
	)

	arr := strings.Split(args, " ")

	flags := flag.NewFlagSet("factchange", flag.ContinueOnError)
	flags.StringVar(&channel, "channel", "", "the channel")
	flags.StringVar(&owner, "onwer", "", "the channel")
	flags.StringVar(&by, "by", "", "the channel")
	flags.StringVar(&text, "text", "", "the channel")
	err := flags.Parse(arr)
	if err != nil {
		f.Logger.Println("find error:", err)
		return "Usage: factfind [-channel channel] [-owner nick] [-by nick] [text]", nil
	}

	if flags.NArg() < 1 {
		return "Usage: factfind [-channel channel] [-owner nick] [-by nick] [text]", nil
	}

	text = flags.Arg(0)
	//	if channel == "" {
	//		channel = req.channel
	//	}

	fact := &Factoid{
		Network: req.irc.config.Name,
		Channel: channel,
		Nick:    owner,
		RefUser: by,
		Keyword: text,
	}

	var facts []*Factoid

	facts, err = f.factoids.Find(fact)
	if err != nil {
		f.Logger.Println("find error:", err)
		return err.Error(), nil
	}

	if len(facts) == 0 {
		return "No factoids found", nil
	}

	var result []string
	for _, fact = range facts {
		result = append(result, fmt.Sprintf("[%s] %s",
			fact.Channel, fact.Keyword))
	}

	return strings.Join(result, " "), nil
}

// factinfo [channel] <keyword>
func (f *FactoidProcessor) factinfo(req *MessageRequest, args string) (string, error) {
	var (
		channel string
		keyword string
	)

	arr := strings.SplitN(args, " ", 2)

	if len(arr) < 1 {
		return "Usage: factinfo [channel] <keyword>", nil
	}

	if len(arr) == 2 {
		channel, keyword = arr[0], arr[1]
	} else {
		channel, keyword = "global", arr[0]
	}

	fact := &Factoid{
		Network: req.irc.config.Name,
		Channel: channel,
		Keyword: keyword,
	}

	factoid, err := f.factoids.Get(fact)
	if err != nil {
		f.Logger.Println("find error:", err)
		return err.Error(), nil
	}

	chanstr := factoid.Channel
	if chanstr == "global" {
		chanstr = "all channels"
	}
	return fmt.Sprintf("%s: Factoid submitted by %s for %s on %s,"+
		" referenced %d times (last by %s on %s)",
		factoid.Keyword, factoid.Nick, chanstr, factoid.Created,
		factoid.RefCount, factoid.RefUser, factoid.RefTime), nil
}

func (f *FactoidProcessor) factshow(req *MessageRequest, args string) (string, error) {
	var (
		channel string
		keyword string
	)

	arr := strings.SplitN(args, " ", 2)

	if len(arr) < 1 {
		return "Usage: factshow [channel] <keyword>", nil
	}

	if len(arr) == 2 {
		channel, keyword = arr[0], arr[1]
	} else {
		channel, keyword = "global", arr[0]
	}

	fact := &Factoid{
		Network: req.irc.config.Name,
		Channel: channel,
		Keyword: keyword,
	}

	factoid, err := f.factoids.Get(fact)
	if err != nil {
		f.Logger.Println("find error:", err)
		return err.Error(), nil
	}

	return fmt.Sprintf("%s: %s", factoid.Keyword, factoid.Desc), nil
}

func (f *FactoidProcessor) factset(req *MessageRequest, args string) (string, error) {
	return "not implemented yet", nil
}

// fact <channel> <keyword> [arguments]
func (f *FactoidProcessor) factcall(req *MessageRequest, args string) (string, error) {
	var (
		channel string
		keyword string
	)

	arr := strings.SplitN(args, " ", 2)

	if len(arr) < 1 {
		return "Usage: fact [channel] <keyword>", nil
	}

	if len(arr) == 2 {
		channel, keyword = arr[0], arr[1]
	} else if req.ischan {
		channel, keyword = req.channel, arr[0]
	} else {
		channel, keyword = "global", arr[0]
	}

	fact := &Factoid{
		Network: req.irc.config.Name,
		Channel: channel,
		Keyword: keyword,
	}

	factoid, err := f.factoids.Get(fact)
	if err != nil {
		f.Logger.Println("find error:", err)
		return err.Error(), nil
	}
	req.prefix = false
	return factoid.Desc, nil
}

func (f *FactoidProcessor) handleMessage(data interface{}) {
	req, ok := data.(*MessageRequest)
	if !ok {
		return
	}

	if req.keyword == "" {
		return
	}

	fact := &Factoid{
		Network: req.irc.config.Name,
		Channel: req.channel,
		Keyword: req.keyword,
	}
	if req.channel == "" {
		fact.Channel = "global"
	}

	factoid, err := f.factoids.Get(fact)
	if err != nil {
		f.Logger.Println("find error:", err)
		return
	}
	req.prefix = false
	req.irc.sendReply(factoid.Desc, req)
	return
}
