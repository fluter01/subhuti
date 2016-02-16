// Copyright 2016 Alex Fluter

package bot

import "encoding/json"
import "fmt"
import "os"
import "bytes"
import "log"

type BotConfig struct {
	path            string
	Proxy           string
	Name            string
	IrcServer       string
	Port            int
	Ssl             bool
	BotNick         string
	Username        string
	RealName        string
	Identify_passwd string

	HomeDir string
	LogDir  string

	IRCTrigger     byte
	ChannelTrigger byte
	BotTrigger     byte

	AutoConnect bool

	Channels []struct {
		Name           string
		Trigger        byte
		IgnoreURLTitle bool
	}

	RawLogging bool

	CompileServer string
}

func (config *BotConfig) String() string {
	return fmt.Sprintf("%s:%d %s %s %s",
		config.IrcServer,
		config.Port,
		config.BotNick,
		config.Username,
		config.RealName)
}

func NewConfig(path string) *BotConfig {
	return &BotConfig{path: path}
}

func (config *BotConfig) Save() error {
	return SaveToFile(config, config.path)
}

func (config *BotConfig) Load() error {
	nc, err := LoadFromFile(config.path)
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

func (config *BotConfig) Trigger(channel string) string {
	var c byte

	for _, ch := range config.Channels {
		if ch.Name == channel {
			c = ch.Trigger
			break
		}
	}
	if c == 0 {
		c = config.ChannelTrigger
	}
	return string(c)
}

func (config *BotConfig) IgnoreURLTitle(channel string) bool {
	var ignore bool = false

	for _, ch := range config.Channels {
		if ch.Name == channel {
			ignore = ch.IgnoreURLTitle
			break
		}
	}
	return ignore
}
