// Copyright 2016 Alex Fluter

package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	DefaultBotTrigger     = '/'
	DefaultChannelTrigger = '!'
	DefaultChannelLang    = "C"
)

type ChannelConfig struct {
	Name           string
	Trigger        byte
	IgnoreURLTitle bool
	Lang           string
	Repaste        bool
}

type IRCConfig struct {
	Name            string
	Server          string
	Port            int
	Ssl             bool
	BotNick         string
	Username        string
	RealName        string
	Identify_passwd string
	Trigger         byte
	RawLogging      bool
	AutoConnect     bool
	DebugMode       bool
	RedirectTo      string
	Channels        []*ChannelConfig
}

type BotConfig struct {
	path          string
	Proxy         string
	HomeDir       string
	LogDir        string
	DataDir       string
	DB            string
	Trigger       byte
	CompileServer string
	YoutubeAPIKey string
	IRC           []*IRCConfig
}

func (config *IRCConfig) String() string {
	return fmt.Sprintf("%s %s:%d %s",
		config.Name,
		config.Server, config.Port, config.BotNick)
}

func (config *BotConfig) String() string {
	return fmt.Sprintf("%s:%s",
		config.path,
		config.IRC)
}

func NewConfig() *BotConfig {
	return new(BotConfig)
}

func (config *BotConfig) Save(path string) error {
	return SaveToFile(config, path)
}

func (config *BotConfig) Load(path string) error {
	nc, err := LoadFromFile(path)
	if err != nil {
		return err
	}
	*config = *nc
	if config.HomeDir == "" {
		config.HomeDir, _ = os.Getwd()
	}
	if config.LogDir == "" {
		config.LogDir = config.HomeDir + "/log"
		if _, err := os.Stat(config.LogDir); os.IsNotExist(err) {
			if err := os.Mkdir(config.LogDir, os.ModeDir); err != nil {
				panic(err)
			}
		}
	}
	return err
}

func LoadFromFile(path string) (*BotConfig, error) {
	var err error
	var file *os.File
	var dec *json.Decoder
	var config *BotConfig

	file, err = os.Open(path)
	if err != nil {
		log.Printf("Failed to open file %s: %s", path, err)
		return nil, err
	}
	defer file.Close()

	dec = json.NewDecoder(file)
	config = new(BotConfig)
	err = dec.Decode(config)
	if err != nil {
		log.Print("Failed to read config:", err)
		return nil, err
	}
	config.path = path

	return config, nil
}

func SaveToFile(config *BotConfig, path string) error {
	var err error
	var file *os.File
	var enc *json.Encoder
	var buf bytes.Buffer

	file, err = os.Create(path)
	if err != nil {
		log.Printf("Failed to create file %s: %s", path, err)
		return err
	}
	defer file.Close()
	enc = json.NewEncoder(&buf)
	err = enc.Encode(config)
	if err != nil {
		log.Print("Failed to write config:", err)
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, buf.Bytes(), "", "\t")
	out.WriteTo(file)

	return nil
}

func (config *BotConfig) GetTrigger() byte {
	c := config.Trigger
	if c == 0 {
		c = DefaultBotTrigger
	}
	return c
}

func (config *BotConfig) GetIRC(server string) *IRCConfig {
	for i := range config.IRC {
		if config.IRC[i].Server == server {
			return config.IRC[i]
		}
	}
	return nil
}

func (config *IRCConfig) GetTrigger(channel string) string {
	var c byte

	for _, ch := range config.Channels {
		if ch.Name == channel {
			c = ch.Trigger
			break
		}
	}
	if c == 0 {
		c = config.Trigger
	}
	if c == 0 {
		c = DefaultChannelTrigger
	}
	return string(c)
}

func (config *IRCConfig) IgnoreURLTitle(channel string) bool {
	var ignore bool = false

	for _, ch := range config.Channels {
		if ch.Name == channel {
			ignore = ch.IgnoreURLTitle
			break
		}
	}
	return ignore
}

func (config *IRCConfig) ChannelRepaste(channel string) bool {
	for _, ch := range config.Channels {
		if ch.Name == channel {
			return ch.Repaste
		}
	}
	return false
}

func (config *IRCConfig) ChannelLang(channel string) string {
	var lang string
	for _, ch := range config.Channels {
		if ch.Name == channel {
			lang = ch.Lang
		}
	}
	if lang == "" {
		lang = DefaultChannelLang
	}
	return lang
}
