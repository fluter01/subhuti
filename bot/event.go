// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
)

// Priority defines the order of the invoke of the handlers.
// Handlers with High priority is invoke first.
type Priority int

const (
	High Priority = iota
	Low
)

type EventType int
type EventHandler func(interface{})
type EventHandlers *[2][]EventHandler

const (
	Input EventType = iota
	UserJoin
	UserPart
	UserQuit
	UserNick
	Pong
	PrivateMessage
	ChannelMessage
	Disconnect
	EventCount
	MessageParseEvent
)

func (evt EventType) String() string {
	var eventNames [EventCount]string = [...]string{
		"Input",
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
	return fmt.Sprintf("%d", evt)
}

type Event struct {
	evt  EventType
	data interface{}
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

type EventBase struct {
	bot  *Bot
	from string
	nick string
	user string
	host string
}

type UserJoinData struct {
	EventBase
	irc     *IRC
	channel string
}

type UserPartData struct {
	EventBase
	irc     *IRC
	channel string
	msg     string
}

type UserQuitData struct {
	EventBase
	irc *IRC
	msg string
}

type UserNickData struct {
	EventBase
	irc     *IRC
	newNick string
}

// PrivateMessage
type PrivateMessageData struct {
	EventBase
	irc  *IRC
	text string
}

// ChannelMessage
type ChannelMessageData struct {
	PrivateMessageData
	channel string
}

type PongData struct {
	bot    *Bot
	irc    *IRC
	from   string
	origin string
}
