// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/api/youtube/v3"
)

// API key for youtube client
type APIKey struct {
	key string
}

func (k *APIKey) Get() (string, string) {
	return "key", k.key
}

type Youtube struct {
	BaseModule
	bot    *Bot
	client *http.Client
	key    *APIKey
}

func init() {
	RegisterInitModuleFunc(NewYoutube)
}

func NewYoutube(bot *Bot) Module {
	yt := new(Youtube)
	yt.bot = bot
	yt.Name = "Youtube"
	yt.Logger = bot.Logger

	if bot.config.YoutubeAPIKey != "" {
		yt.key = &APIKey{bot.config.YoutubeAPIKey}
		yt.client = &http.Client{}
	} else {
		yt.key = nil
		yt.client = nil
	}
	return yt
}

func (yt *Youtube) Init() error {
	yt.Logger.Println("Initializing Youtube")
	yt.State = Initialized
	return nil
}

func (yt *Youtube) Start() error {
	yt.Logger.Println("Starting Youtube")
	if yt.key != nil {
		yt.bot.RegisterEventHandler(MessageParseEvent, yt.parseMessage)
	}
	yt.State = Running
	return nil
}

func (yt *Youtube) Stop() error {
	yt.Logger.Println("Youtube stopped")
	yt.State = Stopped
	return nil
}

func (yt *Youtube) String() string {
	return yt.Name
}

func (yt *Youtube) Status() string {
	return yt.State.String()
}

func (yt *Youtube) Run() {
}

func (yt *Youtube) parseMessage(data interface{}) {
	req, ok := data.(*MessageRequest)
	if !ok {
		return
	}

	if req.neturl == nil {
		return
	}

	host := req.neturl.Host
	if !(host == "youtube.com" || host == "www.youtube.com" || host == "youtu.be") {
		return
	}

	id := yt.getId(req.neturl)
	if id == "" {
		req.irc.Logger.Println("cannot parse id")
		return
	}
	youtube, err := youtube.New(&http.Client{})
	if err != nil {
		req.irc.Logger.Println("youtube new service:", err)
		return
	}

	opt := yt.key
	call := youtube.Videos.List("snippet,contentDetails,statistics").Id(id)
	rsp, err := call.Do(opt)
	if err != nil {
		req.irc.Logger.Println("youtube call error:", err)
		return
	}
	if len(rsp.Items) == 0 {
		req.irc.Logger.Printf("list of %s returned no result", id)
		return
	}

	var res string
	v := rsp.Items[0]
	s := v.Snippet
	c := v.ContentDetails
	t := v.Statistics

	if s == nil {
		req.irc.Logger.Printf("no snippet returned")
		return
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

	req.irc.sendReply(res, req)
	req.cleanURL()

	return
}

func (yt *Youtube) getId(u *url.URL) string {
	switch u.Host {
	case "youtu.be":
		return strings.TrimPrefix(u.Path, "/")
	case "youtube.com", "www.youtube.com":
		return u.Query().Get("v")
	default:
		return ""
	}
}
