// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"strconv"
	"strings"
)

func matchNickUserHost(target string) (string, string, string) {
	var nick, user, host string
	m := nickRe.FindStringSubmatch(target)
	if m != nil {
		nick, user, host = m[1], m[2], m[3]
	} else {
		nick = target
	}
	return nick, user, host
}

func (irc *IRC) getReplyBySpace(param string) string {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	return arr[1]
}

func (irc *IRC) getReplyByColon(param string) string {
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	return arr[1]
}

// IRC event handlers

func (irc *IRC) onPing(prefix, param string) {
	irc.Pong(param)
}

func (irc *IRC) onPong(prefix, param string) {
}

func (irc *IRC) onJoin(from, cha string) {
	var nick, user, host string
	var ch *Channel

	nick, user, host = matchNickUserHost(from)

	// confirm of channel join from server
	if nick == irc.bot.config.BotNick {
		irc.Logger().Println("New channel:", cha)
		ch = irc.JoinChannel(cha)
		ch.Start(nick)
	}

	// other users joined the channel I'm in
	ch = irc.GetChannel(cha)
	ch.onJoin(from)

	irc.bot.AddEvent(
		NewEvent(
			UserJoin,
			&UserJoinData{
				from, nick, user, host, cha}))
}

func (irc *IRC) onPart(from, param string) {
	var nick, user, host string
	var arr []string
	var chn string
	var partMsg string

	nick, user, host = matchNickUserHost(from)

	arr = strings.SplitN(param, ":", 2)
	chn = strings.TrimSpace(arr[0])
	if len(arr) == 2 {
		partMsg = arr[1]
	} else {
		partMsg = ""
	}

	var ch *Channel

	ch = irc.GetChannel(chn)
	ch.onPart(from, partMsg)

	if nick == irc.bot.config.BotNick {
		irc.Logger().Println("Leaving channel:", chn)
		ch = irc.LeaveChannel(chn)
		ch.Stop()
	}

	irc.bot.AddEvent(
		NewEvent(
			UserPart,
			&UserPartData{
				from, nick, user, host, chn, partMsg}))
}

func (irc *IRC) onQuit(from, param string) {
	var nick, user, host string
	var ch *Channel
	var quitMsg string = param[1:]

	nick, user, host = matchNickUserHost(from)

	for _, ch = range irc.channels {
		ch.onQuit(nick, from, param)
		if nick == irc.bot.config.BotNick {
			irc.Logger().Println("Leaving channel:", ch.name)
			irc.LeaveChannel(ch.name)
			ch.Stop()
		}
	}

	irc.bot.AddEvent(
		NewEvent(
			UserQuit,
			&UserQuitData{
				from, nick, user, host, quitMsg}))
}

func (irc *IRC) onNick(from, param string) {
	var nick, user, host string
	var newNick string

	newNick = param[1:]

	nick, user, host = matchNickUserHost(from)

	var ch *Channel
	for _, ch = range irc.channels {
		ch.onNick(nick, newNick)
	}

	irc.bot.AddEvent(
		NewEvent(
			UserNick,
			&UserNickData{
				from, nick, user, host, newNick}))
}

func (irc *IRC) onInvite(from, param string) {
	var nick string
	var me string
	var channel string

	nick, _, _ = matchNickUserHost(from)
	channel = irc.getReplyBySpace(param)
	me = strings.SplitN(param, " ", 2)[0]

	if me != irc.bot.config.BotNick {
		// suspicous invite message not directing to me
		panic(me)
	}

	irc.Logger().Printf("%s is inviting me to join %s", nick, channel)
	irc.Join(channel)
}

func (irc *IRC) onPrivmsg(from, param string) {
	//    var reply string
	var nick, user, host string
	var to string
	var msg string
	var i int

	nick, user, host = matchNickUserHost(from)

	for i = 0; i < len(param)-2; i++ {
		if param[i] == ' ' && param[i+1] == ':' {
			to = param[:i]
			msg = param[i+2:]
			break
		}
	}

	msg = strings.TrimSpace(msg)

	if IsChannel(to) {
		// get channel
		// send message to channel
		var ch *Channel

		ch = irc.GetChannel(to)
		ch.onPrivmsg(nick, msg)

		irc.bot.AddEvent(NewEvent(ChannelMessage,
			NewChannelMessageData(from, nick,
				user, host, msg, to)))
	} else {
		// handle ctcp
		if msg[0] == xdelim && msg[len(msg)-1] == xdelim {
			irc.onCtcp(nick, msg)
			return
		}
		irc.Logger().Printf("<%s> %s", nick, msg)
		irc.bot.AddEvent(NewEvent(PrivateMessage,
			NewPrivateMessageData(from,
				nick, user, host, msg)))
	}
}

func (irc *IRC) onNotice(from, param string) {
	var me string
	var msg string
	var arr []string
	var nick string

	nick, _, _ = matchNickUserHost(from)
	if nick != "" {
		from = nick
	}
	arr = strings.SplitN(param, ":", 2)
	me = strings.TrimSpace(arr[0])
	msg = arr[1]
	if me != irc.bot.config.BotNick {
		irc.Logger().Printf("Notice from %s to %s: %s", from, me, msg)
	} else {
		irc.Logger().Printf("Notice from %s: %s", from, msg)
	}
}

// format:
// user :candice MODE candice :+w
// channel :ChanServ!ChanServ@services. MODE #freenode +q *!*@183.185.132.59
func (irc *IRC) onMode(from, param string) {
	var target string
	var msg string
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	target, msg = arr[0], arr[1]

	if IsChannel(target) {
		ch := irc.GetChannel(target)
		if ch != nil {
			ch.onMode(msg, from)
		}
	} else {
		if from != target {
			panic(from + " != " + target)
		}
		// TODO(fluter): check first char with + or -
		irc.mode = msg
	}
}

func (irc *IRC) onError(from, param string) {
	irc.Logger().Printf("Error %s", param)
	irc.disconnect()
}

// handle numeric replies
// F: 001 candice :Welcome to the freenode Internet Relay Chat Network candice
func (irc *IRC) onRPL_WELCOME(from, param string) {
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	irc.Logger().Println(arr[1])
}

// F: 002 candice :Your host is rajaniemi.freenode.net[195.148.124.79/7000], running version ircd-seven-1.1.3
func (irc *IRC) onRPL_YOURHOST(from, param string) {
	var info string

	info = irc.getReplyByColon(param)
	m := rpl_002Re.FindStringSubmatch(info)
	irc.server.host, irc.server.version = m[1], m[3]

	irc.Logger().Println(info)
}

func (irc *IRC) onRPL_CREATED(from, param string) {
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_MYINFO(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_ISUPPORT(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_STATSCONN(from, param string) {
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_LUSERCLIENT(from, param string) {
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_LUSEROP(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_LUSERUNKNOWN(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_LUSERCHANNELS(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_LUSERME(from, param string) {
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_LOCALUSERS(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

func (irc *IRC) onRPL_GLOBALUSERS(from, param string) {
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	irc.Logger().Println(arr[1])
}

// format: 311 candice candice ~gbot unaffiliated/fluter/bot/candice * :Dr Hu Shih
func (irc *IRC) onRPL_WHOISUSER(from, param string) {
	var nick, user, host, name string
	var userstr string
	var arr []string

	userstr = irc.getReplyBySpace(param)
	arr = strings.SplitN(userstr, " ", 5)
	nick, user, host, name = arr[0], arr[1], arr[2], arr[4][1:]

	irc.Logger().Printf("[%s] (%s@%s): %s", nick, user, host, name)
}

// format: 312 candice candice rajaniemi.freenode.net :Helsinki, FI, EU
func (irc *IRC) onRPL_WHOISSERVER(from, param string) {
	var nick, svr, geo string
	var server string
	var arr []string

	server = irc.getReplyBySpace(param)
	arr = strings.SplitN(server, " ", 3)
	nick, svr, geo = arr[0], arr[1], arr[2][1:]
	irc.Logger().Printf("[%s] %s (%s)", nick, svr, geo)
}

// format:
func (irc *IRC) onRPL_WHOISOPERATOR(from, param string) {
	userstr := irc.getReplyBySpace(param)
	irc.Logger().Printf("OP: %s", userstr)
}

func (irc *IRC) onRPL_WHOWASUSER(from, param string) {
	userstr := irc.getReplyBySpace(param)
	irc.Logger().Printf("WAS: %s", userstr)
}

func (irc *IRC) onRPL_ENDOFWHO(from, param string) {
	userstr := irc.getReplyBySpace(param)
	irc.Logger().Printf("ENDWHO: %s", userstr)
}

// format: 317 candice candice 30 1452309570 :seconds idle, signon time
func (irc *IRC) onRPL_WHOISIDLE(from, param string) {
	var nick, idles, signons string
	var idle int
	var arr []string

	idlestr := irc.getReplyBySpace(param)
	arr = strings.SplitN(idlestr, " ", 4)
	nick, idles, signons = arr[0], arr[1], arr[2]

	var day, hour, min, sec int

	idle, _ = strconv.Atoi(idles)
	day = idle / (60 * 60 * 24)
	hour = (idle % (60 * 60 * 24)) / (60 * 60)
	min = ((idle % (60 * 60 * 24)) % (60 * 60)) / 60
	sec = ((idle % (60 * 60 * 24)) % (60 * 60)) % 60

	if day > 0 {
		irc.Logger().Printf("[%s] idle: "+
			"%d %s %d %s %d %s %d %s, "+
			"signon at: %s",
			nick,
			day, sp("day", "days", day),
			hour, sp("hour", "hours", hour),
			min, sp("minute", "minutes", min),
			sec, sp("second", "seconds", sec),
			unixTimeStr(signons))
	} else {
		irc.Logger().Printf("[%s] idle: "+
			"%d %s %d %s %d %s, "+
			"signon at: %s",
			nick,
			hour, sp("hour", "hours", hour),
			min, sp("minute", "minutes", min),
			sec, sp("second", "seconds", sec),
			unixTimeStr(signons))
	}
}

// format: 318 candice candice :End of /WHOIS list.
func (irc *IRC) onRPL_ENDOFWHOIS(from, param string) {
	var nick, info string
	var arr []string
	endstr := irc.getReplyBySpace(param)
	arr = strings.SplitN(endstr, " ", 2)
	nick, info = arr[0], arr[1][1:]
	irc.Logger().Printf("[%s] %s", nick, info)
}

// format: 319 candice candice :#rdma #candice
func (irc *IRC) onRPL_WHOISCHANNELS(from, param string) {
	var nick string
	var chanlist string
	var chanstr string
	var arr []string

	chanstr = irc.getReplyBySpace(param)
	arr = strings.SplitN(chanstr, " ", 2)
	nick, chanlist = arr[0], arr[1][1:]

	irc.Logger().Printf("[%s] %s", nick, chanlist)
}

func (irc *IRC) onRPL_WHOISSPECIAL(from, param string) {
	idlestr := irc.getReplyBySpace(param)
	irc.Logger().Printf("CHANNELS: %s", idlestr)
}

// format: 328 candice #freenode :http://freenode.net/
func (irc *IRC) onRPL_CHANNELURL(from, param string) {
	var chn, url string
	var arr []string

	urlstr := irc.getReplyBySpace(param)
	arr = strings.SplitN(urlstr, " ", 2)
	chn, url = arr[0], arr[1][1:]

	//	irc.Logger().Printf("URL for %s: %s", chn, url)

	var ch *Channel

	ch = irc.GetChannel(chn)
	if ch != nil {
		ch.onRPL_CHANNELURL(url)
	}
}

// format:
func (irc *IRC) onRPL_CREATIONTIME(from, param string) {
}

// format: 330 candice fluter fluter :is logged in as
func (irc *IRC) onRPL_WHOISLOGGEDIN(from, param string) {
	var nick, as, info string
	var arr []string

	logstr := irc.getReplyBySpace(param)
	arr = strings.SplitN(logstr, " ", 3)
	nick, as, info = arr[0], arr[1], arr[2][1:]
	irc.Logger().Printf("[%s] %s %s", nick, info, as)
}

func (irc *IRC) onRPL_NOTOPIC(from, param string) {
}

// format: 332 candice #hpc :meh - All things High Performance Computing (HPC) / Parallel Programming / Decoding 42
func (irc *IRC) onRPL_TOPIC(from, param string) {
	var chn, topic string
	var arr []string

	topicstr := irc.getReplyBySpace(param)
	arr = strings.SplitN(topicstr, " ", 2)
	chn, topic = arr[0], arr[1][1:]

	var ch *Channel
	ch = irc.GetChannel(chn)
	if ch != nil {
		ch.onRPL_TOPIC(topic)
	}
}

// format: 333 candice #hpc EOF!~hamiltonh@dsl-173-206-247-218.tor.primus.ca 1290112601
func (irc *IRC) onRPL_TOPICWHOTIME(from, param string) {
	var chn, who, time string
	var arr []string

	topicstr := irc.getReplyBySpace(param)
	arr = strings.SplitN(topicstr, " ", 3)
	chn, who, time = arr[0], arr[1], arr[2]

	var nick, user, host string
	var timestr string

	nick, user, host = matchNickUserHost(who)
	timestr = unixTimeStr(time)

	if nick != "" {
		who = fmt.Sprintf("%s (%s@%s)", nick, user, host)
	}
	//	irc.Logger().Printf("Topic for %s set by %s on %s", chn,
	//		who, timestr)

	var ch *Channel
	ch = irc.GetChannel(chn)
	if ch != nil {
		ch.onRPL_TOPICWHOTIME(who, timestr)
	}
}

func (irc *IRC) onRPL_NAMREPLY(from, param string) {
	var chn string
	var me string
	var mode byte
	var arr []string
	var nicks string

	arr = strings.SplitN(param, ":", 2)
	if len(arr) != 2 {
		panic(arr)
	}
	nicks = arr[1]
	arr = strings.Split(strings.TrimSpace(arr[0]), " ")
	if len(arr) != 3 {
		panic(arr)
	}
	me, mode, chn = arr[0], arr[1][0], arr[2]

	if me != irc.bot.config.BotNick {
		panic(me)
	}

	var ch *Channel

	ch = irc.GetChannel(chn)
	ch.mode = mode
	ch.onRPL_NAMREPLY(nicks)
}

// format: 366 fluter #botters-test :End of /NAMES list.
func (irc *IRC) onRPL_ENDOFNAMES(from, param string) {
	var chn string
	var me string
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	if len(arr) != 2 {
		panic(arr)
	}
	// arr[1] = :End of /NAMES list.
	arr = strings.Split(strings.TrimSpace(arr[0]), " ")
	if len(arr) != 2 {
		panic(arr)
	}
	me, chn = arr[0], arr[1]

	if me != irc.bot.config.BotNick {
		panic(me)
	}
	if irc.GetChannel(chn) == nil {
		panic(chn)
	}

	var ch *Channel

	ch = irc.GetChannel(chn)
	ch.onRPL_ENDOFNAMES()
}

func (irc *IRC) onRPL_MOTD(from, param string) {
	var motd string
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	motd = arr[1]
	irc.Logger().Printf("MOTD: %s", motd)
}

func (irc *IRC) onRPL_MOTDSTART(from, param string) {
	var start string
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	start = arr[1]
	irc.Logger().Printf("      %s", start)
}

func (irc *IRC) onRPL_ENDOFMOTD(from, param string) {
	// end of motd
}

// format: 378 candice candice :is connecting from *@69.4.235.219 69.4.235.219
func (irc *IRC) onRPL_WHOISHOST(from, param string) {
	var nick, info string
	var arr []string

	hoststr := irc.getReplyBySpace(param)
	arr = strings.SplitN(hoststr, " ", 2)
	nick, info = arr[0], arr[1][1:]

	irc.Logger().Printf("[%s] %s", nick, info)
}

func (irc *IRC) onRPL_HOSTHIDDEN(from, param string) {
	var me string
	var mask string
	var arr []string

	arr = strings.SplitN(param, ":", 2)
	if len(arr) != 2 {
		panic(arr)
	}

	// arr[1]=:is now your hidden host (set by services.)
	arr = strings.Split(strings.TrimSpace(arr[0]), " ")
	if len(arr) != 2 {
		panic(arr)
	}
	me = arr[0]
	if me != irc.bot.config.BotNick {
		panic(me)
	}
	mask = arr[1]
	irc.host = mask
}

func (irc *IRC) onERR_INVITEONLYCHAN(from, param string) {
	var errmsg string
	var arr []string

	arr = strings.SplitN(param, " ", 2)
	errmsg = arr[1]
	irc.Logger().Printf("%s", errmsg)
}

// format: 671 candice fluter :is using a secure connection
func (irc *IRC) onRPL_WHOISSECURE(from, param string) {
	var nick, info string
	var arr []string

	secstr := irc.getReplyBySpace(param)
	arr = strings.SplitN(secstr, " ", 2)
	nick, info = arr[0], arr[1][1:]

	irc.Logger().Printf("[%s] %s", nick, info)
}

// end IRC command handlers
