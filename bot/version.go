// Copyright 2016 Alex Fluter

package bot

import "fmt"

const (
	G = "Subhuti"
)

const (
	Major = 0
	Minor = 3
	Rel   = 1
)

func Version() string {
	return fmt.Sprintf("%s version %d.%d.%d",
		G, Major, Minor, Rel)
}

func Source() string {
	return "http://github.com/fluter01/subhuti"
}
