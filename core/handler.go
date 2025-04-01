package main

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	le     *LuaEngine
	bot    *TgBot
	config *Config
}

func NewHandler(le *LuaEngine, bot *TgBot, config *Config) *Handler {
	return &Handler{le: le, bot: bot, config: config}
}

func (h *Handler) HandleUpdate(update *tgbotapi.Update) {
	//обработка кнопок
	if update.CallbackQuery != nil {
		lCtx := FromCallbackQueryToLuaContext(update.CallbackQuery)
		logrus.Infof("cbdata handler is %s", lCtx.CbData.Handler)
		scriptPath := fmt.Sprintf("scripts/%s", lCtx.CbData.Handler)
		if err := h.le.ExecuteScript(scriptPath, lCtx); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
		return
	}
	//отбрасываем с пустым сообщением(надо будет убрать для обработки кнопок)
	if update.Message == nil {
		return

	}

	//отбрасываем все кроме команд
	if !update.Message.IsCommand() {
		return

	}

	//TODO возможность перезаписать через yml
	//обрабатываем help
	if update.Message.Command() == "help" {
		h.bot.SendMessage(update.Message.Chat.ID, formatHelpMessage(h.config.Commands))
	}

	// поиск команды
	for _, cmd := range h.config.Commands {
		if update.Message.Command() == cmd.Name {
			h.handleCommand(update, &cmd)
			break
		}
	}
}

func (h *Handler) handleCommand(upd *tgbotapi.Update, cmd *Command) {
	//запуск скрипта
	if cmd.Handler != nil && *cmd.Handler != "" {
		scriptPath := fmt.Sprintf("scripts/%s", *cmd.Handler)
		if err := h.le.ExecuteScript(scriptPath, FromTgUpdateToLuaContext(upd)); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
	}

	//обработка блока Reply
	if cmd.Reply != nil {
		if cmd.Reply.Msg != nil && *cmd.Reply.Msg != "" {
			h.le.bot.SendMessage(upd.Message.Chat.ID, *cmd.Reply.Msg)
		}

		if cmd.Reply.Keyboard != nil && *cmd.Reply.Keyboard != "" {
			rMessage := ""
			keyboard := make([][]tgbotapi.InlineKeyboardButton, 0)
			//поиск keyboard в конфиге
			for _, kb := range h.config.Keyboards {
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

			h.le.bot.Send(replyMessage)
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
