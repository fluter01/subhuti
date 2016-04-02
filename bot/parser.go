// Copyright 2016 Alex Fluter

package bot

import "errors"

// ErrNotParsed indicates that this parser does not parse
// this kind of request.
var ErrNotParsed = errors.New("Not parsed")

// Parser is the interface that provides the Parse method.
//
// Parse parses the message given in the MessageRequest, and
// returns an error if any.
// The bot is given as a way to access global bot information
// and operations, most parsers do not use this.
//
// If the message is not the kind that the parser could parse,
// then it should return ErrNotParsed, so that other parsers
// will proceed parsing the message.
type Parser interface {
	Parse(bot *Bot, req *MessageRequest) error
}
