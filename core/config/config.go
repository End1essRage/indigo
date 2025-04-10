package config

import (
	"fmt"
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
	Form        *string `yaml:"form,omitempty"`
}

type Button struct {
	Text   string  `yaml:"text"`
	Data   *string `yaml:"data,omitempty"`
	Script *string `yaml:"script,omitempty"`
}

type KeyboardRow struct {
	Row []Button `yaml:"row"`
}

type Keyboard struct {
	Name    string         `yaml:"name"`
	Message *string        `yaml:"message,omitempty"`
	Buttons *[]KeyboardRow `yaml:"buttons,omitempty"`
}

type YamlConfig struct {
	Bot       BotConfig     `yaml:"bot"`
	Cache     CacheConfig   `yaml:"cache"`
	Storage   StorageConfig `yaml:"storage"`
	HTTP      *HTTPConfig   `yaml:"http,omitempty"`
	Commands  []Command     `yaml:"commands"`
	Keyboards []Keyboard    `yaml:"keyboards,omitempty"`
	Forms     []Form        `yaml:"forms,omitempty"`
}

type CacheConfig struct {
	Type  string `yaml:"type"`
	Redis *struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis,omitempty"`
}

type Form struct {
	Name        string      `yaml:"name"`
	Description *string     `yaml:"description,omitempty"`
	Stages      []FormStage `yaml:"stages"`
	Script      string      `yaml:"script"`
}

type FormStage struct {
	Field      string          `yaml:"field"`
	Message    string          `yaml:"message"`
	Validation *map[string]any `yaml:"validation,omitempty"`
	Keyboard   *string         `yaml:"keyboard,omitempty"`
	Script     *string         `yaml:"script,omitempty"`
}

type StorageConfig struct {
	Type string `yaml:"type"`
	File *struct {
		Path string `yaml:"path"`
	} `yaml:"file,omitempty"`
	Mongo *struct {
		Uri string `yaml:"uri"`
		Db  string `yaml:"db"`
	} `yaml:"mongo,omitempty"`
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
	Cache     CacheConfig
	Storage   StorageConfig
	Commands  map[string]*Command
	Keyboards map[string]*Keyboard
	Forms     map[string]*Form
}

type ValidationErr error

// Config loader
func LoadConfig(path string, validate bool) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var yConfig YamlConfig
	if err := yaml.Unmarshal(data, &yConfig); err != nil {
		return nil, err
	}

	if validate {
		ok, desc := Validate(&yConfig)
		if !ok {
			return nil, fmt.Errorf("ошибка валидации: %s", desc)
		}
	}

	var config Config
	config.Bot = yConfig.Bot
	config.HTTP = yConfig.HTTP
	config.Storage = yConfig.Storage
	config.Cache = yConfig.Cache

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

	//fill forms
	config.Forms = make(map[string]*Form)
	for _, f := range yConfig.Forms {
		config.Forms[f.Name] = &f
	}

	return &config, nil
}
