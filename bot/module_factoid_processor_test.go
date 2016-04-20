package bot

import (
	"net"
	"testing"
)

func TestFactoidProcessor(t *testing.T) {
	ch := make(chan bool)
	bot := newTestBot(ch)

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

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd global hi hello")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd global int16 16bits")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd #c NULL Null is null pointer")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd #c int Integer")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd #c int32_t 32bits integer")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd #c int64_t 64bits int")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factadd #c malloc malloc")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factfind -channel foo -by bar hi")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factfind -channel #c int")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factfind int")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factinfo int16")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factinfo #c int")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": factshow #c int32_t")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": fact #c int32_t")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+G+": fact int32_t")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#c :"+G+": fact int32_t")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#c :"+G+": fact int16")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#c :"+G+": int32_t")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "#c :?int32_t")
	readLog(r, t)

	irc.onCommand("PRIVMSG", "foo", "foo :hi")
	readLog(r, t)

	irc.conn = nil
	delTestBot(bot, t, ch)
}
