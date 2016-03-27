// Copyright 2016 Alex Fluter

package bot

import (
	"io"
	"log"
	"os"
	"runtime"
)

func NewLogger(name string) *log.Logger {
	var f io.Writer
	var err error

	if name == "" {
		if runtime.GOOS == "windows" {
			f, err = os.OpenFile("NUL", os.O_RDWR, 0664)
		} else {
			f, err = os.OpenFile("/dev/null", os.O_RDWR, 0664)
		}
	} else if name == "stdout" {
		f = os.Stdout
	} else if name == "stderr" {
		f = os.Stderr
	} else {
		f, err = os.OpenFile(name,
			os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC,
			0664)
	}
	if err != nil {
		panic(err)
	}
	return log.New(f, "", log.LstdFlags|log.Lshortfile)
}
