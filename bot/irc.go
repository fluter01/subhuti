// Copyright 2016 Alex Fluter

package bot

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"time"
)

const (
	Ping_interval        = 1 * time.Minute
	connect_wait         = 5 * time.Second
	cr            byte   = '\r'
	lf            byte   = '\n'
	crlf          string = "\r\n"

	msgPtn     = "^(:[^ \000]+ )?([[:alpha:]]+|[[:digit:]]{3}) (.*)$"
	tgtPtn     = "^([A-Za-z\\[\\]\\\\`_\\^{|}][A-Za-z0-9\\[\\]\\\\`_\\^{|}-]{0,15})!([^ @\000]+)@([a-zA-Z0-9:/.-]*)$"
	rpl_002Ptn = "Your host is ([a-zA-Z0-9.-]*)(\\[[0-9./]*\\])?, running version (.*)"
)

var (
	msgRe     *regexp.Regexp
	nickRe    *regexp.Regexp
	rpl_002Re *regexp.Regexp
)

func init() {
	msgRe = regexp.MustCompile(msgPtn)
	nickRe = regexp.MustCompile(tgtPtn)
	rpl_002Re = regexp.MustCompile(rpl_002Ptn)
}

type IRCCommand struct {
	cmd, prefix, param string
}

// IRC command handler, args: prefix, params
type CommandHandler func(string, string)

type IRC struct {
	BaseModule

	bot       *Bot
	config    *IRCConfig
	rawLogger *log.Logger

	stopping bool
	conn     net.Conn

	host     string
	version  string
	mode     string
	supports []string
	cloak    string

	lag time.Duration

	msgCh     chan string
	cmdCh     chan *IRCCommand
	msgExCh   chan bool
	cmdExCh   chan bool
	timer     *time.Ticker
	timerExCh chan bool

	handlers    map[string]CommandHandler
	channels    map[string]*Channel
	interpreter *Interpreter
}

func NewIRC(bot *Bot, config *IRCConfig) *IRC {
	var irc *IRC
	irc = new(IRC)
	irc.bot = bot
	irc.Name = config.Name
	irc.config = config
	irc.Logger = NewLoggerFunc(fmt.Sprintf("%s/%s-%s",
		bot.config.LogDir, bot.Name, config.Name))
	if config.RawLogging {
		irc.rawLogger = NewLoggerFunc(fmt.Sprintf("%s/%s-%s-raw",
			bot.config.LogDir, bot.Name, config.Name))
		irc.rawLogger.SetFlags(0)
	} else {
		irc.rawLogger = NewLoggerFunc("")
	}
	irc.State = Disconnected
	irc.exitCh = make(chan bool)
	irc.msgCh = make(chan string)
	irc.cmdCh = make(chan *IRCCommand)
	irc.msgExCh = make(chan bool)
	irc.cmdExCh = make(chan bool)
	irc.timer = nil
	irc.timerExCh = make(chan bool)
	irc.channels = make(map[string]*Channel)
	irc.interpreter = NewInterpreter(irc)
	// IRC internal handlers, plugins should use Events to register
	irc.handlers = map[string]CommandHandler{
		"PING":    irc.onPing,
		"PONG":    irc.onPong,
		"PRIVMSG": irc.onPrivmsg,
		"JOIN":    irc.onJoin,
		"PART":    irc.onPart,
		"QUIT":    irc.onQuit,
		"NICK":    irc.onNick,
		"INVITE":  irc.onInvite,
		"NOTICE":  irc.onNotice,
		"MODE":    irc.onMode,
		"ERROR":   irc.onError,

		"RPL_WELCOME":  irc.onRPL_WELCOME,
		"RPL_YOURHOST": irc.onRPL_YOURHOST,
		"RPL_CREATED":  irc.onRPL_CREATED,
		"RPL_MYINFO":   irc.onRPL_MYINFO,
		"RPL_ISUPPORT": irc.onRPL_ISUPPORT,

		"RPL_STATSCONN":     irc.onRPL_STATSCONN,
		"RPL_LUSERCLIENT":   irc.onRPL_LUSERCLIENT,
		"RPL_LUSEROP":       irc.onRPL_LUSEROP,
		"RPL_LUSERUNKNOWN":  irc.onRPL_LUSERUNKNOWN,
		"RPL_LUSERCHANNELS": irc.onRPL_LUSERCHANNELS,
		"RPL_LUSERME":       irc.onRPL_LUSERME,

		"RPL_LOCALUSERS":  irc.onRPL_LOCALUSERS,
		"RPL_GLOBALUSERS": irc.onRPL_GLOBALUSERS,

		"RPL_WHOISUSER":     irc.onRPL_WHOISUSER,
		"RPL_WHOISSERVER":   irc.onRPL_WHOISSERVER,
		"RPL_WHOISOPERATOR": irc.onRPL_WHOISOPERATOR,
		"RPL_WHOWASUSER":    irc.onRPL_WHOWASUSER,
		"RPL_ENDOFWHO":      irc.onRPL_ENDOFWHO,

		"RPL_WHOISIDLE":     irc.onRPL_WHOISIDLE,
		"RPL_ENDOFWHOIS":    irc.onRPL_ENDOFWHOIS,
		"RPL_WHOISCHANNELS": irc.onRPL_WHOISCHANNELS,
		"RPL_WHOISSPECIAL":  irc.onRPL_WHOISSPECIAL,

		"RPL_CHANNELURL":    irc.onRPL_CHANNELURL,
		"RPL_CREATIONTIME":  irc.onRPL_CREATIONTIME,
		"RPL_WHOISLOGGEDIN": irc.onRPL_WHOISLOGGEDIN,
		"RPL_NOTOPIC":       irc.onRPL_NOTOPIC,
		"RPL_TOPIC":         irc.onRPL_TOPIC,
		"RPL_TOPICWHOTIME":  irc.onRPL_TOPICWHOTIME,

		"RPL_NAMREPLY":   irc.onRPL_NAMREPLY,
		"RPL_ENDOFNAMES": irc.onRPL_ENDOFNAMES,

		"RPL_MOTD":      irc.onRPL_MOTD,
		"RPL_MOTDSTART": irc.onRPL_MOTDSTART,
		"RPL_ENDOFMOTD": irc.onRPL_ENDOFMOTD,

		"RPL_WHOISHOST": irc.onRPL_WHOISHOST,

		"RPL_HOSTHIDDEN": irc.onRPL_HOSTHIDDEN,

		"ERR_INVITEONLYCHAN": irc.onERR_INVITEONLYCHAN,

		"RPL_WHOISSECURE": irc.onRPL_WHOISSECURE,
	}
	return irc
}

func (irc *IRC) String() string {
	return fmt.Sprintf("%s(%s)", "IRC", irc.Name)
}

// Module interface
func (irc *IRC) Init() error {
	irc.Logger.Printf("Initializing module %s", irc)
	err := irc.interpreter.Init()
	irc.State = Disconnected
	return err
}

func (irc *IRC) Status() string {
	if irc.conn != nil {
		return fmt.Sprintf("Connected to: %s %s as %s@%s\n"+
			"State: %s\nChannels(%d): %s",
			irc.host, irc.version, irc.config.BotNick, irc.cloak,
			irc.State, len(irc.channels), irc.channels)
	} else {
		return fmt.Sprintf("Not connected, State: %s",
			irc.State)
	}
}

func (irc *IRC) Start() error {
	irc.Logger.Printf("Starting module %s", irc)
	err := irc.interpreter.Start()
	return err
}

func (irc *IRC) Run() {
	irc.wait.Add(2)
	go irc.messageLoop()
	go irc.commandLoop()
	irc.interpreter.Run()

	if irc.config.AutoConnect {
		if irc.connect() != nil {
			return
		}
	}
	irc.State = Running
}

func (irc *IRC) Stop() error {
	irc.stopping = true
	if irc.conn != nil {
		irc.Quit("Exiting...")
		irc.disconnect()
	}
	irc.interpreter.Stop()
	irc.msgExCh <- true
	irc.cmdExCh <- true

	irc.wait.Wait()
	close(irc.msgExCh)
	close(irc.cmdExCh)
	close(irc.msgCh)
	close(irc.cmdCh)
	return nil
}

// IRC control methods
func (irc *IRC) connect() error {
	var err error
	var addr string
	var raddr *net.TCPAddr
	var tcpConn *net.TCPConn

	for {
		addr = fmt.Sprintf("%s:%d",
			irc.config.Server,
			irc.config.Port)
		irc.Logger.Printf("Connecting to IRC server %s", addr)
		raddr, err = net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			irc.Logger.Printf("Failed to resolve irc server %s", addr)
			irc.Logger.Println(err)
			goto fail
		}

		tcpConn, err = net.DialTCP("tcp", nil, raddr)
		if err != nil {
			irc.Logger.Printf("Failed to connect to irc server: %s", err)
			goto fail
		}
		irc.conn = tcpConn
		if irc.config.Ssl {
			irc.Logger.Println("Connecting using tls")
			var tlsConn *tls.Conn
			var tlsConfig tls.Config

			// tlsConfig.InsecureSkipVerify = true
			tlsConfig.ServerName = "freenode.net"
			tlsConn = tls.Client(tcpConn, &tlsConfig)
			//  conn, err = tls.Dial("tcp", addr, &tlsConfig)
			err = tlsConn.Handshake()
			if err != nil {
				irc.Logger.Printf("Failed to run tls handshake: %s", err)
				goto fail
			}
			// fmt.Println(tlsConn)
			state := tlsConn.ConnectionState()
			irc.Logger.Printf("Version %X Cipher %X",
				state.Version,
				state.CipherSuite)
			irc.Logger.Printf("receiving %d certificates",
				len(state.PeerCertificates))
			for i, cert := range state.PeerCertificates {
				irc.Logger.Printf("certificate %d: %s", i, dumpCert(cert))
			}
			// fmt.Println("ct", tlsConn.ConnectionState().PeerCertificates)
			irc.conn = tlsConn
		}

		irc.Logger.Printf("Connected %s <--> %s",
			irc.conn.LocalAddr(), irc.conn.RemoteAddr())
		irc.State = Connected
		irc.Logger.Println("IRC connected")

		err = irc.register()
		if err != nil {
			irc.Logger.Println("Failed to register to server", err)
			goto fail
		}
		irc.State = Identified
		irc.Logger.Println("IRC regiested")

		err = irc.joinChannels()
		if err != nil {
			irc.Logger.Println("Failed to join pre-configured channels", err)
			goto fail
		}
		break
	fail:
		time.Sleep(connect_wait)
	}

	irc.timer = time.NewTicker(Ping_interval)
	irc.Logger.Printf("IRC timer started")

	irc.wait.Add(2)
	go irc.runTimer()
	go irc.readLoop()
	return nil
}

func (irc *IRC) disconnect() {
	if irc.conn != nil {
		irc.conn.Close()
		irc.timer.Stop()
		irc.timerExCh <- true
		for ch := range irc.channels {
			irc.LeaveChannel(ch)
		}
		irc.conn = nil
		irc.State = Disconnected
		irc.Logger.Println("IRC disconnected")
	} else {
		irc.Logger.Println("IRC already disconnected")
	}
}

// end IRC control methods

// IRC internal communications
func (irc *IRC) register() error {
	var err error
	err = irc.sendMsg("PASS " + irc.config.Identify_passwd)
	if err != nil {
		return err
	}
	err = irc.sendMsg("NICK " + irc.config.BotNick)
	if err != nil {
		return err
	}
	err = irc.sendMsg(fmt.Sprintf("USER %s %d * :%s",
		irc.config.Username, 8,
		irc.config.RealName))
	if err != nil {
		return err
	}
	return nil
}

func (irc *IRC) joinChannels() error {
	var err error
	for _, ch := range irc.config.Channels {
		err = irc.Join(ch.Name)
		if err != nil {
			irc.Logger.Printf("Failed to join %s: %s",
				ch.Name, err)
			return err
		} else {
			irc.Logger.Printf("Joined %s",
				ch.Name)
		}
	}
	return nil
}

func (irc *IRC) readLoop() {
	var err error
	var msg []byte
	var frag []byte
	var fraglen int
	var n int
	var i int
	var start int
	var t time.Time

	if irc.State < Connected {
		return
	}

	msg = make([]byte, 5120)
	frag = make([]byte, 512)
	fraglen = 0

	for {
		n, err = irc.conn.Read(msg)
		t = time.Now()
		irc.rawLogger.Printf("%s\t%s\t%s\t%s",
			dateTime(t),
			irc.config.Name,
			IN,
			string(msg[:n]))
		if n > 0 {
			start = 0
			for i = 0; i < n-1; i++ {
				if msg[i] == cr && msg[i+1] == lf {
					if fraglen > 0 {
						var long []byte
						long = make([]byte, fraglen+i-start)
						copy(long, frag[:fraglen])
						copy(long[fraglen:], msg[start:i])
						irc.msgCh <- string(long)
						fraglen = 0
					} else {
						irc.msgCh <- string(msg[start:i])
					}
					start = i + 2
					i += 1
				}
			}
			if start < n {
				copy(frag[fraglen:], msg[start:])
				fraglen += n - start
			}
		}
		if err != nil {
			irc.Logger.Print("Read error:", err)
			if !irc.stopping {
				irc.bot.AddEvent(NewEvent(Disconnect, irc))
				if irc.config.AutoConnect {
					defer irc.reconnect()
				}
			}
			break
		}
	}
	irc.wait.Done()
	irc.Logger.Print("IRC read loop done")
}

func (irc *IRC) reconnect() {
	irc.conn = nil
	irc.State = Disconnected
	irc.connect()
}

func (irc *IRC) sendMsg(msg string) error {
	var err error
	var n int
	var total int
	var sent int
	var cmd string
	var data []byte

	cmd = fmt.Sprintf("%s%s", msg, crlf)
	data = []byte(cmd)
	total = len(data)
	sent = 0

	for sent < total {
		n, err = irc.conn.Write(data[sent:])
		if err != nil {
			irc.Logger.Printf("Failed to send message: %s", msg)
			irc.Logger.Println(err)
			irc.bot.AddEvent(NewEvent(Disconnect, irc))
			if irc.config.AutoConnect {
				defer irc.reconnect()
			}
			return err
		}

		sent += n
	}
	t := time.Now()
	irc.rawLogger.Printf("%s\t%s\t%s\t%s",
		dateTime(t),
		irc.config.Name,
		OUT,
		msg)
	return nil
}

func (irc *IRC) runTimer() {
	var stop bool

	for !stop {
		select {
		case <-irc.timer.C:
			irc.Ping(irc.host)
		case stop = <-irc.timerExCh:
			break
		}
	}
	irc.wait.Done()
	irc.Logger.Printf("IRC timer stopped")
}

func (irc *IRC) messageLoop() {
	var data string
	var quit bool
	for !quit {
		select {
		case data = <-irc.msgCh:
			irc.onMessage(data)
		case quit = <-irc.msgExCh:
			break
		}
	}
	irc.wait.Done()
	irc.Logger.Print("IRC message loop exited")
}

func (irc *IRC) commandLoop() {
	var cmd *IRCCommand
	var quit bool
	for !quit {
		select {
		case cmd = <-irc.cmdCh:
			irc.onCommand(cmd.cmd, cmd.prefix, cmd.param)
		case quit = <-irc.cmdExCh:
			break
		}
	}
	irc.wait.Done()
	irc.Logger.Print("IRC command loop exited")
}

// end IRC internal communications

// bot command to irc, triggered by bot event
//func (irc *IRC) handleCommand(cmd string) {
//	irc.Logger.Println("IRC command", cmd)
//	cmd = strings.TrimSpace(cmd)
//	args := strings.Split(cmd, " ")
//
//	switch args[0] {
//	default:
//		irc.sendMsg(cmd)
//	case "ctcp", "CTCP":
//		var target, typ, param string
//		if len(args) > 1 {
//			target = args[1]
//		}
//		if len(args) > 2 {
//			typ = args[2]
//		}
//		if len(args) > 3 {
//			param = strings.Join(args[3:], "")
//		}
//		irc.Ctcp(target, typ, param)
//	}
//}

// IRC message handling
func (irc *IRC) onMessage(msg string) {
	var m []string

	m = msgRe.FindStringSubmatch(msg)
	if m == nil {
		irc.Logger.Printf("Funky message found: %s", msg)
		return
	}
	prefix, cmd, param := m[1], m[2], m[3]
	// trim tailing space
	if prefix != "" {
		prefix = prefix[1 : len(prefix)-1]
	}
	if numeric(cmd) {
		n, err := strconv.Atoi(cmd)
		if err != nil {
			panic(err)
		}
		cmd = Numerics[n]
	}
	if _, ok := irc.handlers[cmd]; ok {
		irc.cmdCh <- &IRCCommand{cmd, prefix, param}
	} else {
		fmt.Printf("%s|%s|%s\n", cmd, prefix, param)
		irc.Logger.Printf("unhandled message %s", msg)
	}
}

func (irc *IRC) onCommand(cmd, prefix, param string) {
	var proc CommandHandler

	proc = irc.handlers[cmd]
	proc(prefix, param)
}

// end IRC message handling

// Channel management
func (irc *IRC) GetChannel(ch string) *Channel {
	var channel *Channel
	var ok bool
	if channel, ok = irc.channels[ch]; !ok {
		return nil
	}
	return channel
}

func (irc *IRC) JoinChannel(ch string) *Channel {
	var channel *Channel

	if _, ok := irc.channels[ch]; ok {
		panic(ch)
	}

	channel = NewChannel(irc, ch)
	irc.channels[ch] = channel
	return channel
}

func (irc *IRC) LeaveChannel(ch string) *Channel {
	var channel *Channel
	if ch, ok := irc.channels[ch]; !ok {
		panic(ch)
	}
	delete(irc.channels, ch)
	return channel
}

// end channel management
