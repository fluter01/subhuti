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
	bot.Logger = NewLogger(fmt.Sprintf("%s-bot", bot.Name))
	bot.Logger = NewLogger("stdout")
	bot.handlers = make(map[EventType]EventHandlers)

	bot.RegisterEventHandler(Input, bot.handleInput)
	bot.RegisterEventHandler(PrivateMessage, bot.handlePrivateMessage)
	bot.RegisterEventHandler(ChannelMessage, bot.handleChannelMessage)
	bot.RegisterEventHandler(UserJoin, HandleUserJoin)
	bot.RegisterEventHandler(UserPart, HandleUserPart)
	bot.RegisterEventHandler(UserQuit, HandleUserQuit)
	bot.RegisterEventHandler(UserNick, HandleUserNick)
	bot.RegisterEventHandler(Pong, HandlePong)

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
	bot.loop()
}

func (bot *Bot) loop() {
	var event *Event
	var exit bool
	bot.wait.Add(1)
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
	bot.exitCh <- true
	bot.wait.Wait()
	bot.State = Stopped
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

	input = data.(string)
	bot.engine.Submit(input)
}

func (bot *Bot) handlePrivateMessage(data interface{}) {
	privMsgData := data.(*PrivateMessageData)
	text := privMsgData.text
	trigger := bot.config.GetIRC("TODO").GetTrigger("")

	if !strings.HasPrefix(text, trigger) {
		text = fmt.Sprintf("%s%s", trigger, text)
	}
	req := NewMessageRequest(
		nil,
		false,
		privMsgData.from,
		privMsgData.nick,
		privMsgData.user,
		privMsgData.host,
		"",
		text)
	req.irc.interpreter.RequestChan() <- req
}

func (bot *Bot) handleChannelMessage(data interface{}) {
	chanMsgData := data.(*ChannelMessageData)

	req := NewMessageRequest(
		nil,
		true,
		chanMsgData.from,
		chanMsgData.nick,
		chanMsgData.user,
		chanMsgData.host,
		chanMsgData.channel,
		chanMsgData.text)
	req.irc.interpreter.RequestChan() <- req
}
