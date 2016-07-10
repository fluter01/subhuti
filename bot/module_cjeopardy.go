// Copyright 2016 Alex Fluter

package bot

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	cjeopardy_file  = "cjeopardy.txt"
	cjeopardy_url   = "https://raw.githubusercontent.com/pragma-/pbot/master/modules/cjeopardy/cjeopardy.txt"
	cjeopardy_id_re = "^([0-9]+)\\) (.+)$"

	cjeopardy_modnick = "candide"
	cjeopardy_prefix  = "\x03\x31\x33Next question\x0f: "
	cjeopardy_key     = "!what"
)

type Cjeopardy struct {
	BaseModule
	bot     *Bot
	re      *regexp.Regexp
	db      map[int]string
	modnick string // nickname of the moderator
	prefix  string // prefix of the question
	key     string // keyword to prepend to the answer
}

func init() {
	RegisterInitModuleFunc(NewCjeopardy)
}

func NewCjeopardy(bot *Bot) Module {
	cj := new(Cjeopardy)
	cj.bot = bot
	cj.Name = "Cjeopardy"
	cj.Logger = bot.Logger
	cj.db = make(map[int]string)
	cj.re = regexp.MustCompile(cjeopardy_id_re)
	cj.modnick = cjeopardy_modnick
	cj.prefix = cjeopardy_prefix
	cj.key = cjeopardy_key

	return cj
}

func (cj *Cjeopardy) Init() error {
	cj.Logger.Println("Initializing Cjeopardy")
	cj.State = Initialized
	return nil
}

func (cj *Cjeopardy) Start() error {
	cj.Logger.Println("Starting Cjeopardy")

	var dbReader io.ReadCloser

	fpath := fmt.Sprintf("%s/%s", cj.bot.config.DataDir, cjeopardy_file)
	f, err := os.Open(fpath)
	if err != nil {
		cj.Logger.Println("CJeopardy: open", fpath, "failed:", err)
	} else {
		dbReader = f
	}

	if err := cj.load(dbReader); err != nil {
		return err
	}

	cj.bot.RegisterEventHandler(ChannelMessage, cj.handleMessage)
	cj.bot.foreachIRC(func(irc *IRC) {
		irc.interpreter.RegisterCommand("cjeopardy", cj.handleCommand)
	})
	cj.State = Running
	return nil
}

func (cj *Cjeopardy) load(r io.ReadCloser) error {
	if r == nil {
		return nil
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		m := cj.re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if len(m) < 2 {
			continue
		}
		n, err := strconv.ParseInt(m[1], 10, 0)
		if err != nil {
			continue
		}

		var pos int
		var pos2 int = len(m[2])
		for i, c := range m[2] {
			if c == '|' && m[2][i-1] != '\\' {
				pos = i
				break
			}
		}
		for i := pos + 1; i < len(m[2]); i++ {
			if (m[2][i] == '|' && m[2][i-1] != '\\') || m[2][i] == '{' {
				pos2 = i
				break
			}
		}
		ans := strings.Replace(m[2][pos+1:pos2], "\\|", "|", -1)

		cj.db[int(n)] = ans
	}
	cj.Logger.Printf("%d entries loaded into cjeopardy", len(cj.db))

	return nil
}

func (cj *Cjeopardy) Stop() error {
	cj.Logger.Println("Cjeopardy stopped")
	cj.State = Stopped
	cj.db = nil
	return nil
}

func (cj *Cjeopardy) String() string {
	return cj.Name
}

func (cj *Cjeopardy) Status() string {
	return cj.State.String()
}

func (cj *Cjeopardy) Run() {
}

func (cj *Cjeopardy) handleMessage(data interface{}) {

	msg, ok := data.(*ChannelMessageData)
	if !ok {
		cj.Logger.Println("corrupted channel message data")
		return
	}
	if msg.nick != cj.modnick {
		return
	}
	if !strings.HasPrefix(msg.text, cj.prefix) {
		return
	}
	question := strings.TrimPrefix(msg.text, cj.prefix)
	question = strings.TrimSpace(question)
	pos := strings.IndexByte(question, ')')
	if pos == -1 {
		return
	}
	if id, err := strconv.ParseInt(question[:pos], 10, 0); err == nil {
		time.AfterFunc(time.Duration(rand.Intn(15))*time.Second, func() {
			msg.irc.Privmsg(msg.channel, fmt.Sprintf("%s %s", cj.key, cj.db[int(id)]))
		})
		//cj.Logger.Println(msg.channel, ":", fmt.Sprintf("%s %s", cj.key, cj.db[int(id)]))
	}
}

func (cj *Cjeopardy) handleCommand(req *MessageRequest, args string) (string, error) {
	return "", nil
}
