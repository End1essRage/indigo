package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	Token      string
	ConfigPath string
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	//parse env
	if err := godotenv.Load(); err != nil {
		logrus.Warning("error while reading environment", err.Error())
	}

	Token = os.Getenv("BOT_TOKEN")
	if Token == "" {
		logrus.Warn("cant set Token")
	}

	ConfigPath = os.Getenv("CONFIG_PATH")
	if ConfigPath == "" {
		logrus.Warn("cant set ConfigPath")
	}
}

func main() {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	//загружаем конфиг
	config, err := LoadConfig(path.Join(curDir, ConfigPath))
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logrus.Infof("config data: %v", config)

	if Token == "" {
		panic("no token provided")
	}

	// инициализация тг бота
	tBot, err := tgbotapi.NewBotAPI(Token)
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
	go func() {
		for update := range updates {
			handler.HandleUpdate(&update)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Server stopped")
}
