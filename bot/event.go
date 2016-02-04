// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
)

type EventType int

const (
	UserInput EventType = iota

	PrivateMessage

	ChannelMessage

	EventCount
)

type Event struct {
	evt  EventType
	data interface{}
}

type EventHandler func(interface{})

func (evt EventType) String() string {
	switch evt {
	case UserInput:
		return "UserInput"
	case PrivateMessage:
		return "PrivateMessage"
	case ChannelMessage:
		return "ChannelMessage"
	default:
		return "Unknown"
	}
}

func (event Event) String() string {
	return fmt.Sprintf("Type %s, data [%s]",
		event.evt,
		event.data)
}

func NewEvent(typ EventType, data interface{}) *Event {
	return &Event{evt: typ,
		data: data}
}

// event data

// PrivateMessage
type PrivateMessageData struct {
	from, nick, user, host, text string
}

func NewPrivateMessageData(from, nick, user, host, text string) *PrivateMessageData {
	return &PrivateMessageData{from, nick, user, host, text}
}

// ChannelMessage
type ChannelMessageData struct {
	PrivateMessageData
	channel string
}

func NewChannelMessageData(from, nick, user, host, text, channel string) *ChannelMessageData {
	return &ChannelMessageData{PrivateMessageData{from, nick, user, host, text}, channel}
}
