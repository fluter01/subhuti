// Copyright 2016 Alex Fluter

package bot

type Module interface {
	Init() error
	Start() error
	Stop() error
	Loop()
	Run()
	Status() string

	Logger() Logger
}
