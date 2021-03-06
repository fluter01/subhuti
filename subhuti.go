// Copyright 2016 Alex Fluter

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fluter01/subhuti/bot"
)

var (
	ircserver       string = "irc.freenode.net"
	port            int    = 7000
	ssl             bool   = true
	botnick         string = "candice"
	username        string = "gbot"
	realname        string = "Go BOT"
	identify_passwd string = "****"
)

var (
	cfg         string
	help        bool
	noproxy     bool
	logtostderr bool
)

func usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&cfg, "config", "bot.cfg", "bot configuration file")
	flag.BoolVar(&help, "help", false, "show help message")
	flag.BoolVar(&noproxy, "noproxy", false, "do not use proxy")
	flag.BoolVar(&noproxy, "np", false, "do not use proxy")
	flag.BoolVar(&logtostderr, "stderr", false, "log to stderr")
	flag.Parse()

	if help {
		usage()
		return
	}

	var err error
	var config *bot.BotConfig

	config = bot.NewConfig()
	err = config.Load(cfg)
	if err != nil {
		fmt.Println("config error")
		return
	}

	fmt.Println("Hello subhuti")
	fmt.Println("Running config:", config)

	if !noproxy && config.Proxy != "" {
		os.Setenv("HTTP_PROXY", config.Proxy)
	}

	if logtostderr {
		bot.NewLoggerFunc = bot.NewTestLogger
	}

	var b *bot.Bot

	b = bot.NewBot("GoBot", config)
	fmt.Println(b)

	b.Start()

	fmt.Println("GBot exiting")
}
