// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
)

// IRC commands, defined by RFC 2812

func (irc *IRC) Nick(nick string) error {
	msg := fmt.Sprintf("NICK %s", nick)
	return irc.sendMsg(msg)
}

func (irc *IRC) Oper(name, passwd string) error {
	msg := fmt.Sprintf("OPER %s %s", name, passwd)
	return irc.sendMsg(msg)
}

func (irc *IRC) Mode(nick string, mode string) error {
	msg := fmt.Sprintf("MODE %s %s", nick, mode)
	return irc.sendMsg(msg)
}

func (irc *IRC) Service(nick, dist, typ, info string) error {
	msg := fmt.Sprintf("SERVICE %s * %s %s * %s",
		nick, dist, typ, info)
	return irc.sendMsg(msg)
}

func (irc *IRC) Quit(qmsg string) error {
	var msg string
	if qmsg == "" {
		msg = fmt.Sprintf("QUIT")
	} else {
		msg = fmt.Sprintf("QUIT :%s", qmsg)
	}
	return irc.sendMsg(msg)
}

func (irc *IRC) Join(channel string) error {
	msg := fmt.Sprintf("JOIN %s", channel)
	return irc.sendMsg(msg)
}

func (irc *IRC) Part(channel, partMsg string) error {
	var msg string
	if partMsg != "" {
		msg = fmt.Sprintf("PART %s %s", channel, partMsg)
	} else {
		msg = fmt.Sprintf("PART %s", channel)
	}
	return irc.sendMsg(msg)
}

// func (irc *IRC) Squit(server, command string) error
// func (irc *IRC) ModeChan(server, command string) error

func (irc *IRC) GetTopic(channel string) error {
	msg := fmt.Sprintf("TOPIC %s", channel)
	return irc.sendMsg(msg)
}

func (irc *IRC) SetTopic(channel, topic string) error {
	msg := fmt.Sprintf("TOPIC %s :%s", channel, topic)
	return irc.sendMsg(msg)
}

func (irc *IRC) Names(channel string) error {
	msg := fmt.Sprintf("NAMES %s", channel)
	return irc.sendMsg(msg)
}

// list
// invite
// :adams.freenode.net 341 fluter candice #rdma
func (irc *IRC) Invite(nick, channel string) error {
	msg := fmt.Sprintf("INVITE %s %s", nick, channel)
	return irc.sendMsg(msg)
}
// kick

func (irc *IRC) Privmsg(to, msg string) error {
	var line string

	if IsChannel(to) {
		var ch *Channel

		ch = irc.GetChannel(to)
		ch.onPrivmsg(irc.bot.config.BotNick, msg)
	}
	line = fmt.Sprintf("PRIVMSG %s :%s", to, msg)

	return irc.sendMsg(line)
}

func (irc *IRC) Notice(to, msg string) error {
	var line string

	line = fmt.Sprintf("NOTICE %s :%s", to, msg)

	return irc.sendMsg(line)
}

// server query
func (irc *IRC) Motd() error {
	return irc.sendMsg("MOTD")
}

func (irc *IRC) Lusers() error {
	return irc.sendMsg("LUSERS")
}

func (irc *IRC) Version() error {
	return irc.sendMsg("VERSION")
}

func (irc *IRC) Stats(query string) error {
	msg := fmt.Sprintf("STATS %s", query)
	return irc.sendMsg(msg)
}

func (irc *IRC) Links() error {
	return irc.sendMsg("LINKS")
}

func (irc *IRC) Time() error {
	return irc.sendMsg("TIME")
}

// connect

func (irc *IRC) Trace(target string) error {
	msg := fmt.Sprintf("TRACE %s", target)
	return irc.sendMsg(msg)
}

// admin

func (irc *IRC) Info(target string) error {
	msg := fmt.Sprintf("INFO %s", target)
	return irc.sendMsg(msg)
}

// servlist
// squery - not supported by freenode

// user based queries

func (irc *IRC) Who(mask string, flag string) error {
	msg := fmt.Sprintf("WHO %s %s", mask, flag)
	return irc.sendMsg(msg)
}

func (irc *IRC) Whois(target string, mask string) error {
	msg := fmt.Sprintf("WHOIS %s %s", target, mask)
	return irc.sendMsg(msg)
}

func (irc *IRC) Whowas(nick string) error {
	msg := fmt.Sprintf("WHOWAS %s", nick)
	return irc.sendMsg(msg)
}

// miscellaneous messages

func (irc *IRC) Kill(nick, comment string) error {
	msg := fmt.Sprintf("KILL %s %s", nick, comment)
	return irc.sendMsg(msg)
}

func (irc *IRC) Ping(server string) error {
	msg := fmt.Sprintf("PING %s", server)
	return irc.sendMsg(msg)
}

func (irc *IRC) Pong(server string) error {
	msg := fmt.Sprintf("PONG %s", server)
	return irc.sendMsg(msg)
}

// error

func (irc *IRC) Away(text string) error {
	msg := fmt.Sprintf("AWAY %s", text)
	return irc.sendMsg(msg)
}

// rehash
// die
// restart
// summon

func (irc *IRC) Users(target string) error {
	msg := fmt.Sprintf("USERS %s", target)
	return irc.sendMsg(msg)
}

func (irc *IRC) WallOps(text string) error {
	msg := fmt.Sprintf("WALLOPS %s", text)
	return irc.sendMsg(msg)
}

func (irc *IRC) Userhost(nick string) error {
	msg := fmt.Sprintf("USERHOST %s", nick)
	return irc.sendMsg(msg)
}

func (irc *IRC) IsOn(nick string) error {
	msg := fmt.Sprintf("ISON %s", nick)
	return irc.sendMsg(msg)
}

// end IRC commands
