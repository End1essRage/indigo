package config

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
	Reply       *string `yaml:"reply,omitempty"`
	Keyboard    *string `yaml:"keyboard,omitempty"`
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
	Script  *string        `yaml:"script,omitempty"`
	Message *string        `yaml:"message,omitempty"`
	Buttons *[]KeyboardRow `yaml:"buttons,omitempty"`
}

type YamlConfig struct {
	Bot       BotConfig   `yaml:"bot"`
	HTTP      *HTTPConfig `yaml:"http,omitempty"`
	Commands  []Command   `yaml:"commands"`
	Keyboards []Keyboard  `yaml:"keyboards,omitempty"`
}

type HTTPConfig struct {
	Address   string     `yaml:"address"`
	Endpoints []Endpoint `yaml:"endpoints"`
	Schemes   []Scheme   `yaml:"schemes"`
}

type Endpoint struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
	Scheme string `yaml:"scheme"`
	Script string `yaml:"script"`
}

type Scheme struct {
	Name   string  `yaml:"name"`
	Fields []Field `yaml:"fields"`
}

type Field struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // string, number, boolean
	Required bool   `yaml:"required"`
	Source   string `yaml:"source"` // body, query, header
}

type Config struct {
	Bot       BotConfig
	HTTP      *HTTPConfig
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
	config.HTTP = yConfig.HTTP

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
