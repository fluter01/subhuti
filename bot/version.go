// Copyright 2016 Alex Fluter

package bot

import "fmt"

const (
	G = "Subhuti"
)

const (
	Major = 0
	Minor = 1
	Rel   = 1
)

func Version() string {
	return fmt.Sprintf("%s version %d.%d.%d",
		G, Major, Minor, Rel)
}

func Source() string {
	return "http://github.com/fluter01/subhuti"
}

func HandleVersion(*MessageRequest, string) (string, error) {
	return Version(), nil
}

func HandleSource(*MessageRequest, string) (string, error) {
	return Source(), nil
}
