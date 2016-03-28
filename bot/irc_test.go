// Copyright 2016 Alex Fluter

package bot

import (
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
