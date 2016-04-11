// Copyright 2016 Alex Fluter

package bot

import (
	"fmt"
)

type EventType int
type EventHandler func(interface{})
type EventHandlers []EventHandler

var (
	// EventMap is the global map where each event handlers register itself.
	eventMap = make(map[EventType]EventHandlers)
)

// RegisterEventHandler add event handler to the event map.
func RegisterEventHandler(typ EventType, h EventHandler) {
	eventMap[typ] = append(eventMap[typ], h)
}

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
