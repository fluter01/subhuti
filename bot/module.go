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

type BaseModule struct {
	Name   string
	State  ModState
	exitCh chan bool
	wait   sync.WaitGroup
	Logger *log.Logger
}

var (
	// initModuleFuncs is a slice that contains the functions to create new modules.
	// The functions will be called when the bot is initialized.
	initModuleFuncs []func(*Bot) Module
)

// RegisterInitModuleFunc registers f to initModuleFuncs.
func RegisterInitModuleFunc(f func(*Bot) Module) {
	initModuleFuncs = append(initModuleFuncs, f)
}
