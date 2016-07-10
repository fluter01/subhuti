package bot

import (
	"net"
	"testing"
)

func TestCjeopardy(t *testing.T) {
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

	//	bot.AddEvent(NewEvent(ChannelMessage, nil))

	s1 := "8) [3.6 Terms, definitions, and symbols] A byte is composed of a contiguous sequence of bits, the number of which is this.|implementation-defined{Bet you thought it was 8!}"
	s2 := "9) [3.6 Terms, definitions, and symbols] The least significant bit is called this.|low-order bit"
	irc.onCommand("PRIVMSG", "foo", "#candice :"+s1)

	irc.onCommand("PRIVMSG", "foo", "#candice :"+cjeopardy_prefix+s1)
	irc.onCommand("PRIVMSG", "foo", "#candice :"+cjeopardy_prefix+s2)

	irc.onCommand("PRIVMSG", cjeopardy_modnick, "#candice :"+cjeopardy_prefix+s1)
	irc.onCommand("PRIVMSG", cjeopardy_modnick, "#candice :"+cjeopardy_prefix+s2)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	t.Log(string(buf[:n]))
	n, _ = r.Read(buf)
	t.Log(string(buf[:n]))

	irc.conn = nil
	delTestBot(bot, t, ch)
}
