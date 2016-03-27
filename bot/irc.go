// Copyright 2016 Alex Fluter

package bot

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const cr byte = '\r'
const lf byte = '\n'
const crlf string = "\r\n"

const PING_INTERVAL = 5 * time.Second

var (
	msgRe     *regexp.Regexp
	nickRe    *regexp.Regexp
	rpl_002Re *regexp.Regexp
)

const (
	msgPtn     = "^(:[^ \000]+ )?([[:alpha:]]+|[[:digit:]]{3}) (.*)$"
	tgtPtn     = "^([A-Za-z\\[\\]\\\\`_\\^{|}][A-Za-z0-9\\[\\]\\\\`_\\^{|}-]{0,15})!([^ @\000]+)@([a-zA-Z0-9:/.-]*)$"
	rpl_002Ptn = "Your host is ([a-zA-Z0-9.-]*)(\\[[0-9./]*\\])?, running version (.*)"
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

type ServerInfo struct {
	host    string
	version string

	supports []string
}

type IRC struct {
	BaseModule

	bot       *Bot
	config    *IRCConfig
	rawLogger *log.Logger

	host string
	mode string

	server ServerInfo

	state ModState

	cx chan bool

	cMsg  chan string
	cCmd  chan IRCCommand
	cxMsg chan bool
	cxCmd chan bool

	timer   *time.Ticker
	cxTimer chan bool

	conn net.Conn

	handlers map[string]CommandHandler

	channels map[string]*Channel

	interpreter Interpreter
}

func NewIRC(bot *Bot, config *IRCConfig) *IRC {
	var irc *IRC
	irc = new(IRC)
	irc.bot = bot
	irc.config = config
	irc.Logger = NewLoggerFunc(bot.Name)
	if irc.bot.config.RawLogging {
		irc.rawLogger = NewLoggerFunc(bot.Name + "-raw")
		irc.rawLogger.SetFlags(0)
	} else {
		irc.rawLogger = NewLoggerFunc(bot.Name)
	}
	irc.state = Disconnected
	irc.cx = make(chan bool)
	irc.cMsg = make(chan string)
	irc.cCmd = make(chan IRCCommand)
	irc.cxMsg = make(chan bool)
	irc.cxCmd = make(chan bool)
	irc.timer = nil
	irc.cxTimer = make(chan bool)
	irc.channels = make(map[string]*Channel)
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
	return fmt.Sprintf("%s: %s %s", "IRC", irc.host, irc.mode)
}

// Module interface

func (irc *IRC) Init() error {
	irc.Logger.Printf("Initializing module %s", irc)
	irc.state = Disconnected
	return nil
}

func (irc *IRC) Start() error {
	irc.Logger.Printf("Starting module %s", irc)

	if irc.config.AutoConnect {
		return irc.connect()
	}

	return nil
}

func (irc *IRC) Loop() {
	go irc.messageLoop()
	irc.commandLoop()
}

func (irc *IRC) Run() {
	go irc.Loop()
	irc.state = Running
}

func (irc *IRC) Status() string {
	return fmt.Sprintf("Server: %s %s\nState: %s\nChannels(%d): %s",
		irc.server.host, irc.server.version,
		irc.state, len(irc.channels), irc.channels)
}

func (irc *IRC) Stop() error {
	if irc.state >= Connected {
		irc.Quit("Exiting...")
	}
	irc.cxMsg <- true
	irc.cxCmd <- true
	irc.Logger.Println("Message loop finished")
	irc.Logger.Println("Command loop finished")

	close(irc.cxMsg)
	close(irc.cxCmd)
	close(irc.cMsg)
	close(irc.cCmd)
	return nil
}

func (irc *IRC) RawLogger() *log.Logger {
	return irc.rawLogger
}

// end module interface

// IRC control methods
func (irc *IRC) disconnect() {
	if irc.state > Disconnected && irc.conn != nil {
		irc.conn.Close()
		irc.timer.Stop()
		irc.cxTimer <- true
		for ch := range irc.channels {
			irc.LeaveChannel(ch)
		}
		irc.conn = nil
		irc.state = Disconnected
		irc.Logger.Println("IRC disconnected")
	} else {
		irc.Logger.Println("IRC already disconnected")
	}
}

func (irc *IRC) connect() error {
	var err error
	var addr string
	var raddr *net.TCPAddr
	var tcpConn *net.TCPConn

	addr = fmt.Sprintf("%s:%d",
		irc.config.Server,
		irc.config.Port)
	irc.Logger.Printf("Connecting to IRC server %s", addr)
	raddr, err = net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		irc.Logger.Printf("Failed to resolve irc server %s", addr)
		irc.Logger.Println(err)
		return err
	}

	tcpConn, err = net.DialTCP("tcp", nil, raddr)
	if err != nil {
		irc.Logger.Printf("Failed to connect to irc server: %s", err)
		return err
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
			return err
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
	irc.state = Connected
	irc.Logger.Println("IRC connected")

	err = irc.register()
	if err != nil {
		irc.Logger.Println("Failed to register to server")
		return err
	}
	irc.state = Identified
	irc.Logger.Println("IRC regiested")

	err = irc.joinChannels()
	if err != nil {
		irc.Logger.Println("Failed to join pre-configured channels")
		return err
	}

	irc.timer = time.NewTicker(PING_INTERVAL)
	irc.Logger.Printf("IRC timer started")

	go irc.runTimer()
	go irc.readLoop()
	return nil
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
	irc.config.Trigger = '?'
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

	if irc.state < Connected {
		return
	}

	msg = make([]byte, 5120)
	frag = make([]byte, 512)
	fraglen = 0

	for {
		n, err = irc.conn.Read(msg)
		t = time.Now()
		irc.RawLogger().Printf("%s\t%s\t%s\t%s",
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
						irc.cMsg <- string(long)
						fraglen = 0
					} else {
						irc.cMsg <- string(msg[start:i])
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
			irc.disconnect()
			// network caused disconnect, reconnecting
			defer func() {
				irc.Logger.Print("Reconnecting to server")
				irc.connect()
			}()
			break
		}
	}
	irc.Logger.Print("IRC read loop done")
}

func (irc *IRC) runTimer() {
	var stop bool

	for !stop {
		select {
		case <-irc.timer.C:
			//irc.Logger.Printf("::::tick %s", t)
			//irc.Ping(irc.server.host)
		case stop = <-irc.cxTimer:
			break
		}
	}
	irc.Logger.Printf("IRC timer stopped")
}

func (irc *IRC) messageLoop() {
	var data string
	var quit bool
	for !quit {
		select {
		case data = <-irc.cMsg:
			irc.onMessage(data)
		case quit = <-irc.cxMsg:
			break
		}
	}
	irc.Logger.Print("IRC message loop done")
}

func (irc *IRC) commandLoop() {
	var cmd IRCCommand
	var quit bool
	for !quit {
		select {
		case cmd = <-irc.cCmd:
			irc.onCommand(cmd.cmd, cmd.prefix, cmd.param)
		case quit = <-irc.cxCmd:
			break
		}
	}
	irc.Logger.Print("IRC command loop done")
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
			irc.disconnect()
			return err
		}

		sent += n
	}
	t := time.Now()
	irc.RawLogger().Printf("%s\t%s\t%s\t%s",
		dateTime(t),
		irc.config.Name,
		OUT,
		msg)
	return nil
}

// end IRC internal communications

// bot command to irc, triggered by bot event
func (irc *IRC) handleCommand(cmd string) {
	irc.Logger.Println("IRC command", cmd)
	cmd = strings.TrimSpace(cmd)
	args := strings.Split(cmd, " ")

	switch args[0] {
	default:
		irc.sendMsg(cmd)
	case "ctcp", "CTCP":
		var target, typ, param string
		if len(args) > 1 {
			target = args[1]
		}
		if len(args) > 2 {
			typ = args[2]
		}
		if len(args) > 3 {
			param = strings.Join(args[3:], "")
		}
		irc.Ctcp(target, typ, param)
	}
}

// IRC message handling
func (irc *IRC) canHandleCommand(cmd string) bool {
	var ok bool
	_, ok = irc.handlers[cmd]
	return ok
}

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
	if irc.canHandleCommand(cmd) {
		irc.cCmd <- IRCCommand{cmd, prefix, param}
	} else {
		fmt.Printf("%s|%s|%s\n", cmd, prefix, param)
		irc.Logger.Printf("unhandled message %s", msg)
	}
}

func (irc *IRC) AddHandler(cmd string, h CommandHandler) {
}

func (irc *IRC) RemoveHandler(cmd string) {
}

func (irc *IRC) onCommand(cmd, prefix, param string) {
	//	irc.Logger.Printf("handling command %s", cmd)
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
