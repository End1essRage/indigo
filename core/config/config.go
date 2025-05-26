package config

import (
	"fmt"
	"os"

	yaml "github.com/goccy/go-yaml"
)

type YamlConfig struct {
	Bot          BotConfig      `yaml:"bot"`
	Cache        CacheConfig    `yaml:"cache"`
	Storage      StorageConfig  `yaml:"storage"`
	Media        MediaConfig    `yaml:"media"`
	Api          *ApiConfig     `yaml:"api,omitempty"`
	Commands     []Command      `yaml:"commands"`
	Keyboards    []Keyboard     `yaml:"keyboards,omitempty"`
	Forms        []Form         `yaml:"forms,omitempty"`
	Interceptors []Interceptor  `yaml:"interceptors,omitempty"`
	Modules      []ModuleConfig `yaml:"modules,omitempty"`
	Secrets      []Secret       `yaml:"secrets,omitempty"`
}

type Config struct {
	Bot          BotConfig
	HTTP         *ApiConfig
	Cache        CacheConfig
	Storage      StorageConfig
	Commands     map[string]*Command
	Keyboards    map[string]*Keyboard
	Forms        map[string]*Form
	Interceptors []Interceptor
	Modules      []ModuleConfig
	Secrets      []Secret
	Media        MediaConfig
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
	config.HTTP = yConfig.Api
	config.Storage = yConfig.Storage
	config.Cache = yConfig.Cache
	config.Interceptors = yConfig.Interceptors
	config.Modules = yConfig.Modules
	config.Secrets = yConfig.Secrets
	config.Media = yConfig.Media

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

// MEDIA
type MediaConfig struct {
	Type string `yaml:"type"`
}

// SECRETS
type Secret struct {
	Name string `yaml:"name"`
}

// BOT
type BotConfig struct {
	Mode       string `yaml:"mode"`
	AllowGroup bool   `yaml:"allow_group"`
	IsAdmin    bool   `yaml:"is_admin"`
	Debug      bool   `yaml:"debug"`
	Roles      bool   `yaml:"roles"`
}

type Command struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Script      *string `yaml:"script,omitempty"`
	Reply       *string `yaml:"reply,omitempty"`
	Keyboard    *string `yaml:"keyboard,omitempty"`
	Form        *string `yaml:"form,omitempty"`
	Role        *string `yaml:"role,omitempty"`
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

// Api
type ApiConfig struct {
	Address   string     `yaml:"address"`
	Endpoints []Endpoint `yaml:"endpoints"`
	Schemes   []Scheme   `yaml:"schemes"`
}

type Endpoint struct {
	Path   string  `yaml:"path"`
	Method string  `yaml:"method"`
	Scheme *string `yaml:"scheme,omitempty"`
	Script string  `yaml:"script"`
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

// STRUCTUAL
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

// DATA
type CacheConfig struct {
	Type  CacheType `yaml:"type"`
	Redis *struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis,omitempty"`
}

type StorageConfig struct {
	Type StorageType `yaml:"type"`
	File *struct {
		Path string `yaml:"path"`
	} `yaml:"file,omitempty"`
	Mongo *struct {
		Address  string `yaml:"address"`
		Login    string `yaml:"login"`
		Password string `yaml:"password"`
		Db       string `yaml:"db"`
	} `yaml:"mongo,omitempty"`
}

// HANDLING INTERCEPTORS
type AffectMode string

const AffectMode_All AffectMode = "all"
const AffectMode_Commands AffectMode = "commands"
const AffectMode_Text AffectMode = "text"
const AffectMode_Buttons AffectMode = "buttons"
const AffectMode_Media AffectMode = "media"
const AffectMode_Regex AffectMode = "regex"
const AffectMode_Url AffectMode = "url"
const AffectMode_Filter AffectMode = "filter"

type Module string

const TRACK_USER Module = "track_user"

type Interceptor struct {
	Affects AffectMode `yaml:"affects"`
	Scripts []string   `yaml:"scripts,omitempty"`
	Modules []string   `yaml:"modules,omitempty"`
}

type ModuleConfig struct {
	Name Module            `yaml:"name"`
	Cfg  map[string]string `yaml:"cfg"`
}
