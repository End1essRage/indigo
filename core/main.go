package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

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

	//получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tBot.GetUpdatesChan(u)

	logrus.Info("start processing")
	// обработка обновлений
	for update := range updates {
		//обработка кнопок
		if update.CallbackQuery != nil {
			lCtx := FromCallbackQueryToLuaContext(update.CallbackQuery)
			logrus.Infof("cbdata handler is %s", lCtx.CbData.Handler)
			scriptPath := fmt.Sprintf("scripts/%s", lCtx.CbData.Handler)
			if err := le.ExecuteScript(scriptPath, lCtx); err != nil {
				logrus.Errorf("Error executing script: %v", err)
			}
			continue
		}
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
				handleCommand(config, le, &update, &cmd)
				break
			}
		}
	}
}

func handleCommand(config *Config, le *LuaEngine, upd *tgbotapi.Update, cmd *Command) {
	//запуск скрипта
	if cmd.Handler != nil && *cmd.Handler != "" {
		scriptPath := fmt.Sprintf("scripts/%s", *cmd.Handler)
		if err := le.ExecuteScript(scriptPath, FromTgUpdateToLuaContext(upd)); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
	}

	//обработка блока Reply
	if cmd.Reply != nil {
		if cmd.Reply.Msg != nil && *cmd.Reply.Msg != "" {
			le.bot.SendMessage(upd.Message.Chat.ID, *cmd.Reply.Msg)
		}

		if cmd.Reply.Keyboard != nil && *cmd.Reply.Keyboard != "" {
			rMessage := ""
			keyboard := make([][]tgbotapi.InlineKeyboardButton, 0)
			//поиск keyboard в конфиге
			for _, kb := range config.Keyboards {
				if *cmd.Reply.Keyboard == kb.Name {
					logrus.Infof("keyboard is %+v", kb)
					// текст сообщения с клавиатурой
					rMessage = kb.Message
					//проходимся по блоку Buttons по каждому Row
					for _, r := range kb.Buttons {
						row := make([]tgbotapi.InlineKeyboardButton, 0)
						//проходимся по кнопкам внутри Row
						for _, b := range r.Row {
							//заполняем CallBackData
							data := ""
							if b.Handler != nil {
								data = *b.Handler
							}
							btn := tgbotapi.NewInlineKeyboardButtonData(b.Text, data)
							row = append(row, btn)
						}
						keyboard = append(keyboard, row)
					}
					break
				}
			}

			replyMessage := tgbotapi.NewMessage(upd.Message.Chat.ID, rMessage)
			replyMessage.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}

			le.bot.Send(replyMessage)
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
