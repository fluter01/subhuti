package bot

/////////////////////////////////////////////////////////////////////////////
// default event handlers
func HandleUserJoin(data interface{}) {
	var join *UserJoinData

	join = data.(*UserJoinData)

	NewNopLogger(nil, "").Printf("%s joined %s", join.nick, join.channel)
}

func HandleUserPart(data interface{}) {
	var part *UserPartData

	part = data.(*UserPartData)

	NewNopLogger(nil, "").Printf("%s parted %s: %s", part.nick, part.channel, part.msg)
}

func HandleUserQuit(data interface{}) {
	var quit *UserQuitData

	quit = data.(*UserQuitData)

	NewNopLogger(nil, "").Printf("%s quited: %s", quit.nick, quit.msg)
}

func HandleUserNick(data interface{}) {
	var nick *UserNickData

	nick = data.(*UserNickData)

	NewNopLogger(nil, "").Printf("%s changed nick %s", nick.nick, nick.newNick)
}

func HandlePong(data interface{}) {
	var pong *PongData

	pong = data.(*PongData)

	NewNopLogger(nil, "").Printf("pong from %s: %s", pong.from, pong.origin)
}
