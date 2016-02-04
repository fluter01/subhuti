// Copyright 2016 Alex Fluter

package bot

import "fmt"
import "os"
import "io"

type NopLogger struct {
}

func NewNopLogger(bot *Bot, prefix string) Logger {
	return new(NopLogger)
}

// nop
func (logger *NopLogger) Output(calldepth int, s string) error {
	return nil
}

func (logger *NopLogger) Fatal(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (logger *NopLogger) Fatalf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (logger *NopLogger) Fatalln(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func (logger *NopLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *NopLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *NopLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	logger.Output(2, s)
	panic(s)
}

func (logger *NopLogger) Print(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
}

func (logger *NopLogger) Printf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
}

func (logger *NopLogger) Println(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
}

func (logger *NopLogger) Flags() int {
	return 0
}

func (logger *NopLogger) Prefix() string {
	return ""
}

func (logger *NopLogger) SetFlags(flag int) {
}

func (logger *NopLogger) SetPrefix(prefix string) {
}

func (logger *NopLogger) SetOutput(w io.Writer) {
}
