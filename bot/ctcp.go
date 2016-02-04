// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"strings"
	"time"
)

const xdelim = '\001'

type CTCP string

const (
	FINGER     CTCP = "FINGER"
	VERSION    CTCP = "VERSION"
	SOURCE     CTCP = "SOURCE"
	USERINFO   CTCP = "USERINFO"
	CLIENTINFO CTCP = "CLIENTINFO"
	ERRMSG     CTCP = "ERRMSG"
	PING       CTCP = "PING"
	TIME       CTCP = "TIME"
)

// commands
func (irc *IRC) Ctcp(target, typ, param string) {
	typ = strings.ToUpper(typ)
	switch CTCP(typ) {
	case FINGER:
		irc.Ctcp_Finger(target)
	case VERSION:
		irc.Ctcp_Version(target)
	case SOURCE:
		irc.Ctcp_Source(target)
	case USERINFO:
		irc.Ctcp_Userinfo(target)
	case CLIENTINFO:
		irc.Ctcp_Clientinfo(target)
	case ERRMSG:
		irc.Ctcp_Errmsg(target, param)
	case PING:
		irc.Ctcp_Ping(target)
	case TIME:
		irc.Ctcp_Time(target)
	}
}

func (irc *IRC) ctcp(target string, typ CTCP, param string) error {
	var msg string

	if param == "" {
		msg = fmt.Sprintf("%c%s%c", xdelim, typ, xdelim)
	} else {
		msg = fmt.Sprintf("%c%s %s%c", xdelim, typ, param, xdelim)
	}

	return irc.Privmsg(target, msg)
}

func (irc *IRC) Ctcp_Finger(target string) error {
	return irc.ctcp(target, FINGER, "")
}

func (irc *IRC) Ctcp_Version(target string) error {
	return irc.ctcp(target, VERSION, "")
}

func (irc *IRC) Ctcp_Source(target string) error {
	return irc.ctcp(target, SOURCE, "")
}

func (irc *IRC) Ctcp_Userinfo(target string) error {
	return irc.ctcp(target, USERINFO, "")
}

func (irc *IRC) Ctcp_Clientinfo(target string) error {
	return irc.ctcp(target, CLIENTINFO, "")
}

func (irc *IRC) Ctcp_Errmsg(target, msg string) error {
	return irc.ctcp(target, ERRMSG, msg)
}

func (irc *IRC) Ctcp_Ping(target string) error {
	var ts int64

	ts = time.Now().UnixNano()

	return irc.ctcp(target, PING, fmt.Sprintf("%d", ts))
}

func (irc *IRC) Ctcp_Time(target string) error {
	return irc.ctcp(target, TIME, "")
}

// events

func (irc *IRC) onCtcp(target, msg string) {
	var typ, param string
	msg = msg[1 : len(msg)-1]
	arr := strings.Split(msg, " ")

	typ = arr[0]
	if len(arr) > 1 {
		param = strings.Join(arr[1:], " ")
	}

	switch CTCP(typ) {
	default:
		irc.Logger().Printf("Unknow CTCP requested by %s: %s",
			target, msg)
	case FINGER:
		irc.onCtcp_Finger(target)
	case VERSION:
		irc.onCtcp_Version(target)
	case SOURCE:
		irc.onCtcp_Source(target)
	case USERINFO:
		irc.onCtcp_Userinfo(target)
	case CLIENTINFO:
		irc.onCtcp_Clientinfo(target)
	case ERRMSG:
		irc.onCtcp_Errmsg(target, param)
	case PING:
		irc.onCtcp_Ping(target, param)
	case TIME:
		irc.onCtcp_Time(target)
	}
}

func (irc *IRC) CtcpReply(typ CTCP, target, reply string) {
	var msg string

	if reply == "" {
		msg = fmt.Sprintf("%c%s%c", xdelim, typ, xdelim)
	} else {
		msg = fmt.Sprintf("%c%s %s%c", xdelim, typ, reply, xdelim)
	}
	irc.Notice(target, msg)
}

func (irc *IRC) onCtcp_Finger(target string) {
	irc.CtcpReply(FINGER, target, Version())
}

func (irc *IRC) onCtcp_Version(target string) {
	irc.CtcpReply(VERSION, target, Version())
}

func (irc *IRC) onCtcp_Source(target string) {
	irc.CtcpReply(SOURCE, target, Source())
}

func (irc *IRC) onCtcp_Userinfo(target string) {
	reply := fmt.Sprintf("%s (%s)",
		irc.bot.config.BotNick,
		irc.bot.config.RealName)
	irc.CtcpReply(USERINFO, target, reply)
}

func (irc *IRC) onCtcp_Clientinfo(target string) {
	reply := "CLIENTINFO FINGER PING SOURCE TIME USERINFO VERSION"
	irc.CtcpReply(CLIENTINFO, target, reply)
}

func (irc *IRC) onCtcp_Errmsg(target, param string) {
}

func (irc *IRC) onCtcp_Ping(target, param string) {
	irc.CtcpReply(PING, target, param)
}

func (irc *IRC) onCtcp_Time(target string) {
	irc.CtcpReply(TIME, target, time.Unix(0, 0).Format(time.RFC1123))
}
