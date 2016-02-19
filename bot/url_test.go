// Copyright 2016 Alex Fluter

package bot

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mvdan/xurls"
	"golang.org/x/net/html"
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
	urls := []string{
		"http://irc-bot-science.clsr.net/long",
		"http://irc-bot-science.clsr.net/ctcp",
		"http://irc-bot-science.clsr.net/internet.gz",
		"http://irc-bot-science.clsr.net/tags",
		"http://irc-bot-science.clsr.net/fakelength"}
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
	t.Log(http.DefaultClient.Timeout)
}

func getClient() *http.Client {
	client := new(http.Client)

	client.Timeout = 3 * time.Second
	return client
}

func TestURL3(t *testing.T) {
	s := "http://irc-bot-science.clsr.net/fakelength"
	var u *url.URL
	var err error

	client := getClient()

	u, err = url.Parse(s)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", u)

	rsp, err := client.Head(s)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", rsp.Header)
	t.Logf("%s\n", rsp.Header.Get("Content-Type"))
	t.Logf("%s\n", rsp.Header.Get("content-length"))
	if strings.HasPrefix(rsp.Header.Get("Content-Type"), "text/") {
		t.Log("ok:", s)
	}

	rsp, err = client.Get(s)
	if err != nil {
		t.Error(err)
	}
	z := html.NewTokenizer(rsp.Body)
	for {
		typ := z.Next()
		t.Log(typ)
		if typ == html.ErrorToken {
			break
		}
		t.Log(z.Token())
		if typ == html.StartTagToken {
			tk := z.Token()
			if tk.Data == "title" {
				z.Next()
				tk := z.Token()
				t.Log("title:", tk.Data)
				break
			}
		}
	}
}
