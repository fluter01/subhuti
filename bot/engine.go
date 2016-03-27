// Copyright 2016 Alex Fluter

package bot

import (
	"strings"
	"unicode"
)

type CommandEngine struct {
	BaseModule
	bot      *Bot
	inputCh  chan string
	commands map[string]func(string) error
}

func NewCommandEngine(bot *Bot) *CommandEngine {
	e := new(CommandEngine)
	e.bot = bot
	e.Name = "Engine"
	e.State = Initialized
	e.exitCh = make(chan bool)
	e.inputCh = make(chan string)
	e.Logger = bot.Logger

	e.commands = make(map[string]func(string) error)
	e.commands["SAVE"] = e.onSave
	e.commands["SHOW"] = e.onShow
	e.commands["STATUS"] = e.onStatus
	e.commands["EXIT"] = e.onExit
	e.commands["CONNECT"] = e.onConnect
	e.commands["DISCONNECT"] = e.onDisconnect
	e.commands["RECONNECT"] = e.onReconnect

	return e
}

func (e *CommandEngine) String() string {
	return "Engine"
}

func (e *CommandEngine) Init() error {
	return nil
}

func (e *CommandEngine) Start() error {
	e.Logger.Println("Command engine starting")
	return nil
}

func (e *CommandEngine) Stop() error {
	e.Logger.Println("Command engine stopping")
	e.exitCh <- true
	close(e.exitCh)
	close(e.inputCh)
	e.wait.Wait()
	e.State = Stopped
	return nil
}

func (e *CommandEngine) Status() string {
	return e.State.String()
}

func (e *CommandEngine) loop() {
	var exit bool
	var input string
	for !exit {
		select {
		case input = <-e.inputCh:
			e.handleCommand(input)
			break
		case exit = <-e.exitCh:
			break
		}
	}
	e.Logger.Println("Command engine loop exited")
	e.wait.Done()
}

func (e *CommandEngine) Run() {
	e.wait.Add(1)
	go e.loop()
	e.State = Running
}

func (e *CommandEngine) Submit(input string) {
	e.inputCh <- input
}

func (e *CommandEngine) handleCommand(input string) {
	var err error
	var data string

	if input[0] != e.bot.config.GetTrigger() {
		return
	}

	if unicode.IsSpace(rune(input[1])) {
		e.Logger.Print("Invalid command, space after trigger")
		return
	}

	cmd := input[1:]
	arr := strings.SplitN(cmd, " ", 2)
	cmd = arr[0]
	cmd = strings.ToUpper(cmd)
	if len(arr) == 2 {
		data = arr[1]
	} else {
		data = ""
	}
	f, ok := e.commands[cmd]
	if !ok {
		e.Logger.Print("Unknown command: ", cmd)
		return
	}

	e.Logger.Printf("Bot command %s", cmd)
	err = f(data)
	if err != nil {
		e.Logger.Printf("Failed to handle command %s: %s", cmd, err)
	}
}

// maitains the comand map
func (e *CommandEngine) AddCommand(cmd string, f func(string) error) {
	e.commands[cmd] = f
}

func (e *CommandEngine) DelCommand(cmd string) {
	delete(e.commands, cmd)
}

// bot command handlers
func (e *CommandEngine) onSave(string) error {
	return e.bot.config.Save(e.bot.config.path)
}

func (e *CommandEngine) onShow(string) error {
	e.Logger.Println("======== Config: ========")
	e.Logger.Printf("%#v\n", e.bot.config)
	e.Logger.Print("=========================")
	return nil
}

func (e *CommandEngine) onStatus(string) error {
	var mod Module
	e.Logger.Print("======== Status: ========")
	for _, mod = range e.bot.modules {
		e.Logger.Printf("Module %s %s", mod, mod.Status())
	}
	e.Logger.Print("=========================")
	return nil
}

func (e *CommandEngine) onExit(string) error {
	go e.bot.Stop()
	return nil
}

func (e *CommandEngine) onConnect(string) error {
	return nil
	//return bot.IRC().connect()
}

func (e *CommandEngine) onDisconnect(string) error {
	//bot.IRC().disconnect()
	return nil
}

func (e *CommandEngine) onReconnect(string) error {
	return nil
}
