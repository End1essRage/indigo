package main

import (
	"log"
	"os"
	"path"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

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

	handler := NewHandler(le, bot, config)

	//получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tBot.GetUpdatesChan(u)

	logrus.Info("start processing")
	// обработка обновлений
	for update := range updates {
		handler.HandleUpdate(&update)
	}
}
