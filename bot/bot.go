// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

const (
	MOD_IRC int = iota
	MOD_STDIN
	//	MOD_NET
	MOD_INTERPRETER

	// count
	MOD_LAST
)

type Bot struct {
	name   string
	config *BotConfig

	modules  []Module
	handlers []EventHandlers
	cmdMap   CmdMap
	start    time.Time
	logger   Logger
	eventQ   chan *Event
	quit     chan bool
}

func NewBot(name string, config *BotConfig) *Bot {
	var bot *Bot

	bot = new(Bot)
	bot.start = time.Now()
	bot.name = name
	bot.config = config
	bot.quit = make(chan bool)

	bot.modules = make([]Module, MOD_LAST)
	bot.modules[MOD_IRC] = NewIRC(bot)
	bot.modules[MOD_STDIN] = NewStdin(bot)
	bot.modules[MOD_INTERPRETER] = NewInterpreter(bot)

	bot.handlers = make([]EventHandlers, EventCount)
	bot.RegisterEventHandler(UserInput, bot.handleUserInput)
	bot.RegisterEventHandler(PrivateMessage, bot.handlePrivateMessage)
	bot.RegisterEventHandler(ChannelMessage, bot.handleChannelMessage)
	bot.RegisterEventHandler(UserJoin, HandleUserJoin)
	bot.RegisterEventHandler(UserPart, HandleUserPart)
	bot.RegisterEventHandler(UserQuit, HandleUserQuit)
	bot.RegisterEventHandler(UserNick, HandleUserNick)
	bot.RegisterEventHandler(Pong, HandlePong)

	bot.cmdMap = make(map[string]CmdFunc)

	bot.cmdMap["SAVE"] = bot.onSave
	bot.cmdMap["SHOW"] = bot.onShow
	bot.cmdMap["STATUS"] = bot.onStatus
	bot.cmdMap["EXIT"] = bot.onExit
	bot.cmdMap["CONNECT"] = bot.onConnect
	bot.cmdMap["DISCONNECT"] = bot.onDisconnect
	bot.cmdMap["RECONNECT"] = bot.onReconnect

	bot.eventQ = make(chan *Event)

	bot.logger = NewStdFileLogger(bot, "bot")

	return bot
}

func (bot *Bot) String() string {
	return fmt.Sprintf("Bot %s: %s", bot.name, bot.config)
}

func (bot *Bot) Start() {
	var err error
	bot.Logger().Printf("bot %s starting", bot.name)

	var mod Module
	for _, mod = range bot.modules {
		err = mod.Init()
		if err != nil {
			bot.Logger().Printf("module %s init failed: %s", mod, err)
			return
		}
		err = mod.Start()
		if err != nil {
			bot.Logger().Printf("module %s start failed: %s", mod, err)
			return
		}
		mod.Run()
		bot.Logger().Printf("Module %s running", mod)
	}

	bot.Loop()
}

func (bot *Bot) Loop() {
	var event *Event
	for {
		event = bot.GetEvent()
		if event == nil {
			break
		}
		//bot.Logger().Printf("Event %s", event)
		bot.handleEvent(event)
		//bot.handlers[event.evt](event.data)
	}
	bot.Logger().Print("Bot loop exiting")
}

func (bot *Bot) IRC() *IRC {
	mod := bot.modules[MOD_IRC]
	return mod.(*IRC)
}

func (bot *Bot) Interpreter() *Interpreter {
	mod := bot.modules[MOD_INTERPRETER]
	return mod.(*Interpreter)
}

func (bot *Bot) Logger() Logger {
	return bot.logger
}

// event methods
func (bot *Bot) handleEvent(event *Event) {
	var hs EventHandlers
	var h EventHandler

	if event.evt < EventCount {
		hs = bot.handlers[event.evt]
		if hs == nil {
			bot.Logger().Printf("BUG: %s unhandled", event.evt)
			return
		}
		for _, h = range hs {
			h(event.data)
		}
	} else {
		bot.Logger().Println("Unknown event type:", event.evt)
	}
}

func (bot *Bot) AddEvent(event *Event) {
	bot.eventQ <- event
}

func (bot *Bot) GetEvent() *Event {
	return <-bot.eventQ
}

func (bot *Bot) RegisterEventHandler(evt EventType, h EventHandler) {
	bot.handlers[evt] = append(bot.handlers[evt], h)
}

// end event methods

func (bot *Bot) handleUserInput(data interface{}) {
	var input string
	var first byte

	input = data.(string)
	input = strings.TrimSpace(input)
	first = input[0]

	if unicode.IsSpace(rune(input[1])) {
		bot.Logger().Print("Invalid command, space after trigger")
		return
	}
	if first == bot.config.BotTrigger {
		bot.handleCommand(input[1:])
	} else if first == bot.config.IRCTrigger {
		bot.IRC().handleCommand(input[1:])
	} else {
		fmt.Println("Unknow commands")
	}
}

func (bot *Bot) handlePrivateMessage(data interface{}) {
	privMsgData := data.(*PrivateMessageData)
	text := privMsgData.text
	trigger := bot.config.Trigger("")

	if !strings.HasPrefix(text, trigger) {
		text = fmt.Sprintf("%s%s", trigger, text)
	}
	req := NewMessageRequest(
		bot.IRC(),
		false,
		privMsgData.from,
		privMsgData.nick,
		privMsgData.user,
		privMsgData.host,
		"",
		text)
	bot.Interpreter().RequestChan() <- req
}

func (bot *Bot) handleChannelMessage(data interface{}) {
	chanMsgData := data.(*ChannelMessageData)

	req := NewMessageRequest(
		bot.IRC(),
		true,
		chanMsgData.from,
		chanMsgData.nick,
		chanMsgData.user,
		chanMsgData.host,
		chanMsgData.channel,
		chanMsgData.text)
	bot.Interpreter().RequestChan() <- req
}

func (bot *Bot) handleCommand(cmd string) {
	var err error
	var arr []string
	var data string
	var f CmdFunc
	var ok bool

	arr = strings.SplitN(cmd, " ", 2)
	cmd = arr[0]
	cmd = strings.ToUpper(cmd)
	if len(arr) == 2 {
		data = arr[1]
	} else {
		data = ""
	}
	f, ok = bot.cmdMap[cmd]
	if !ok {
		bot.Logger().Print("Unhandled command: ", cmd)
		return
	}

	bot.Logger().Printf("Bot command %s", cmd)
	err = f(data)
	if err != nil {
		bot.Logger().Printf("Failed to handle command %s: %s", cmd, err)
	}
}

// bot command handlers
func (bot *Bot) onSave(string) error {
	return bot.config.Save()
}

func (bot *Bot) onShow(string) error {
	bot.Logger().Println("======== Config: ========")
	bot.Logger().Printf("%#v\n", bot.config)
	bot.Logger().Print("=========================")
	return nil
}

func (bot *Bot) onStatus(string) error {
	var mod Module
	bot.Logger().Print("======== Status: ========")
	for _, mod = range bot.modules {
		bot.Logger().Printf("Module %s %s", mod, mod.Status())
	}
	bot.Logger().Print("=========================")
	return nil
}

func (bot *Bot) onExit(string) error {
	var err error
	var mod Module
	for _, mod = range bot.modules {
		bot.Logger().Printf("Stopping module %s", mod)
		err = mod.Stop()
		if err != nil {
			bot.Logger().Printf("Stop module %s failed: %s", mod, err)
		}
	}
	close(bot.eventQ)
	return nil
}

func (bot *Bot) onConnect(string) error {
	return bot.IRC().connect()
}

func (bot *Bot) onDisconnect(string) error {
	bot.IRC().disconnect()
	return nil
}

func (bot *Bot) onReconnect(string) error {
	bot.IRC().disconnect()
	return bot.IRC().connect()
}

// end event handlers
