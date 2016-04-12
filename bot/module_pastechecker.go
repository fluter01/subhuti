// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"

	"github.com/fluter01/lotsawa"
	"github.com/fluter01/paste"
)

type CodePasteChecker struct {
	BaseModule
	bot *Bot
	cs  *lotsawa.CompileServiceStub
}

func init() {
	RegisterInitModuleFunc(NewCodePasteChecker)
}

func NewCodePasteChecker(bot *Bot) Module {
	cp := new(CodePasteChecker)
	cp.bot = bot
	cp.Name = "CodePasteChecker"
	cp.Logger = bot.Logger

	return cp
}

func (cp *CodePasteChecker) Init() error {
	cp.Logger.Println("Initializing CodePasteChecker")
	cp.State = Initialized
	return nil
}

func (cp *CodePasteChecker) Start() error {
	cp.Logger.Println("Starting CodePasteChecker")
	if cp.bot.config.CompileServer != "" {
		cs, err := lotsawa.NewCompileServiceStub("tcp", cp.bot.config.CompileServer)
		if err != nil {
			cp.bot.Logger.Println("Failed to dial rpc server:", err)
			cp.cs = nil
		} else {
			cp.cs = cs
			cp.bot.RegisterEventHandler(MessageParseEvent, cp.handleMessage)
		}
	}
	cp.State = Running
	return nil
}

func (cp *CodePasteChecker) Stop() error {
	cp.Logger.Println("CodePasteChecker stopped")
	cp.State = Stopped
	if cp.cs != nil {
		cp.cs.Close()
	}
	return nil
}

func (cp *CodePasteChecker) String() string {
	return cp.Name
}

func (cp *CodePasteChecker) Status() string {
	return cp.State.String()
}

func (cp *CodePasteChecker) Run() {
}

func (cp *CodePasteChecker) handleMessage(data interface{}) {
	req, ok := data.(*MessageRequest)
	if !ok {
		return
	}

	if req.neturl == nil {
		return
	}

	code, err := paste.Get(req.url)
	if err != nil {
		if err != paste.ErrNotSupported {
			cp.Logger.Println("Get paste error:", err)
		}
		return
	}

	lang := req.irc.config.ChannelLang(req.channel)
	if lang == "" {
		cp.Logger.Println("No language defined for channel", req.channel)
		return
	}
	cmplres, have_issues := cp.processCode(code, lang)
	if cmplres == "" {
		return
	}
	res := fmt.Sprintf("%s's paste: %s", req.nick, cmplres)
	if have_issues {
		res += fmt.Sprintf(" -- issues found, please address them first!")
		req.irc.sendReply(res, req)
	} else {
		// do not repaste from sprunge
		if req.neturl.Host == "sprunge.us" {
			return
		}
		if req.irc.config.ChannelRepaste(req.channel) {
			req.irc.sendReply(res, req)
		}
	}
	req.cleanURL()
}

func (cp *CodePasteChecker) processCode(code string, lang string) (string, bool) {
	var err error
	var issues string
	var with_issues bool

	var arg lotsawa.CompileArgs = lotsawa.CompileArgs{code, lang}
	var res lotsawa.CompileReply
	err = cp.cs.Compile(&arg, &res)

	if err != nil {
		cp.Logger.Println("Failed to call rpc service:", err)
		return "", false
	}

	if res.C_Output != "" || res.C_Error != "" {
		issues = fmt.Sprintf("%s%s", res.C_Output, res.C_Error)
		with_issues = true
	}

	if !with_issues {
		id, err := paste.Paste(code)
		if err != nil {
			return "", false
		}
		return id, false
	}

	data := fmt.Sprintf("%s\n\n"+
		"----------------------------------------------------------------\n"+
		"%s", code, issues)
	id, err := paste.Paste(data)
	if err != nil {
		return "", false
	}

	return id, true
}
