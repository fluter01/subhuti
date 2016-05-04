// Copyright 2016 Alex Fluter

package bot

import (
	"io"
	"net"
	"testing"
)

func matchNickRe(s string, t *testing.T) {
	var m []string
	m = nickRe.FindStringSubmatch(s)
	if m == nil {
		t.Fail()
	}
	t.Log(m)
}

func TestNickRe(t *testing.T) {
	var s string

	s = "fe!~fe@servx.ru"
	matchNickRe(s, t)

	s = "jayeshsolanki!~jayeshsol@219.91.250.106"
	matchNickRe(s, t)

	s = "j!~jayeshsol@219.91.250.106"
	matchNickRe(s, t)

	s = "de-facto_!~de-facto@unaffiliated/de-facto"
	matchNickRe(s, t)

	s = "c!~c@freenode"
	matchNickRe(s, t)
}

func newTestBot(ch chan bool) *Bot {
	config := &BotConfig{
		Trigger: '/',
		DataDir: "../data",
		DB:      "testbot",
		IRC: []*IRCConfig{
			&IRCConfig{
				Name:        "Localhost",
				Server:      "127.0.0.1",
				AutoConnect: false,
				Trigger:     '?',
				BotNick:     G,
				Channels: []*ChannelConfig{
					&ChannelConfig{
						Name:    "#candice",
						Repaste: true,
					},
				},
			},
		},
		CompileServer: "127.0.0.1:1234",
	}
	bot := NewBot("TestBot", config)
	go func() {
		bot.Start()
		ch <- true
	}()
	for bot.State != Running {
	}
	return bot
}

func delTestBot(bot *Bot, t *testing.T, ch chan bool) {
	bot.Stop()
	<-ch
	if bot.State != Stopped {
		t.Fail()
	}
}

func readLog(r io.Reader, t *testing.T) {
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	t.Log(string(buf[:n]))
}

var (
	pc *int
	c  chan bool
)

func bashCmd(*MessageRequest, string) (string, error) {
	(*pc)++
	c <- true
	return "", nil
}

func TestIRCCommandChannel(t *testing.T) {
	ch := make(chan bool)
	bot := newTestBot(ch)
	if bot.State != Running {
		t.Fail()
	}

	var irc *IRC
	for _, mod := range bot.modules {
		if _, ok := mod.(*IRC); ok {
			irc = mod.(*IRC)
			break
		}
	}
	if irc == nil {
		t.Fail()
	}

	var counter int
	c = make(chan bool)
	pc = &counter
	irc.interpreter.RegisterCommand("bash", bashCmd)

	irc.onCommand("PRIVMSG", "", "#candice :hello")
	// channel messages
	irc.onCommand("PRIVMSG", "", "#candice :bash")
	// call with trigger
	irc.onCommand("PRIVMSG", "", "#candice :?bash")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :!bash")
	irc.onCommand("PRIVMSG", "", "#candice :?bash arg1 arg2")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :,bash arg1 arg2")
	// call with nick, 1st form
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti: bash")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti, bash")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti; bash")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti. bash")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti! bash")
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti: bash arg1 arg2")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti, bash arg1 arg2")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti; bash arg1 arg2")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti. bash arg1 arg2")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :Subhuti! bash arg1 arg2")
	// call with nick, 2nd form
	irc.onCommand("PRIVMSG", "", "#candice :bash, Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash. Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash: Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash; Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash! Subhuti")
	irc.onCommand("PRIVMSG", "", "#candice :bash arg1 arg2, Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash arg1 arg2. Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash arg1 arg2: Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash arg1 arg2; Subhuti")
	<-c
	irc.onCommand("PRIVMSG", "", "#candice :bash arg1 arg2! Subhuti")

	if counter != 18 {
		t.Fail()
	}

	for irc.interpreter.total != 26 {
	}

	delTestBot(bot, t, ch)
}

func TestIRCCommandPrivate(t *testing.T) {
	ch := make(chan bool)
	bot := newTestBot(ch)
	if bot.State != Running {
		t.Fail()
	}

	var irc *IRC
	for _, mod := range bot.modules {
		if _, ok := mod.(*IRC); ok {
			irc = mod.(*IRC)
			break
		}
	}
	if irc == nil {
		t.Fail()
	}

	var counter int
	c = make(chan bool)
	pc = &counter
	irc.interpreter.RegisterCommand("bash", bashCmd)

	irc.onCommand("PRIVMSG", "", "foo :hello")
	// channel messages
	irc.onCommand("PRIVMSG", "", "foo :bash")
	<-c
	// call with trigger
	irc.onCommand("PRIVMSG", "", "foo :?bash")
	<-c
	irc.onCommand("PRIVMSG", "", "foo :!bash")
	irc.onCommand("PRIVMSG", "", "foo :?bash arg1 arg2")
	<-c
	irc.onCommand("PRIVMSG", "", "foo :,bash arg1 arg2")

	if counter != 3 {
		t.Fail()
	}

	for irc.interpreter.total != 6 {
	}

	delTestBot(bot, t, ch)
}

func TestIRCURLParserGetTitle(t *testing.T) {
	ch := make(chan bool)
	bot := newTestBot(ch)
	if bot.State != Running {
		t.Fail()
	}

	var irc *IRC
	for _, mod := range bot.modules {
		if _, ok := mod.(*IRC); ok {
			irc = mod.(*IRC)
			break
		}
	}
	if irc == nil {
		t.Fail()
	}

	r, w := net.Pipe()
	irc.conn = w

	irc.onCommand("PRIVMSG", "foo", "#candice :https://www.bing.com")
	readLog(r, t)

	irc.conn = nil
	delTestBot(bot, t, ch)
}

func TestIRCURLParserPaste(t *testing.T) {
	ch := make(chan bool)
	bot := newTestBot(ch)
	if bot.State != Running {
		t.Fail()
	}

	var irc *IRC
	for _, mod := range bot.modules {
		if _, ok := mod.(*IRC); ok {
			irc = mod.(*IRC)
			break
		}
	}
	if irc == nil {
		t.Fail()
	}

	r, w := net.Pipe()
	irc.conn = w

	irc.onCommand("PRIVMSG", "foo", "#candice :http://ideone.com/FllowW")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :http://sprunge.us/RWOP")
	readLog(r, t)

	// no repaste and no reply
	irc.onCommand("PRIVMSG", "foo", "#candice :http://sprunge.us/UjQf")

	irc.conn = nil
	delTestBot(bot, t, ch)
}

func TestIRCURLParserYoutube(t *testing.T) {
	ch := make(chan bool)
	bot := newTestBot(ch)
	if bot.State != Running {
		t.Fail()
	}

	var irc *IRC
	for _, mod := range bot.modules {
		if _, ok := mod.(*IRC); ok {
			irc = mod.(*IRC)
			break
		}
	}
	if irc == nil {
		t.Fail()
	}

	r, w := net.Pipe()
	irc.conn = w

	irc.onCommand("PRIVMSG", "foo", "#candice :https://www.youtube.com/watch?v=Pd12BmxP-08")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :https://youtube.com/watch?v=Pd12BmxP-08")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :https://youtu.be/QEllLECo4OM")
	readLog(r, t)

	irc.conn = nil
	delTestBot(bot, t, ch)
}
