package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	b "github.com/end1essrage/indigo-core/bot"
	ca "github.com/end1essrage/indigo-core/cache"
	"github.com/end1essrage/indigo-core/client"
	c "github.com/end1essrage/indigo-core/config"
	l "github.com/end1essrage/indigo-core/lua"
	s "github.com/end1essrage/indigo-core/server"
	st "github.com/end1essrage/indigo-core/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	Token       string
	ConfigPath  string
	ScriptsPath string
)

const (
	defaultConfigPath  = "/config/config.yaml"
	defaultScriptsPath = "/app/scripts"
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
		logrus.Warn("cant set ConfigPath, setting to default")
		ConfigPath = defaultConfigPath
	}

	ScriptsPath = os.Getenv("SCRIPTS_PATH")
	if ScriptsPath == "" {
		logrus.Warn("cant set ScriptsPath, setting to default")
		ScriptsPath = defaultScriptsPath
	}
}

func main() {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	//загружаем конфиг
	config, err := c.LoadConfig(path.Join(curDir, ConfigPath))
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

	tBot.Debug = false
	logrus.Infof("Authorized on account %s", tBot.Self.UserName)

	//обертка над тг ботом
	bot := b.NewBot(tBot)

	//buffer
	buffer := ca.NewInMemoryCache(5 * time.Minute)

	//кэш
	var cache l.Cache
	switch config.Cache.Type {
	case "redis":
		redis, err := ca.NewRedisCache(config.Cache.Redis.Address, config.Cache.Redis.Password, config.Cache.Redis.DB)
		if err != nil {
			panic(err)
		}
		cache = redis
	default:
		cache = buffer
	}

	//хранилище
	var storage l.Storage
	switch config.Storage.Type {
	case "mongo":
		storage, err = st.NewMongoStorage(config.Storage.Mongo.Uri, config.Storage.Mongo.Db)
		if err != nil {
			panic(err)
		}
	default:
		storage, err = st.NewFileStorage(config.Storage.File.Path)
		if err != nil {
			panic(err)
		}
	}

	//http клиент
	client := client.NewHttpClient()

	//луа движок
	le := l.NewLuaEngine(bot, cache, client, storage, ScriptsPath)

	//обрабатывающий сервер
	server := s.NewServer(le, bot, config, buffer)

	//получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tBot.GetUpdatesChan(u)

	logrus.Info("start processing")
	// обработка обновлений
	server.Start(updates)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	tBot.StopReceivingUpdates()

	server.Stop()

	logrus.Info("Server stopped")
}
