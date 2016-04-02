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

	"github.com/fluter01/lotsawa"
	"github.com/fluter01/paste"

	"golang.org/x/net/html"
	"google.golang.org/api/youtube/v3"
)

const (
	fnamePtn    = "filename=\"(.*)\""
	httpTimeout = 5
	maxRedirect = 3
)

var fnameRe = regexp.MustCompile(fnamePtn)

// API key for youtube client
type APIKey struct {
	key string
}

func (k *APIKey) Get() (string, string) {
	return "key", k.key
}

type URLParser struct {
	i      *Interpreter
	client *http.Client
	key    *APIKey
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
	if i.irc.bot.config.YoutubeAPIKey != "" {
		p.key = &APIKey{i.irc.bot.config.YoutubeAPIKey}
	} else {
		p.key = nil
	}
	p.client = client
	return p
}

func (p *URLParser) sendReply(res string, req *MessageRequest) error {
	if res == "" {
		return nil
	}
	if req.ischan {
		return p.i.irc.Privmsg(req.channel, res)
	}
	return p.i.irc.Privmsg(req.nick, res)
}

func (p *URLParser) Parse(bot *Bot, req *MessageRequest) error {
	urls := xurls.Strict.FindString(req.text)
	if len(urls) == 0 {
		return ErrNotParsed
	}

	p.i.Logger.Printf("URL parser: %s", urls)

	var u *url.URL
	var err error

	u, err = url.Parse(urls)
	if err != nil {
		p.i.Logger.Printf("Invalid URL: %s", urls)
		return ErrNotParsed
	}

	host := strings.ToLower(u.Host)
	switch host {
	case "youtube.com", "www.youtube.com", "youtu.be":
		if p.key != nil {
			res := p.parseYoutube(urls, u)
			if res != "" {
				return p.sendReply(fmt.Sprintf("%s's video: %s",
					req.nick, res), req)
			}
		}
		fallthrough
	// handle pastebins
	default:
		data, err := paste.Get(urls)
		if err != nil {
			if err == paste.ErrNotSupported {
				p.i.Logger.Printf("Unhandled URL, getting title")
				if p.i.irc.config.IgnoreURLTitle(req.channel) {
					p.i.Logger.Printf("Ignoring url title for %s", req.channel)
					break
				}
				title := p.getTitle(urls)
				if title != "" {
					var res string
					if !req.ischan {
						res = fmt.Sprintf("Your link: %s", title)
					} else {
						res = fmt.Sprintf("%s's link: %s", req.nick, title)
					}
					return p.sendReply(res, req)
				}
			} else {
				p.i.Logger.Printf("error get url: %s", err)
			}
			break
		}
		p.i.Logger.Printf("code paste, compiling")
		cmplres, have_issues := p.processCode(data)
		if cmplres == "" {
			return nil
		}
		res := fmt.Sprintf("%s's paste: %s", req.nick, cmplres)
		if have_issues {
			res += fmt.Sprintf(" -- issues found, please address them first!")
			return p.sendReply(res, req)
		} else {
			// do not repaste from sprunge
			if host == "sprunge.us" {
				return nil
			}
			if req.irc.config.ChannelRepaste(req.channel) {
				return p.sendReply(res, req)
			}
		}
	}
	return nil
}

func (p *URLParser) getTitle(urls string) string {
	const maxTitleLen = 256
	var resp *http.Response
	var err error
	var result string

	// only proceed with text documents
	resp, err = p.client.Head(urls)
	if err != nil {
		p.i.Logger.Printf("Http Head error: %s", err)
		return result
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/") {
		p.i.Logger.Println("Not text document, skipping download")
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
		p.i.Logger.Printf("Http Get error: %s", err)
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

	p.i.Logger.Println("Returning title:", result)

	return result
}

func (p *URLParser) parseYoutubeGetId(u *url.URL) string {
	switch u.Host {
	case "youtu.be":
		return strings.TrimPrefix(u.Path, "/")
	case "youtube.com", "www.youtube.com":
		return u.Query().Get("v")
	default:
		return ""
	}
	return ""
}

func (p *URLParser) parseYoutube(urls string, u *url.URL) string {
	id := p.parseYoutubeGetId(u)
	if id == "" {
		p.i.Logger.Println("cannot parse id")
		return ""
	}
	youtube, err := youtube.New(&http.Client{})
	if err != nil {
		p.i.Logger.Println("youtube new service:", err)
		return ""
	}

	opt := p.key
	call := youtube.Videos.List("snippet,contentDetails,statistics").Id(id)
	rsp, err := call.Do(opt)
	if err != nil {
		p.i.Logger.Println("youtube call error:", err)
		return ""
	}
	if len(rsp.Items) == 0 {
		p.i.Logger.Printf("list of %s returned no result", id)
		return ""
	}

	var res string
	v := rsp.Items[0]
	s := v.Snippet
	c := v.ContentDetails
	t := v.Statistics

	if s == nil {
		p.i.Logger.Printf("no snippet returned")
		return ""
	}

	res = fmt.Sprintf("%s by %s", s.Title, s.ChannelTitle)
	if c != nil {
		var h, m, s int
		var duration string
		_, err = fmt.Sscanf(c.Duration, "PT%dH%dM%dS", &h, &m, &s)
		if err != nil {
			_, err = fmt.Sscanf(c.Duration, "PT%dM%dS", &m, &s)
			if err != nil {
				duration = ""
			} else {
				duration = fmt.Sprintf("%d:%02d", m, s)
			}
		} else {
			duration = fmt.Sprintf("%d:%02d:%02d", h, m, s)
		}
		if duration != "" {
			res += fmt.Sprintf(" %s", duration)
		}
	}
	if t != nil {
		res += fmt.Sprintf(" | view %d like %d dislike %d comment %d",
			t.ViewCount, t.LikeCount, t.DislikeCount, t.CommentCount)
	}

	return res
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

	s, err = lotsawa.NewCompileServiceStub("tcp", p.i.irc.bot.config.CompileServer)
	if err != nil {
		p.i.Logger.Println("Failed to dial rpc server:", err)
		return "", err
	}
	defer s.Close()
	var arg lotsawa.CompileArgs = lotsawa.CompileArgs{code, "C"}
	var res lotsawa.CompileReply
	err = s.Compile(&arg, &res)

	if err != nil {
		p.i.Logger.Println("Failed to call rpc service:", err)
		return "", err
	}

	var issues string
	if res.C_Output != "" || res.C_Error != "" {
		issues = fmt.Sprintf("%s%s", res.C_Output, res.C_Error)
	}
	return issues, nil
}
