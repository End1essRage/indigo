package main

import (
	"os"

	yaml "github.com/goccy/go-yaml"
)

type BotConfig struct {
	Token string `yaml:"token"`
	Mode  string `yaml:"mode"`
}

type Command struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Handler     *string `yaml:"handler,omitempty"`
	Reply       *Reply  `yaml:"reply,omitempty"`
}

type Reply struct {
	Msg      *string `yaml:"msg,omitempty"`
	Keyboard *string `yaml:"keyboard,omitempty"`
}

type Button struct {
	Name         string  `yaml:"name"`
	Text         string  `yaml:"text"`
	CallbackData *string `yaml:"callback_data,omitempty"`
	Handler      *string `yaml:"handler,omitempty"`
}

type KeyboardRow struct {
	Row []Button `yaml:"row"`
}

type Keyboard struct {
	Name    string        `yaml:"name"`
	Message string        `yaml:"message"`
	Type    string        `yaml:"type"`
	Buttons []KeyboardRow `yaml:"buttons"`
}

type Config struct {
	Bot       BotConfig  `yaml:"bot"`
	Commands  []Command  `yaml:"commands"`
	Keyboards []Keyboard `yaml:"keyboards,omitempty"`
}

// Config loader
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
