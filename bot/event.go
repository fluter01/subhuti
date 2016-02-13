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

	EventCount
)

var EventNames [EventCount]string = [...]string{
	"UserInput",
	"UserJoin",
	"UserPart",
	"UserQuit",
	"UserNick",
	"Pong",
	"PrivateMessage",
	"ChannelMessage",
}

type Event struct {
	evt  EventType
	data interface{}
}

type EventHandler func(interface{})
type EventHandlers []EventHandler

func (evt EventType) String() string {
	if evt < EventCount {
		return EventNames[evt]
	}
	return "Unknown"
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

// UserJoin
type UserJoinData struct {
	from, nick, user, host string
	channel string
}

type UserPartData struct {
	from, nick, user, host string
	channel string
	msg string
}

type UserQuitData struct {
	from, nick, user, host string
	msg string
}

type UserNickData struct {
	from, nick, user, host string
	newNick string
}

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
