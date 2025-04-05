package main

import (
	"os"

	yaml "github.com/goccy/go-yaml"
)

type BotConfig struct {
	Mode string `yaml:"mode"`
}

type Command struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Script      *string `yaml:"script,omitempty"`
	Reply       *Reply  `yaml:"reply,omitempty"`
}

type Reply struct {
	Msg      *string `yaml:"msg,omitempty"`
	Keyboard *string `yaml:"keyboard,omitempty"`
}

type Button struct {
	Name   string  `yaml:"name"`
	Text   string  `yaml:"text"`
	Data   *string `yaml:"data,omitempty"`
	Script *string `yaml:"script,omitempty"`
}

type KeyboardRow struct {
	Row []Button `yaml:"row"`
}

type Keyboard struct {
	Name    string         `yaml:"name"`
	Type    string         `yaml:"type"`
	Script  *string        `yaml:"script,omitempty"`
	Message *string        `yaml:"message,omitempty"`
	Buttons *[]KeyboardRow `yaml:"buttons,omitempty"`
}

type YamlConfig struct {
	Bot       BotConfig  `yaml:"bot"`
	Commands  []Command  `yaml:"commands"`
	Keyboards []Keyboard `yaml:"keyboards,omitempty"`
}

type Config struct {
	Bot       BotConfig
	Commands  map[string]*Command
	Keyboards map[string]*Keyboard
}

// Config loader
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var yConfig YamlConfig
	if err := yaml.Unmarshal(data, &yConfig); err != nil {
		return nil, err
	}

	var config Config
	config.Bot = yConfig.Bot
	//fill commands
	config.Commands = make(map[string]*Command)
	for _, c := range yConfig.Commands {
		config.Commands[c.Name] = &c
	}

	//fill keyboards
	config.Keyboards = make(map[string]*Keyboard)
	for _, k := range yConfig.Keyboards {
		config.Keyboards[k.Name] = &k
	}

	return &config, nil
}
