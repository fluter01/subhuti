// Copyright 2016 Alex Fluter

package bot

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
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

	// computed in the interpreter
	direct bool // directed to me
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

func NewMessageRequest(irc *IRC,
	ischan bool,
	from, nick, user, host string,
	channel, text string) *MessageRequest {
	req := &MessageRequest{
		irc:     irc,
		ischan:  ischan,
		from:    from,
		nick:    nick,
		user:    user,
		host:    host,
		channel: channel,
		text:    text}
	return req
}

// response data
type MessageResponse struct {
	req  *MessageRequest
	text string
}

func (resp *MessageResponse) String() string {
	return fmt.Sprintf("--> %s] %s",
		resp.req.from, resp.text)
}

// Parsers
var NotParsed = errors.New("Not parsed")

type Parser interface {
	Parse(*MessageRequest) (string, error)
}

// irc message command
type run func(*MessageRequest, string) (string, error)

type Interpreter struct {
	bot   *Bot
	state ModState

	cReq  chan *MessageRequest
	cRsp  chan *MessageResponse
	cxReq chan bool
	cxRsp chan bool

	wg *sync.WaitGroup

	parsers  map[string]Parser
	commands map[string]run

	nickRe *regexp.Regexp
}

func NewInterpreter(bot *Bot) *Interpreter {
	i := new(Interpreter)
	i.bot = bot
	i.wg = new(sync.WaitGroup)
	i.cReq = make(chan *MessageRequest)
	i.cRsp = make(chan *MessageResponse)
	i.cxReq = make(chan bool)
	i.cxRsp = make(chan bool)

	i.commands = make(map[string]run)

	i.Register("VERSION", HandleVersion)
	i.Register("SOURCE", HandleSource)

	i.parsers = make(map[string]Parser)
	i.AddParser("URL", NewURLParser(i))

	i.nickRe = regexp.MustCompile(
		fmt.Sprintf("\\b%s\\b", bot.config.BotNick))
	return i
}

func (i *Interpreter) String() string {
	return "Interpreter"
}

func (i *Interpreter) Init() error {
	i.Logger().Printf("Initializing module %s", i)
	i.state = Initialized
	return nil
}

func (i *Interpreter) Start() error {
	i.Logger().Printf("Starting module %s", i)
	i.state = Running
	return nil
}

func (i *Interpreter) Stop() error {
	i.cxReq <- true
	i.cxRsp <- true
	return nil
}

func (i *Interpreter) Loop() {
	i.wg.Add(1)
	go i.requestLoop()
	i.wg.Add(2)
	go i.responseLoop()

	i.wg.Wait()
}

func (i *Interpreter) Run() {
	go i.Loop()
}

func (i *Interpreter) Status() string {
	return fmt.Sprintf("%s", i.state)
}

func (i *Interpreter) Logger() Logger {
	return i.bot.Logger()
}

// Parsers
func (i *Interpreter) AddParser(name string, parser Parser) {
	i.parsers[name] = parser
}

func (i *Interpreter) RemoveParser(name string) {
	delete(i.parsers, name)
}

func (i *Interpreter) ListParser() []Parser {
	return nil
}

// commands management
func (i *Interpreter) Register(command string, proc run) {
	i.commands[command] = proc
}

func (i *Interpreter) GetCommand(command string) run {
	command = strings.ToUpper(command)
	proc, ok := i.commands[command]
	if ok {
		return proc
	}
	return nil
}

// channels
func (i *Interpreter) RequestChan() chan *MessageRequest {
	return i.cReq
}

// process request

func (i *Interpreter) requestLoop() {
	var quit bool
	var req *MessageRequest

	for !quit {
		select {
		case req = <-i.cReq:
			i.handleRequest(req)
			break
		case quit = <-i.cxReq:
			break
		}
	}
	i.wg.Done()
}

func (i *Interpreter) responseLoop() {
	var quit bool
	var resp *MessageResponse

	for !quit {
		select {
		case resp = <-i.cRsp:
			i.handleResponse(resp)
			break
		case quit = <-i.cxRsp:
			break
		}
	}
	i.wg.Done()
}

func (i *Interpreter) handleRequest(req *MessageRequest) {
	i.Logger().Printf("%s", req)

	var result string
	var err error

	if i.nickRe.FindStringIndex(req.text) != nil {
		req.direct = true
	}

	for n, p := range i.parsers {
		result, err = p.Parse(req)
		if err == nil {
			i.Logger().Printf("Handled by %s", n)
			i.cRsp <- &MessageResponse{req, result}
			return
		}
	}

	var command string
	var text string
	var chn string
	var trigger string

	text, chn = req.text, req.channel

	trigger = i.bot.config.Trigger(chn)
	trigger = regexp.QuoteMeta(trigger)
	// ?version
	msgPtn1 := fmt.Sprintf("^%s(.*)$", trigger)
	// me: version
	msgPtn2 := fmt.Sprintf("^%s[:,;.]?(?:\\s+)?(.*)$", i.bot.config.BotNick)
	// version, me
	msgPtn3 := fmt.Sprintf("^(.*)(?:[,.:;]) %s$", i.bot.config.BotNick)
	// fmt.Println(msgPtn1)
	// fmt.Println(msgPtn2)
	// fmt.Println(msgPtn3)

	msgRe1 := regexp.MustCompile(msgPtn1)
	msgRe2 := regexp.MustCompile(msgPtn2)
	msgRe3 := regexp.MustCompile(msgPtn3)

	var m []string
	m = msgRe1.FindStringSubmatch(text)
	if m != nil {
		command = m[1]
		goto Found
	}
	m = msgRe2.FindStringSubmatch(text)
	if m != nil {
		command = m[1]
		goto Found
	}
	m = msgRe3.FindStringSubmatch(text)
	if m != nil {
		command = m[1]
		goto Found
	}
	fmt.Println("no match")
	return

Found:
	fmt.Println(m)
	fmt.Printf("command is %s\n", command)

	var keyword string
	var arguments string

	arr := strings.SplitN(command, " ", 2)
	keyword = arr[0]
	if len(arr) == 2 {
		arguments = arr[1]
	}

	var cmd run

	cmd = i.GetCommand(keyword)
	fmt.Println("cmd is", cmd)
	if cmd != nil {
		result, err = cmd(req, arguments)
		if err != nil {
			i.Logger().Printf("%s error: %s", keyword, err)
		}
	} else {
		result = fmt.Sprintf("Unknown command: %s", keyword)
	}
	i.cRsp <- &MessageResponse{req, result}

	return
}

func (i *Interpreter) handleResponse(resp *MessageResponse) {
	i.Logger().Printf("%s", resp)
	if resp.text == "" {
		return
	}
	if resp.req.ischan {
		i.bot.IRC().Privmsg(resp.req.channel, resp.text)
	} else {
		i.bot.IRC().Privmsg(resp.req.nick, resp.text)
	}
}
