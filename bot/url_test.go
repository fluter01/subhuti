// Copyright 2016 Alex Fluter

package bot

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/mvdan/xurls"
)

func TestURL1(t *testing.T) {
	s := "https://golang.org/ref/spec#For_statements"

	var b bool
	var m []string
	b = xurls.Strict.MatchString(s)
	if !b {
		t.Log("not match")
		t.Fail()
	}

	m = xurls.Strict.FindAllString(s, -1)
	for i, e := range m {
		t.Logf("%d: %s", i, e)
	}
	if m == nil {
		t.Log("does not contain url")
		t.Fail()
	}
}

func TestURL2(t *testing.T) {
	urls := []string{"http://sprunge.us/CUAZ",
		"http://irc-bot-science.clsr.net/long",
		"http://irc-bot-science.clsr.net/ctcp",
		"http://irc-bot-science.clsr.net/internet.gz",
		"http://irc-bot-science.clsr.net/tags"}
	for _, s := range urls {
		var u *url.URL
		var err error

		u, err = url.Parse(s)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%#v\n", u)

		rsp, err := http.Head(s)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%#v\n", rsp.Header.Get("Content-Type"))
		t.Logf("%#v\n", rsp.Header.Get("content-length"))
		if strings.HasPrefix(rsp.Header.Get("Content-Type"), "text/") {
			t.Log("ok:", s)
		}
	}
}
