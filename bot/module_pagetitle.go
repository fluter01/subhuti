// Copyright 2016 Alex Fluter

package bot

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	fnamePtn    = "filename=\"(.*)\""
	httpTimeout = 5
	maxRedirect = 3
)

var (
	fnameRe = regexp.MustCompile(fnamePtn)
)

type Pagetitle struct {
	BaseModule
	bot    *Bot
	client *http.Client
}

func init() {
	RegisterInitModuleFunc(NewPagetitle)
}

func NewPagetitle(bot *Bot) Module {
	t := new(Pagetitle)
	t.bot = bot
	t.Name = "Pagetitle"
	t.Logger = bot.Logger

	client := &http.Client{}
	client.Timeout = httpTimeout * time.Second
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirect {
			return errors.New(fmt.Sprintf(
				"stopped after %d redirects",
				maxRedirect),
			)
		}
		return nil
	}
	t.client = client

	return t
}

func (t *Pagetitle) Init() error {
	t.Logger.Println("Initializing Pagetitle")
	t.State = Initialized
	return nil
}

func (t *Pagetitle) Start() error {
	t.Logger.Println("Starting Pagetitle")
	t.bot.RegisterEventHandler(MessageParseEvent, t.parseMessage)
	t.State = Running
	return nil
}

func (t *Pagetitle) Stop() error {
	t.Logger.Println("Pagetitle stopped")
	t.State = Stopped
	return nil
}

func (t *Pagetitle) String() string {
	return t.Name
}

func (t *Pagetitle) Status() string {
	return t.State.String()
}

func (t *Pagetitle) Run() {
}

func (t *Pagetitle) parseMessage(data interface{}) {
	req, ok := data.(*MessageRequest)
	if !ok {
		return
	}

	if req.neturl == nil {
		return
	}

	if req.irc.config.IgnoreURLTitle(req.channel) {
		req.irc.Logger.Printf("ignoring url title for %s", req.channel)
		return
	}
	req.irc.Logger.Printf("getting page title")
	title := t.getTitle(req)
	if title != "" {
		var res string
		if !req.ischan {
			res = fmt.Sprintf("Your link: %s", title)
		} else {
			res = fmt.Sprintf("%s's link: %s", req.nick, title)
		}
		req.irc.sendReply(res, req)
	}

	return
}

func (t *Pagetitle) getTitle(req *MessageRequest) string {
	const maxTitleLen = 256
	var resp *http.Response
	var err error
	var result string

	// only proceed with text documents
	resp, err = t.client.Head(req.url)
	if err != nil {
		req.irc.Logger.Printf("Http Head error: %s", err)
		return result
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/") {
		req.irc.Logger.Println("Not text document, skipping download")
		// get file name if present
		disp := resp.Header.Get("Content-Disposition")
		m := fnameRe.FindStringSubmatch(disp)
		if len(m) == 2 && len(m[1]) > 0 {
			result = "filename: " + m[1]
		}
		return result
	}

	resp, err = t.client.Get(req.url)
	if err != nil {
		req.irc.Logger.Printf("Http Get error: %s", err)
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

	req.irc.Logger.Println("Returning title:", result)

	return result
}
