// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	stdLog "log"
	"strings"
)

type Empty struct{}

const (
	IN  = "-->"
	OUT = "<--"
	NOM = "--"
)

type Channel struct {
	irc *IRC

	name string

	topic string

	url string

	mode byte

	users  map[string]Empty
	nop    int
	nvoice int

	logger Logger
}

func NewChannel(irc *IRC, name string) *Channel {
	ch := new(Channel)
	ch.irc = irc
	ch.name = name
	ch.logger = NewFileLogger(irc.bot,
		fmt.Sprintf("%s-%s", irc.bot.config.Name, name))
	ch.logger.SetFlags(stdLog.LstdFlags)
	ch.users = make(map[string]Empty)

	return ch
}

func (ch *Channel) String() string {
	var nicks []string
	nicks = make([]string, 0, len(ch.users))
	i := 0
	for nick := range ch.users {
		nicks = append(nicks, nick)
		i++
		if i > 5 {
			nicks = append(nicks, "...")
			break
		}
	}
	return fmt.Sprintf("%s %d%s", ch.name, len(ch.users), nicks)
}

func (ch *Channel) Log(dir, format string, param ...interface{}) {
	newf := fmt.Sprintf("%s\t%s", dir, format)
	ch.logger.Printf(newf, param...)
}

func (ch *Channel) Start(me string) {
	ch.add(me)
}

func (ch *Channel) Stop() {
	ch.users = make(map[string]Empty)
}

func (ch *Channel) Topic() string {
	return ch.topic
}

func (ch *Channel) SetTopic(topic string) {
	ch.topic = topic
}

// user management
func (ch *Channel) add(nick string) {
	ch.users[nick] = Empty{}
}

func (ch *Channel) contains(nick string) bool {
	_, ok := ch.users[nick]
	return ok
}

func (ch *Channel) remove(nick string) {
	delete(ch.users, nick)
}

// command handlers
func (ch *Channel) onPrivmsg(from, msg string) {
	ch.logger.Printf("<%s>\t%s", from, msg)
}

func (ch *Channel) onJoin(from string) {
	var m []string
	var nick string

	m = nickRe.FindStringSubmatch(from)
	if m != nil {
		nick = m[1]
	} else {
		panic(from)
	}

	ch.add(nick)
	ch.Log(IN, "%s (%s) has joined %s",
		nick, from, ch.name)
}

func (ch *Channel) onPart(from, msg string) {
	var m []string
	var nick string

	m = nickRe.FindStringSubmatch(from)
	if m != nil {
		nick = m[1]
	} else {
		panic(from)
	}

	ch.remove(nick)

	if msg != "" {
		ch.Log(OUT, "%s (%s) has left %s (%s)",
			nick, from, ch.name, msg)
	} else {
		ch.Log(OUT, "%s (%s) has left %s",
			nick, from, ch.name)
	}
}

func (ch *Channel) onQuit(nick, from, msg string) {
	if !ch.contains(nick) {
		return
	}

	ch.remove(nick)
	if msg != "" {
		ch.Log(OUT, "%s (%s) has quit (%s)",
			nick, from, msg[1:])
	} else {
		ch.Log(OUT, "%s (%s) has quit",
			nick, from)
	}
}

func (ch *Channel) onNick(nick, newNick string) {
	if !ch.contains(nick) {
		return
	}
	ch.remove(nick)
	ch.add(newNick)
	ch.Log(NOM, "%s is now known as %s", nick, newNick)
}

func (ch *Channel) onMode(mode, from string) {
	nick, _, _ := matchNickUserHost(from)
	ch.Log(NOM, "Mode %s [%s] by %s", ch.name, mode, nick)
}

func (ch *Channel) onRPL_CHANNELURL(url string) {
	ch.url = url
	ch.Log(NOM, "URL for %s: %s", ch.name, url)
}

func (ch *Channel) onRPL_TOPIC(topic string) {
	ch.SetTopic(topic)
	ch.Log(NOM, "Topic for %s is \"%s\"",
		ch.name, topic)
}

func (ch *Channel) onRPL_TOPICWHOTIME(who, time string) {
	ch.Log(NOM, "Topic set by %s on %s",
		who, time)
}

func (ch *Channel) onRPL_NAMREPLY(nicks string) {
	var arr []string

	arr = strings.Split(nicks, " ")
	for _, nick := range arr {
		switch nick[0] {
		default:
			ch.add(nick)
		case '@':
			ch.add(nick[1:])
			ch.nop++
		case '+':
			ch.add(nick[1:])
			ch.nvoice++
		}
	}
}

func (ch *Channel) onRPL_ENDOFNAMES() {
	ch.Log(NOM, "Channel %s: %d nicks (%d op, %d voices, %d normals)",
		ch.name,
		len(ch.users),
		ch.nop,
		ch.nvoice,
		len(ch.users)-ch.nop-ch.nvoice,
	)
}
