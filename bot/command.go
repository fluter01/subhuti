// Copyright 2016 Alex Fluter

package bot

// bot command
type CmdMap map[string]CmdFunc

type CmdFunc func(string) error
