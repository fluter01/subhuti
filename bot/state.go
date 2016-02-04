// Copyright 2016 Alex Fluter

package bot

type ModState int

const (
	_                    = iota
	Initialized ModState = iota
	Disconnected
	Connected
	Logged
	Identified
	Running
	Stopped
)

func (s ModState) String() string {
	switch s {
	case Initialized:
		return "Initialized "
	case Disconnected:
		return "Disconnected"
	case Connected:
		return "Connected   "
	case Logged:
		return "Logged      "
	case Identified:
		return "Identified  "
	case Running:
		return "Running     "
	case Stopped:
		return "Stopped     "
	default:
		return "Unknown     "
	}
}
