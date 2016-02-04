// Copyright 2016 Alex Fluter

package bot

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Stdin struct {
	Module

	logger Logger

	fd    *os.File
	bot   *Bot
	state ModState
	wait  sync.WaitGroup
}

func NewStdin(bot *Bot) *Stdin {
	var stdin *Stdin

	stdin = new(Stdin)
	stdin.bot = bot
	stdin.fd = os.Stdin
	stdin.logger = NewStdFileLogger(bot, "stdin")
	return stdin
}

func (stdin *Stdin) String() string {
	return "STDIN"
}

func (stdin *Stdin) Init() error {
	stdin.Logger().Printf("Initializing module %s", stdin)
	stdin.state = Initialized
	return nil
}

func (stdin *Stdin) Start() error {
	stdin.Logger().Printf("Starting module %s", stdin)
	stdin.state = Running
	return nil
}

func (stdin *Stdin) Loop2() {
	var err error
	var buf []byte
	var n int
	var line string

	buf = make([]byte, 1024)

	for {
		n, err = stdin.fd.Read(buf)
		fmt.Println(n, err)
		if err != nil {
			stdin.Logger().Println("read error: ", err)
			break
		}
		line = string(buf[:n])
		stdin.Logger().Println("user input:", line)
		stdin.bot.AddEvent(NewEvent(UserInput, line))
	}
	stdin.Logger().Println("Stdin loop exiting")
	stdin.state = Stopped
	stdin.wait.Done()
}

func (stdin *Stdin) Loop() {
	var err error
	var scanner *bufio.Scanner
	var line string
	defer stdin.wait.Done()

	scanner = bufio.NewScanner(stdin.fd)
	for scanner.Scan() {
		line = scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		stdin.Logger().Println("user input:", line)
		stdin.bot.AddEvent(NewEvent(UserInput,
			line))
		//TODO: shortcut EXIT
		if line == fmt.Sprintf("%c%s", stdin.bot.config.BotTrigger, "exit") {
			break
		}
	}

	if err = scanner.Err(); err != nil {
		stdin.Logger().Println("scanner error:", err)
	}
	stdin.Logger().Println("Stdin loop exiting")
	stdin.state = Stopped
}

func (stdin *Stdin) Run() {
	stdin.wait.Add(1)
	go stdin.Loop()
}

func (stdin *Stdin) Status() string {
	return fmt.Sprintf("%s", stdin.state)
}

func (stdin *Stdin) Stop() error {
	stdin.fd.Close()
	stdin.wait.Wait()
	return nil
}

func (stdin *Stdin) Logger() Logger {
	return stdin.logger
}
