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
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	t.Log(string(buf[:n]))

	irc.conn = nil
	delTestBot(bot, t, ch)
}
