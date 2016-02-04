// Copyright 2016 Alex Fluter

package bot

import (
	"testing"
)

func getLogger() Logger {
	config := &BotConfig{Name: "TESTCONFIG", LogDir: "."}
	bot := NewBot("Test Bot", config)
	logger := NewStdFileLogger(bot, "test")
	return logger
}

func TestLogging1(t *testing.T) {
	log := getLogger()

	log.Print("Print from here")
	log.Println("Print after")
	log.Printf("%s\n", "Printf after")
}

func fun1(log Logger) {
	log.Print("log in func1")
}

func TestLogging2(t *testing.T) {
	log := getLogger()

	log.Print("log from here")
	fun1(log)
}

func TestLogging3(t *testing.T) {
	log := getLogger()

	log.Print("log first line")
	log.Print("log second line")
	log.Print("log third line")
}
