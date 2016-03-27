// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
)

type EventType int

const (
	UserInput EventType = iota
	UserJoin
	UserPart
	UserQuit
	UserNick
	Pong
	PrivateMessage
	ChannelMessage
	Disconnect
	EventCount
)

func (evt EventType) String() string {
	var eventNames [EventCount]string = [...]string{
		"UserInput",
		"UserJoin",
		"UserPart",
		"UserQuit",
		"UserNick",
		"Pong",
		"PrivateMessage",
		"ChannelMessage",
		"Disconnect",
	}
	if evt < EventCount {
		return eventNames[evt]
	}
	return "Unknown"
}

type Event struct {
	evt  EventType
	data interface{}
}

type EventHandler func(interface{})
type EventHandlers []EventHandler

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

type EventBase struct {
	bot  *Bot
	from string
	nick string
	user string
	host string
}

type UserJoinData struct {
	EventBase
	channel string
}

type UserPartData struct {
	EventBase
	channel string
	msg     string
}

type UserQuitData struct {
	EventBase
	msg string
}

type UserNickData struct {
	EventBase
	newNick string
}

// PrivateMessage
type PrivateMessageData struct {
	EventBase
	text string
}

// ChannelMessage
type ChannelMessageData struct {
	PrivateMessageData
	channel string
}

type PongData struct {
	bot    *Bot
	from   string
	origin string
}
