// Copyright 2016 Alex Fluter

package bot

// bot command
type CmdMap map[string]CmdFunc

type CmdFunc func(string) error

// maitains the comand map
func (bot *Bot) AddCommand(cmd string, f CmdFunc) {
	bot.cmdMap[cmd] = f
}

func (bot *Bot) DelCommand(cmd string) {
	delete(bot.cmdMap, cmd)
}

// bot command handlers
func (bot *Bot) onSave(string) error {
	return bot.config.Save()
}

func (bot *Bot) onShow(string) error {
	bot.Logger().Println("======== Config: ========")
	bot.Logger().Printf("%#v\n", bot.config)
	bot.Logger().Print("=========================")
	return nil
}

func (bot *Bot) onStatus(string) error {
	var mod Module
	bot.Logger().Print("======== Status: ========")
	for _, mod = range bot.modules {
		bot.Logger().Printf("Module %s %s", mod, mod.Status())
	}
	bot.Logger().Print("=========================")
	return nil
}

func (bot *Bot) onExit(string) error {
	var err error
	var mod Module
	for _, mod = range bot.modules {
		bot.Logger().Printf("Stopping module %s", mod)
		err = mod.Stop()
		if err != nil {
			bot.Logger().Printf("Stop module %s failed: %s", mod, err)
		}
	}
	close(bot.eventQ)
	return nil
}

func (bot *Bot) onConnect(string) error {
	return bot.IRC().connect()
}

func (bot *Bot) onDisconnect(string) error {
	bot.IRC().disconnect()
	return nil
}

func (bot *Bot) onReconnect(string) error {
	bot.IRC().disconnect()
	return bot.IRC().connect()
}
