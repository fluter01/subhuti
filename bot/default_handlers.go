package bot

/////////////////////////////////////////////////////////////////////////////
// default event handlers
func HandleUserJoin(data interface{}) {
	var join *UserJoinData

	join = data.(*UserJoinData)

	NewLogger("").Printf("%s joined %s", join.nick, join.channel)
}

func HandleUserPart(data interface{}) {
	var part *UserPartData

	part = data.(*UserPartData)

	NewLogger("").Printf("%s parted %s: %s", part.nick, part.channel, part.msg)
}

func HandleUserQuit(data interface{}) {
	var quit *UserQuitData

	quit = data.(*UserQuitData)

	NewLogger("").Printf("%s quited: %s", quit.nick, quit.msg)
}

func HandleUserNick(data interface{}) {
	var nick *UserNickData

	nick = data.(*UserNickData)

	NewLogger("").Printf("%s changed nick %s", nick.nick, nick.newNick)
}

func HandlePong(data interface{}) {
	var pong *PongData

	pong = data.(*PongData)

	NewLogger("").Printf("pong from %s: %s", pong.from, pong.origin)
}
