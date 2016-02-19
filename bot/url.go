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

	"github.com/fluter01/lotsawa"

	"github.com/fluter01/paste/bpaste"
	"github.com/fluter01/paste/codepad"
	"github.com/fluter01/paste/ideone"
	"github.com/fluter01/paste/pastebin"
	"github.com/fluter01/paste/pastie"
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
	urls := xurls.Strict.FindString(req.text)
	if len(urls) == 0 {
		return "", NotParsed
	}

	p.i.Logger().Printf("URL parser processing")
	var res string

	res = fmt.Sprintf("URL is %s", urls)

	var u *url.URL
	var err error

	u, err = url.Parse(urls)
	if err != nil {
		p.i.Logger().Printf("Invalid URL: %s", urls)
		return "", NotParsed
	}

	var paste bool = false
	var getID func(string) (string, error)
	var get func(string) (string, error)
	switch strings.ToLower(u.Host) {
	default:
		if p.i.bot.config.IgnoreURLTitle(req.channel) {
			res = ""
			break
		}
		title := p.getTitle(urls)
		if title == "" {
			res = ""
		} else if req.direct {
			res = fmt.Sprintf("%s: Title of the link: %s", req.nick, title)
		} else {
			res = fmt.Sprintf("Title of %s's link: %s", req.nick, title)
		}
	case "youtube.com", "www.youtube.com":
		res = p.parseYoutube(urls)
	case "pastebin.com", "www.pastebin.com":
		paste = true
		getID = pastebin.GetID
		get = pastebin.Get
	case "codepad.org":
		paste = true
		getID = codepad.GetID
		get = codepad.Get
	case "bpaste.net":
		paste = true
		getID = bpaste.GetID
		get = bpaste.Get
	case "ideone.com":
		paste = true
		getID = ideone.GetID
		get = ideone.Get
	case "pastie.org":
		paste = true
		getID = pastie.GetID
		get = pastie.Get
	}
	if paste {
		code, compiled := p.parsePaste(urls, getID, get)
		if code == "" {
			return "", nil
		}
		res = fmt.Sprintf("%s's paste: %s", req.nick, code)
		if compiled {
			res += fmt.Sprintf(" -- issues found, please address them first!")
		}
	}
	return res, nil
}

func (p *URLParser) getTitle(urls string) string {
	const maxTitleLen = 256
	var resp *http.Response
	var err error
	var result string

	// only proceed with text documents
	resp, err = http.Head(urls)
	if err != nil {
		p.i.Logger().Printf("Http Head error: %s", err)
		return result
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/") {
		p.i.Logger().Println("Not text document, skipping")
		return result
	}

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

	if len(result) > maxTitleLen {
		result = result[:maxTitleLen] + "...<too long, truncated>"
	}

	result = strings.Replace(result, "\n", " ", -1)
	result = strings.Replace(result, "\r", "", -1)

	p.i.Logger().Println("Returning title:", result)

	return result
}

func (p *URLParser) parseYoutube(urls string) string {
	fmt.Println("Yutube:", urls)
	return ""
}

func (p *URLParser) parsePaste(urls string, getID func(string) (string, error),
	get func(string) (string, error)) (string, bool) {
	var err error
	var id string
	var data string
	var issues string

	id, err = getID(urls)
	if err != nil {
		return "", false
	}
	data, err = get(id)
	if err != nil {
		return "", false
	}

	var with_issues bool = false
	issues, err = p.submitToCompile(data)

	if err == nil && issues != "" {
		with_issues = true
	}

	if !with_issues {
		id, err = sprunge.Put(data)
		if err != nil {
			return "", false
		}
		return id, false
	}

	data = fmt.Sprintf("%s\n\n"+
		"----------------------------------------------------------------\n"+
		"%s", data, issues)
	id, err = sprunge.Put(data)
	if err != nil {
		return "", false
	}

	return id, true
}

func (p *URLParser) submitToCompile(code string) (string, error) {
	var err error

	var s *lotsawa.CompileServiceStub
	s, err = lotsawa.NewCompileServiceStub("tcp", p.i.bot.config.CompileServer)
	if err != nil {
		p.i.Logger().Println("Failed to dial rpc server:", err)
		return "", err
	}
	var arg lotsawa.CompileArgs = lotsawa.CompileArgs{code, "C"}
	var res lotsawa.CompileReply
	err = s.Compile(&arg, &res)

	if err != nil {
		p.i.Logger().Println("Failed to call rpc service:", err)
		return "", err
	}

	var issues string
	if res.C_Output != "" || res.C_Error != "" {
		issues = fmt.Sprintf("%s%s", res.C_Output, res.C_Error)
	}
	return issues, nil
}
