package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	yaml "github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
)

type BotConfig struct {
	Token string `yaml:"token"`
	Mode  string `yaml:"mode"`
}

type Command struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Handler     string `yaml:"handler"`
}

type Config struct {
	Bot      BotConfig `yaml:"bot"`
	Commands []Command `yaml:"commands"`
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

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	//загружаем конфиг
	config, err := LoadConfig(path.Join(curDir, "config", "config.yaml"))
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logrus.Infof("config data: %v", config)

	// инициализация тг бота
	tBot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		logrus.Fatal(err)
	}

	tBot.Debug = true
	logrus.Infof("Authorized on account %s", tBot.Self.UserName)

	//обертка над тг ботом
	bot := NewBot(tBot)

	le := NewLuaEngine(bot)

	//получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tBot.GetUpdatesChan(u)

	logrus.Info("start processing")
	// обработка обновлений
	for update := range updates {
		//отбрасываем с пустым сообщением(надо будет убрать для обработки кнопок)
		if update.Message == nil {
			continue
		}

		//отбрасываем все кроме команд
		if !update.Message.IsCommand() {
			continue
		}

		//TODO возможность перезаписать через yml
		//обрабатываем help
		if update.Message.Command() == "help" {
			bot.SendMessage(update.Message.Chat.ID, formatHelpMessage(config.Commands))
		}

		// поиск команды
		for _, cmd := range config.Commands {
			if update.Message.Command() == cmd.Name {
				scriptPath := fmt.Sprintf("scripts/%s", cmd.Handler)
				if err := le.ExecuteScript(scriptPath, FromTgUpdateToLuaContext(update)); err != nil {
					logrus.Errorf("Error executing script: %v", err)
				}
			}
		}
	}
}

func formatHelpMessage(cmds []Command) string {
	sb := strings.Builder{}
	for _, c := range cmds {
		sb.WriteString(fmt.Sprintf("%s - %s \n", c.Name, c.Description))
	}
	return sb.String()
}
