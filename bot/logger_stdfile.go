// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"io"
	stdLog "log"
	"os"
)

type StdFileLogger struct {
	FileLogger
}

func init() {
	stdLog.SetFlags(stdLog.LstdFlags | stdLog.Lshortfile)
}

func NewStdFileLogger(bot *Bot, prefix string) *StdFileLogger {
	var logger *StdFileLogger

	logger = new(StdFileLogger)
	logger.FileLogger = *NewFileLogger(bot, prefix)

	return logger
}

func (logger *StdFileLogger) Output(calldepth int, s string) error {
	stdLog.Output(calldepth+1, s)
	return logger.FileLogger.Output(calldepth+1, s)
}

func (logger *StdFileLogger) Fatal(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (logger *StdFileLogger) Fatalf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (logger *StdFileLogger) Fatalln(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func (logger *StdFileLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *StdFileLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *StdFileLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *StdFileLogger) Print(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
}

func (logger *StdFileLogger) Printf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
}

func (logger *StdFileLogger) Println(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
}

func (logger *StdFileLogger) Flags() int {
	return logger.l.Flags()
}

func (logger *StdFileLogger) Prefix() string {
	return logger.l.Prefix()
}

func (logger *StdFileLogger) SetFlags(flag int) {
	logger.l.SetFlags(flag)
}

func (logger *StdFileLogger) SetPrefix(prefix string) {
	logger.l.SetPrefix(prefix)
}

func (logger *StdFileLogger) SetOutput(w io.Writer) {
	logger.l.SetOutput(w)
}
