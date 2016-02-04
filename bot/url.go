// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"github.com/mvdan/xurls"
	_ "io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/fluter01/paste/bpaste"
	"github.com/fluter01/paste/codepad"
	"github.com/fluter01/paste/pastebin"
	"github.com/fluter01/paste/sprunge"
)

type URLParser struct {
	i *Interpreter
}

func NewURLParser(i *Interpreter) *URLParser {
	p := new(URLParser)
	p.i = i
	return p
}

func (p *URLParser) Parse(req *MessageRequest) (string, error) {
	//fmt.Println(req.text)
	urls := xurls.Strict.FindString(req.text)
	if len(urls) == 0 {
		return "", NotParsed
	}

	var res string
	// p.i.Logger().Println("URL:", urls)

	res = fmt.Sprintf("URL is %s", urls)

	var u *url.URL
	var err error

	u, err = url.Parse(urls)
	if err != nil {
		p.i.Logger().Printf("Invalid URL: %s", urls)
		return "", NotParsed
	}

	//fmt.Println(u.Host)

	switch strings.ToLower(u.Host) {
	default:
		if !p.i.bot.config.ShowURLTitle(req.channel) {
			res = ""
			break
		}
		title := p.getTitle(urls)
		if req.direct {
			res = fmt.Sprintf("%s: Title of the link: %s", req.nick, title)
		} else {
			res = fmt.Sprintf("Title of %s's link: %s", req.nick, title)
		}
	case "youtube.com", "www.youtube.com":
		res = p.parseYoutube(urls)
	case "pastebin.com", "www.pastebin.com":
		res = p.parsePastebin(urls)
		res = fmt.Sprintf("%s's paste: %s -- for those who curl",
			req.nick, res)
	case "codepad.org":
		res = p.parseCodepad(urls)
		res = fmt.Sprintf("%s's paste: %s -- for those who curl",
			req.nick, res)
	case "bpaste.net":
		res = p.parseBpaste(urls)
		res = fmt.Sprintf("%s's paste: %s -- for those who curl",
			req.nick, res)
	}
	return res, nil
}

func (p *URLParser) getTitle(urls string) string {
	var resp *http.Response
	var err error
	var result string

	resp, err = http.Get(urls)
	if err != nil {
		p.i.Logger().Printf("Http Get error: %s", err)
		return result
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	for {
		typ := z.Next()
		if typ == html.ErrorToken {
			break
		}
		if typ == html.StartTagToken {
			t := z.Token()
			if t.Data == "title" {
				z.Next()
				t := z.Token()
				result = t.Data
				break
			}
		}
	}
	return result
}

func (p *URLParser) parseYoutube(urls string) string {
	fmt.Println("Yutube:", urls)
	return ""
}

func (p *URLParser) parsePastebin(urls string) string {
	var err error
	var id string
	var data string

	id, err = pastebin.GetID(urls)
	if err != nil {
		return ""
	}
	data, err = pastebin.Get(id)
	if err != nil {
		return ""
	}

	id, err = sprunge.Put(data)
	if err != nil {
		return ""
	}
	return id
}

func (p *URLParser) parseCodepad(urls string) string {
	var err error
	var id string
	var data string

	id, err = codepad.GetID(urls)
	if err != nil {
		return ""
	}
	data, err = codepad.Get(id)
	if err != nil {
		return ""
	}

	id, err = sprunge.Put(data)
	if err != nil {
		return ""
	}
	return id
}

func (p *URLParser) parseBpaste(urls string) string {
	var err error
	var id string
	var data string

	id, err = bpaste.GetID(urls)
	if err != nil {
		return ""
	}
	data, err = bpaste.Get(id)
	if err != nil {
		return ""
	}

	id, err = sprunge.Put(data)
	if err != nil {
		return ""
	}
	return id
}
