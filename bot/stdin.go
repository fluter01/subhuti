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
	BaseModule

	bot   *Bot
	fd    *os.File
	state ModState
	wait  sync.WaitGroup
}

func NewStdin(bot *Bot) *Stdin {
	var stdin *Stdin

	stdin = new(Stdin)
	stdin.bot = bot
	stdin.fd = os.Stdin
	stdin.Logger = NewLoggerFunc(fmt.Sprintf("%s/%s-stdin",
		bot.config.LogDir, bot.Name))
	return stdin
}

func (stdin *Stdin) String() string {
	return "STDIN"
}

func (stdin *Stdin) Init() error {
	stdin.Logger.Printf("Initializing module %s", stdin)
	stdin.state = Initialized
	return nil
}

func (stdin *Stdin) Start() error {
	stdin.Logger.Printf("Starting module %s", stdin)
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
			stdin.Logger.Println("read error: ", err)
			break
		}
		line = string(buf[:n])
		stdin.Logger.Println("user input:", line)
		stdin.bot.AddEvent(NewEvent(Input, line))
	}
	stdin.Logger.Println("Stdin loop exiting")
	stdin.state = Stopped
	stdin.wait.Done()
}

func (stdin *Stdin) loop() {
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
		stdin.Logger.Println("user input:", line)
		stdin.bot.AddEvent(NewEvent(Input,
			line))
		//TODO: shortcut EXIT
		if line == fmt.Sprintf("%c%s", stdin.bot.config.GetTrigger(), "exit") {
			break
		}
	}

	if err = scanner.Err(); err != nil {
		stdin.Logger.Println("scanner error:", err)
	}
	stdin.Logger.Println("Stdin loop exited")
	stdin.state = Stopped
}

func (stdin *Stdin) Run() {
	stdin.wait.Add(1)
	go stdin.loop()
}

func (stdin *Stdin) Status() string {
	return fmt.Sprintf("%s", stdin.state)
}

func (stdin *Stdin) Stop() error {
	stdin.fd.Close()
	stdin.wait.Wait()
	return nil
}
