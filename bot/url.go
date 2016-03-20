// Copyright 2016 Alex Fluter

package bot

import (
	"errors"
	"fmt"
	"github.com/mvdan/xurls"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/fluter01/lotsawa"

	"github.com/fluter01/paste"
)

const (
	fnamePtn    = "filename=\"(.*)\""
	httpTimeout = 5
	maxRedirect = 3
)

var fnameRe = regexp.MustCompile(fnamePtn)

type URLParser struct {
	i      *Interpreter
	client *http.Client
}

func NewURLParser(i *Interpreter) *URLParser {
	p := new(URLParser)
	p.i = i
	client := new(http.Client)
	client.Timeout = httpTimeout * time.Second
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirect {
			return errors.New(fmt.Sprintf("stopped after %d redirects", maxRedirect))
		}
		return nil
	}
	p.client = client
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

	host := strings.ToLower(u.Host)
	switch host {
	case "youtube.com", "www.youtube.com":
		res = p.parseYoutube(urls)
	// handle pastebins
	default:
		data, err := paste.Get(urls)
		if err != nil {
			if err == paste.NotSupported {
				p.i.Logger().Printf("Unhandled URL, getting title")
				if p.i.bot.config.IgnoreURLTitle(req.channel) {
					p.i.Logger().Printf("Ignoring url title for %s", req.channel)
					res = ""
					break
				}
				title := p.getTitle(urls)
				if title == "" {
					res = ""
				} else if !req.ischan {
					res = fmt.Sprintf("Your link: %s", title)
				} else if req.direct {
					res = fmt.Sprintf("%s: Your link: %s", req.nick, title)
				} else {
					res = fmt.Sprintf("%s's link: %s", req.nick, title)
				}
			} else {
				p.i.Logger().Printf("error get url: %s", err)
			}
			break
		}
		p.i.Logger().Printf("code paste, compiling")
		cmplres, issues := p.processCode(data)
		if cmplres == "" {
			return "", nil
		}
		res = fmt.Sprintf("%s's paste: %s", req.nick, cmplres)
		if issues {
			res += fmt.Sprintf(" -- issues found, please address them first!")
		} else {
			// do not repaste from sprunge
			if host == "sprunge.us" {
				return "", nil
			}
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
	resp, err = p.client.Head(urls)
	if err != nil {
		p.i.Logger().Printf("Http Head error: %s", err)
		return result
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/") {
		p.i.Logger().Println("Not text document, skipping download")
		// get file name if present
		disp := resp.Header.Get("Content-Disposition")
		m := fnameRe.FindStringSubmatch(disp)
		if len(m) == 2 && len(m[1]) > 0 {
			result = "filename: " + m[1]
		}
		return result
	}

	resp, err = p.client.Get(urls)
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

	// remove message seperators
	result = strings.Replace(result, "\n", " ", -1)
	result = strings.Replace(result, "\r", "", -1)
	// remove ctcp SOH
	result = strings.Replace(result, "\001", "", -1)

	p.i.Logger().Println("Returning title:", result)

	return result
}

func (p *URLParser) parseYoutube(urls string) string {
	fmt.Println("Yutube:", urls)
	return ""
}

func (p *URLParser) processCode(code string) (string, bool) {
	var err error
	var id string
	var issues string

	var with_issues bool = false
	issues, err = p.submitToCompile(code)

	if err == nil && issues != "" {
		with_issues = true
	}

	if !with_issues {
		id, err = paste.Paste(code)
		if err != nil {
			return "", false
		}
		return id, false
	}

	data := fmt.Sprintf("%s\n\n"+
		"----------------------------------------------------------------\n"+
		"%s", code, issues)
	id, err = paste.Paste(data)
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
	defer s.Close()
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
