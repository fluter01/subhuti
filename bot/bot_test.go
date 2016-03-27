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
	bot := NewBot("tset", config)
	t.Log(bot)
}
