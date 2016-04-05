// Copyright 2016 Alex Fluter

package bot

import (
	"log"
	"sync"
)

type Module interface {
	Init() error
	Start() error
	Stop() error
	Status() string
	Run()
}

type EventModule interface {
	Module
	Handle(data interface{})
}

type BaseModule struct {
	Name   string
	State  ModState
	exitCh chan bool
	wait   sync.WaitGroup
	Logger *log.Logger
}
