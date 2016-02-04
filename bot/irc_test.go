// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"os"
	"testing"
)

func matchTgtRe(s string, t *testing.T) {
	var m []string
	m = nickRe.FindStringSubmatch(s)
	if m == nil {
		t.Fail()
	}
	fmt.Println(m)
}

func TestTgtRe(t *testing.T) {
	var s string

	s = "fe!~fe@servx.ru"
	matchTgtRe(s, t)

	s = "jayeshsolanki!~jayeshsol@219.91.250.106"
	matchTgtRe(s, t)

	s = "j!~jayeshsol@219.91.250.106"
	matchTgtRe(s, t)

	s = "de-facto_!~de-facto@unaffiliated/de-facto"
	matchTgtRe(s, t)

	s = "c!~c@freenode"
	matchTgtRe(s, t)
}

func getBot() *Bot {
	fmt.Println(os.Getwd())
	fmt.Scan()
	config := &BotConfig{Name: "TESTCONFIG", LogDir: "..\\log"}
	bot := NewBot("Test Bot", config)
	return bot
}

func testURLParser(u string, e string, t *testing.T) {
	p := NewURLParser(nil)
	r := p.getTitle(u)
	fmt.Println(r)
	if r != e {
		t.Fail()
	}
}

func TestURLParser1(t *testing.T) {
	var u string
	u = "https://www.reddit.com/r/windows10"
	testURLParser(u, "Windows 10", t)

	u = "https://golang.org/"
	testURLParser(u, "The Go Programming Language", t)
}
