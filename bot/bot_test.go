// Copyright 2016 Alex Fluter

package bot

import (
	_ "fmt"
	_ "strings"
	"testing"
	_ "time"
)

func TestBot(t *testing.T) {
	config := &BotConfig{}
	bot := NewBot("test", config)
	t.Log(bot)
	go func() {
		bot.Start()
	}()
	for bot.State != Running {
	}
	t.Log(bot)
	bot.Stop()
	t.Log(bot)
	if bot.State != Stopped {
		t.Fail()
	}
}

func TestBotEvent(t *testing.T) {
	config := &BotConfig{}
	bot := NewBot("test", config)
	go func() {
		bot.Start()
	}()
	for bot.State != Running {
	}

	//unknown command
	bot.AddEvent(NewEvent(Input, "/foo"))

	bot.AddEvent(NewEvent(Input, "/status"))

	const evt = 255
	//unknown event
	bot.AddEvent(NewEvent(evt, nil))

	//new event
	var hit bool
	bot.RegisterEventHandler(evt, func(interface{}) {
		hit = true
		t.Log("event handled")
	})
	bot.AddEvent(NewEvent(evt, nil))

	bot.Stop()
	if bot.State != Stopped {
		t.Fail()
	}

	if !hit {
		t.Fail()
	}
}
