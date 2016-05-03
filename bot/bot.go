// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const numWorkers = 16

type eventWorker struct {
	id      int
	eventCh chan *Event
	quitCh  chan bool
	wg      *sync.WaitGroup
}

func (w *eventWorker) start(bot *Bot) {
	for {
		select {
		case event := <-w.eventCh:
			if event == nil {
				continue
			}
			bot.handleEvent(event)
			break
		case <-w.quitCh:
			w.wg.Done()
			return
		}
	}
}

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

	// basic handles that keep the bot work
	bot.RegisterEventHandler(Input, bot.handleInput)
	bot.RegisterEventHandler(PrivateMessage, bot.handlePrivateMessage)
	bot.RegisterEventHandler(ChannelMessage, bot.handleChannelMessage)

	bot.stdin = NewStdin(bot)
	bot.engine = NewCommandEngine(bot)
	bot.modules = []Module{
		bot.stdin,
		bot.engine,
	}
	for i := range config.IRC {
		bot.modules = append(bot.modules, NewIRC(bot, config.IRC[i]))
	}

	// create addon modules
	for _, f := range initModuleFuncs {
		mod := f(bot)
		if mod != nil {
			bot.modules = append(bot.modules, mod)
		}
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
	var quit chan bool = make(chan bool)

	for i := 0; i < numWorkers; i++ {
		bot.wait.Add(1)
		w := &eventWorker{i, bot.eventQ, quit, &bot.wait}
		go w.start(bot)
	}

	exit := <-bot.exitCh
	if exit {
		close(quit)
	}

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
	if _, ok := bot.handlers[evt]; !ok {
		bot.handlers[evt] = &[2][]EventHandler{}
	}
	bot.handlers[evt][High] = append(bot.handlers[evt][High], h)
}

func (bot *Bot) RegisterEventHandlerPrio(evt EventType, h EventHandler, p Priority) {
	if _, ok := bot.handlers[evt]; !ok {
		bot.handlers[evt] = &[2][]EventHandler{}
	}
	bot.handlers[evt][p] = append(bot.handlers[evt][p], h)
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
	for _, phs := range hs {
		for _, h = range phs {
			h(event.data)
		}
	}
}

// end events

// event handlers
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
		irc:     privMsgData.irc,
		ischan:  false,
		from:    privMsgData.from,
		nick:    privMsgData.nick,
		user:    privMsgData.user,
		host:    privMsgData.host,
		channel: "",
		text:    text,
		direct:  false}
	req.irc.interpreter.Submit(&req)
}

func (bot *Bot) handleChannelMessage(data interface{}) {
	chanMsgData := data.(*ChannelMessageData)

	req := MessageRequest{
		irc:     chanMsgData.irc,
		ischan:  true,
		from:    chanMsgData.from,
		nick:    chanMsgData.nick,
		user:    chanMsgData.user,
		host:    chanMsgData.host,
		channel: chanMsgData.channel,
		text:    chanMsgData.text,
		direct:  false}
	req.irc.interpreter.Submit(&req)
}

func (bot *Bot) foreachIRC(f func(*IRC)) {
	for _, m := range bot.modules {
		if irc, ok := m.(*IRC); ok {
			f(irc)
		}
	}
}
