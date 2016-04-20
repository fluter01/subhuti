package bot

import (
	"fmt"
	"net/url"
)

// Message data for intepret
type MessageRequest struct {
	irc       *IRC
	ischan    bool
	from      string
	nick      string
	user      string
	host      string
	channel   string
	text      string
	direct    bool
	url       string
	neturl    *url.URL
	prefix    bool
	keyword   string
	arguments string
}

func (req *MessageRequest) String() string {
	if req.ischan {
		return fmt.Sprintf("--> %s -> %s] %s",
			req.from,
			req.channel,
			req.text)
	} else {
		return fmt.Sprintf("--> %s] %s",
			req.from,
			req.text)
	}
}

func (req *MessageRequest) cleanURL() {
	req.url = ""
	req.neturl = nil
}

type Command func(*MessageRequest, string) (string, error)

func VersionCommand(*MessageRequest, string) (string, error) {
	return Version(), nil
}

func SourceCommand(*MessageRequest, string) (string, error) {
	return Source(), nil
}
