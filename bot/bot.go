// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"strings"
	"time"
)

type Bot struct {
	BaseModule

	config   *BotConfig
	stdin    *Stdin
	engine   *CommandEngine
	modules  []Module
	handlers map[EventType]EventHandlers
	start    time.Time
	eventQ   chan *Event
}

func NewBot(name string, config *BotConfig) *Bot {
	var bot *Bot

	bot = new(Bot)
	bot.start = time.Now()
	bot.Name = name
	bot.config = config
	bot.eventQ = make(chan *Event)
	bot.exitCh = make(chan bool)
	bot.Logger = NewLoggerFunc(fmt.Sprintf("%s/%s-bot",
		bot.config.LogDir, bot.Name))
	bot.handlers = make(map[EventType]EventHandlers)

	bot.RegisterEventHandler(Input, bot.handleInput)
	bot.RegisterEventHandler(PrivateMessage, bot.handlePrivateMessage)
	bot.RegisterEventHandler(ChannelMessage, bot.handleChannelMessage)

	for k, v := range eventMap {
		for _, h := range v {
			bot.RegisterEventHandler(k, h)
		}
	}

	bot.stdin = NewStdin(bot)
	bot.engine = NewCommandEngine(bot)
	bot.modules = []Module{
		bot.stdin,
		bot.engine,
	}
	for i := range config.IRC {
		bot.modules = append(bot.modules, NewIRC(bot, config.IRC[i]))
	}

	bot.State = Initialized

	return bot
}

func (bot *Bot) String() string {
	return fmt.Sprintf("Bot %s: {%s} %s", bot.Name, bot.config, bot.State)
}

func (bot *Bot) Start() {
	var err error
	bot.Logger.Printf("bot %s starting", bot.Name)

	var mod Module
	for _, mod = range bot.modules {
		err = mod.Init()
		if err != nil {
			bot.Logger.Printf("module %s init failed: %s", mod, err)
			return
		}
		err = mod.Start()
		if err != nil {
			bot.Logger.Printf("module %s start failed: %s", mod, err)
			return
		}
		mod.Run()
		bot.Logger.Printf("Module %s running", mod)
	}

	bot.State = Running
	bot.wait.Add(1)
	bot.loop()
}

func (bot *Bot) loop() {
	var event *Event
	var exit bool
	for !exit {
		select {
		case event = <-bot.eventQ:
			if event == nil {
				continue
			}
			bot.handleEvent(event)
			break
		case exit = <-bot.exitCh:
			break
		}
	}
	bot.wait.Done()
	bot.Logger.Print("Bot exiting")
}

func (bot *Bot) Stop() {
	var err error
	var mod Module
	bot.Logger.Printf("bot %s stopping", bot.Name)
	for _, mod = range bot.modules {
		err = mod.Stop()
		if err != nil {
			bot.Logger.Printf("module %s stop failed: %s", mod, err)
		} else {
			bot.Logger.Printf("module %s stopped", mod)
		}
	}
	bot.State = Stopped
	bot.exitCh <- true
	bot.wait.Wait()
}

// events
func (bot *Bot) AddEvent(event *Event) {
	bot.eventQ <- event
}

func (bot *Bot) RegisterEventHandler(evt EventType, h EventHandler) {
	bot.handlers[evt] = append(bot.handlers[evt], h)
}

func (bot *Bot) handleEvent(event *Event) {
	var hs EventHandlers
	var h EventHandler

	hs, ok := bot.handlers[event.evt]
	if !ok {
		bot.Logger.Println("Unknown event type:", event.evt)
		return
	}
	if hs == nil {
		bot.Logger.Printf("%s ignored", event.evt)
		return
	}
	for _, h = range hs {
		h(event.data)
	}
}

// end events

func (bot *Bot) handleInput(data interface{}) {
	var input string

	bot.Logger.Println(data)
	input = data.(string)
	bot.engine.Submit(input)
}

func (bot *Bot) handlePrivateMessage(data interface{}) {
	privMsgData := data.(*PrivateMessageData)
	text := privMsgData.text
	trigger := privMsgData.irc.config.GetTrigger("")

	if !strings.HasPrefix(text, trigger) {
		text = fmt.Sprintf("%s%s", trigger, text)
	}
	req := MessageRequest{
		privMsgData.irc,
		false,
		privMsgData.from,
		privMsgData.nick,
		privMsgData.user,
		privMsgData.host,
		"",
		text,
		false}
	req.irc.interpreter.Submit(&req)
}

func (bot *Bot) handleChannelMessage(data interface{}) {
	chanMsgData := data.(*ChannelMessageData)

	req := MessageRequest{
		chanMsgData.irc,
		true,
		chanMsgData.from,
		chanMsgData.nick,
		chanMsgData.user,
		chanMsgData.host,
		chanMsgData.channel,
		chanMsgData.text,
		false}
	req.irc.interpreter.Submit(&req)
}
