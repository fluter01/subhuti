// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"regexp"
	"strings"
)

// Message data for intepret
type MessageRequest struct {
	irc     *IRC
	ischan  bool
	from    string
	nick    string
	user    string
	host    string
	channel string
	text    string
	direct  bool
}

func (req *MessageRequest) String() string {
	if req.ischan {
		return fmt.Sprintf("--> %s -> %s] %s",
			req.from,
			req.channel,
			req.text)
	} else {
		return fmt.Sprintf("--> %s] %s",
			req.from,
			req.text)
	}
}

type Interpreter struct {
	BaseModule

	irc *IRC

	reqCh   chan *MessageRequest
	reqExCh chan bool

	parsers  map[string]Parser
	commands map[string]Command

	nickRe *regexp.Regexp
	msgRe1 *regexp.Regexp
	msgRe2 *regexp.Regexp
	msgRe3 *regexp.Regexp

	total uint
}

func NewInterpreter(irc *IRC) *Interpreter {
	i := new(Interpreter)
	i.irc = irc
	i.Logger = irc.Logger
	i.reqCh = make(chan *MessageRequest)
	i.reqExCh = make(chan bool)

	i.commands = make(map[string]Command)
	i.RegisterCommand("VERSION", VersionCommand)
	i.RegisterCommand("SOURCE", SourceCommand)

	i.parsers = make(map[string]Parser)
	i.AddParser("URL", NewURLParser(i))

	i.nickRe = regexp.MustCompile(
		fmt.Sprintf("\\b%s\\b", irc.config.BotNick))

	// ?version
	trigger := i.irc.config.GetTrigger("")
	trigger = regexp.QuoteMeta(trigger)
	msgPtn1 := fmt.Sprintf("^%s(.*)$", trigger)
	// me: version
	msgPtn2 := fmt.Sprintf("^%s[:,;.]?(?:\\s+)?(.*)$", i.irc.config.BotNick)
	// version, me
	msgPtn3 := fmt.Sprintf("^(.*)(?:[,.:;]) %s$", i.irc.config.BotNick)

	i.msgRe1 = regexp.MustCompile(msgPtn1)
	i.msgRe2 = regexp.MustCompile(msgPtn2)
	i.msgRe3 = regexp.MustCompile(msgPtn3)
	return i
}

func (i *Interpreter) String() string {
	return "Interpreter"
}

func (i *Interpreter) Init() error {
	i.Logger.Printf("Initializing module %s", i)
	i.State = Initialized
	return nil
}

func (i *Interpreter) Start() error {
	i.Logger.Printf("Starting module %s", i)
	return nil
}

func (i *Interpreter) Stop() error {
	i.reqExCh <- true
	i.wait.Wait()
	i.State = Stopped
	return nil
}

func (i *Interpreter) Run() {
	i.wait.Add(1)
	go i.requestLoop()
	i.State = Running
}

func (i *Interpreter) Status() string {
	return fmt.Sprintf("%s", i.State)
}

// Parsers
func (i *Interpreter) AddParser(name string, parser Parser) {
	i.parsers[name] = parser
}

func (i *Interpreter) DelParser(name string) {
	delete(i.parsers, name)
}

func (i *Interpreter) ListParser() []Parser {
	return nil
}

// commands management
func (i *Interpreter) RegisterCommand(name string, cmd Command) {
	name = strings.ToUpper(name)
	i.commands[name] = cmd
}

func (i *Interpreter) DelCommand(name string) {
	name = strings.ToUpper(name)
	delete(i.commands, name)
}

func (i *Interpreter) GetCommand(name string) Command {
	name = strings.ToUpper(name)
	cmd, ok := i.commands[name]
	if ok {
		return cmd
	}
	return nil
}

// channels
func (i *Interpreter) Submit(req *MessageRequest) {
	i.reqCh <- req
}

// process request

func (i *Interpreter) requestLoop() {
	var quit bool
	var req *MessageRequest

	for !quit {
		select {
		case req = <-i.reqCh:
			i.handleRequest(req)
			break
		case quit = <-i.reqExCh:
			break
		}
	}
	i.Logger.Println("request loop exited")
	i.wait.Done()
}

func (i *Interpreter) sendReply(res string, req *MessageRequest) {
	if res == "" {
		return
	}
	if req.ischan {
		if req.direct {
			i.irc.Privmsg(req.channel, fmt.Sprintf("%s: %s", req.nick, res))
		} else {
			i.irc.Privmsg(req.channel, res)
		}
	} else {
		i.irc.Privmsg(req.nick, res)
	}
}

func (i *Interpreter) runParsers(req *MessageRequest) {
	for _, p := range i.parsers {
		err := p.Parse(i.irc.bot, req)
		if err == nil {
			return
		} else if err != ErrNotParsed {
			i.Logger.Printf("%s parse error: %s", p, err)
		}
	}
}

// handle message requests
// feed the message to parsers, if no parser was able to parse
// the request, then parse it as commands
func (i *Interpreter) handleRequest(req *MessageRequest) {
	i.Logger.Printf("%s", req)

	i.total++
	if i.nickRe.FindStringIndex(req.text) != nil {
		req.direct = true
	}

	var command string
	var text string
	var chn string
	var trigger string

	text, chn = req.text, req.channel

	trigger = i.irc.config.GetTrigger(chn)
	trigger = regexp.QuoteMeta(trigger)
	msgPtn1 := fmt.Sprintf("^%s(.*)$", trigger)
	i.msgRe1 = regexp.MustCompile(msgPtn1)

	var m []string
	m = i.msgRe1.FindStringSubmatch(text)
	if m != nil {
		command = m[1]
		goto Found
	}
	m = i.msgRe2.FindStringSubmatch(text)
	if m != nil {
		command = m[1]
		goto Found
	}
	m = i.msgRe3.FindStringSubmatch(text)
	if m != nil {
		command = m[1]
		goto Found
	}
	i.Logger.Printf("Not a command call, calling parsers")
	i.runParsers(req)
	return

Found:
	var keyword string
	var arguments string

	arr := strings.SplitN(command, " ", 2)
	keyword = arr[0]
	if len(arr) == 2 {
		arguments = arr[1]
	}

	var cmd Command

	cmd = i.GetCommand(keyword)
	if cmd != nil {
		i.Logger.Printf("calling %s with [%s]", keyword, arguments)
		res, err := cmd(i.irc, arguments)
		if err == nil {
			i.sendReply(res, req)
		} else {
			i.Logger.Printf("%s error: %s", keyword, err)
		}
	} else {
		i.Logger.Printf("command not handled, calling parsers")
		i.runParsers(req)
	}
	return
}
