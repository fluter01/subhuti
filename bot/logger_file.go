// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
	"io"
	stdLog "log"
	"os"
)

type FileLogger struct {
	l *stdLog.Logger
}

func NewFileLogger(bot *Bot, prefix string) *FileLogger {
	var logger *FileLogger

	logger = new(FileLogger)

	path := fmt.Sprintf("%s/GBot-%s.log.%s",
		bot.config.LogDir,
		prefix,
		bot.start.Format("2006-01-02"))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		panic(err)
	}
	logger.l = stdLog.New(f, "", stdLog.LstdFlags|stdLog.Lshortfile)

	return logger
}

func (logger *FileLogger) Output(calldepth int, s string) error {
	return logger.l.Output(calldepth+1, s)
}

func (logger *FileLogger) Fatal(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (logger *FileLogger) Fatalf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (logger *FileLogger) Fatalln(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func (logger *FileLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *FileLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *FileLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *FileLogger) Print(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
}

func (logger *FileLogger) Printf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
}

func (logger *FileLogger) Println(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
}

func (logger *FileLogger) Flags() int {
	return logger.l.Flags()
}

func (logger *FileLogger) Prefix() string {
	return logger.l.Prefix()
}

func (logger *FileLogger) SetFlags(flag int) {
	logger.l.SetFlags(flag)
}

func (logger *FileLogger) SetPrefix(prefix string) {
	logger.l.SetPrefix(prefix)
}

func (logger *FileLogger) SetOutput(w io.Writer) {
	logger.l.SetOutput(w)
}
