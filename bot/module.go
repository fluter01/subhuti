// Copyright 2016 Alex Fluter

package bot

import "log"

type Module interface {
	Init() error
	Start() error
	Stop() error
	Loop()
	Run()
	Status() string
}

type BaseModule struct {
	Name   string
	Logger *log.Logger
}
