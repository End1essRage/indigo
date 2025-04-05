package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	le       *LuaEngine
	bot      *TgBot
	config   *Config
	stopping bool
	handling bool
	stopped  chan struct{}
	mu       sync.Mutex
}

func NewHandler(le *LuaEngine, bot *TgBot, config *Config) *Handler {
	return &Handler{le: le, bot: bot, config: config, stopped: make(chan struct{})}
}

func (h *Handler) HandleUpdate(update *tgbotapi.Update) {
	h.mu.Lock()
	if h.stopping {
		h.mu.Unlock()
		h.stopped <- struct{}{}
		return
	}
	h.handling = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		h.handling = false
		h.mu.Unlock()
	}()

	//обработка кнопок
	if update.CallbackQuery != nil {
		lCtx := FromCallbackQueryToLuaContext(update.CallbackQuery)

		logrus.Infof("cbdata script is %s", lCtx.CbData.Script)
		logrus.Infof("cbdata data is %s", lCtx.CbData.Data)

		if lCtx.CbData.Script != "" {
			scriptPath := fmt.Sprintf("scripts/%s", lCtx.CbData.Script)
			if err := h.le.ExecuteScript(scriptPath, lCtx); err != nil {
				logrus.Errorf("Error executing script: %v", err)
			}
		} else {
			h.bot.SendMessage(lCtx.ChatId, fmt.Sprintf("no script custom data is %s", lCtx.CbData.Data))
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
		return
	}

	cmd := h.config.Commands[update.Message.Command()]
	if cmd == nil {
		logrus.Error("no such command")
		h.bot.SendMessage(update.Message.Chat.ID, "не распознана команда "+update.Message.Command())
		return
	}

	//обрабатываем команду
	h.handleCommand(update, cmd)
}

func (h *Handler) Stop() {
	h.mu.Lock()
	h.stopping = true
	handling := h.handling
	h.mu.Unlock()

	if handling {
		select {
		case <-h.stopped:
		case <-time.After(5 * time.Second): // Таймаут на случай блокировки
		}
	}
}

func (h *Handler) handleCommand(upd *tgbotapi.Update, cmd *Command) {
	//запуск скрипта
	if cmd.Script != nil && *cmd.Script != "" {
		scriptPath := fmt.Sprintf("scripts/%s", *cmd.Script)
		if err := h.le.ExecuteScript(scriptPath, FromTgUpdateToLuaContext(upd)); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
	}

	//обработка блока Reply
	if cmd.Reply != nil {
		if cmd.Reply.Msg != nil && *cmd.Reply.Msg != "" {
			h.le.bot.SendMessage(upd.Message.Chat.ID, *cmd.Reply.Msg)
		}

		// обработка клавиатуры
		if cmd.Reply.Keyboard != nil && *cmd.Reply.Keyboard != "" {
			// ищем по имени в map
			kb := h.config.Keyboards[*cmd.Reply.Keyboard]
			if kb == nil {
				logrus.Errorf("не удалось найти клавиутуру с именем : %s", *cmd.Reply.Keyboard)
				return
			}

			switch kb.Type {
			case "inline":
				// обрабатываем клавиатуру из конфига
				kbMesh, rMessage := h.parseInlineKeyboard(kb, upd)

				keyboard := createInlineKeyboard(kbMesh)

				replyMessage := tgbotapi.NewMessage(upd.Message.Chat.ID, rMessage)
				replyMessage.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}

				h.le.bot.Send(replyMessage)

			case "reply":
				// обрабатываем клавиатуру из конфига
				kbMesh, rMessage := h.parseReplyKeyboard(kb, upd)

				keyboard := createReplyKeyboard(kbMesh)

				replyMessage := tgbotapi.NewMessage(upd.Message.Chat.ID, rMessage)
				replyMessage.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{Keyboard: keyboard}

				h.le.bot.Send(replyMessage)
			}

		}
	}
}

func (h *Handler) parseInlineKeyboard(kb *Keyboard, upd *tgbotapi.Update) (MeshInlineKeyboard, string) {
	kbMesh := MeshInlineKeyboard{}
	rMessage := ""
	//если скрипт
	if kb.Script != nil && *kb.Script != "" {
		scriptPath := fmt.Sprintf("scripts/%s", *kb.Script)
		if err := h.le.ExecuteScript(scriptPath, FromTgUpdateToLuaContext(upd)); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
	} else {
		rMessage = *kb.Message

		//проходимся по блоку Buttons по каждому Row
		for _, r := range *kb.Buttons {
			row := make([]MeshInlineButton, 0)
			//проходимся по кнопкам внутри Row
			for _, b := range r.Row {
				btn := MeshInlineButton{Name: b.Name, Text: b.Text}
				//заполняем CallBackData
				if b.Script != nil {
					btn.Script = *b.Script
				}
				if b.Data != nil {
					logrus.Warnf("custom data %s ", *b.Data)
					btn.CustomCbData = *b.Data
				}

				row = append(row, btn)
			}
			kbMesh.Rows = append(kbMesh.Rows, row)
		}
	}

	return kbMesh, rMessage
}

func (h *Handler) parseReplyKeyboard(kb *Keyboard, upd *tgbotapi.Update) (MeshReplyKeyboard, string) {
	kbMesh := MeshReplyKeyboard{}
	rMessage := ""
	//если скрипт
	if kb.Script != nil && *kb.Script != "" {
		scriptPath := fmt.Sprintf("scripts/%s", *kb.Script)
		if err := h.le.ExecuteScript(scriptPath, FromTgUpdateToLuaContext(upd)); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
	} else {
		rMessage = *kb.Message

		//проходимся по блоку Buttons по каждому Row
		for _, r := range *kb.Buttons {
			row := make([]MeshReplyButton, 0)
			//проходимся по кнопкам внутри Row
			for _, b := range r.Row {
				btn := MeshReplyButton{Text: b.Text}

				row = append(row, btn)
			}
			kbMesh.Rows = append(kbMesh.Rows, row)
		}
	}

	return kbMesh, rMessage
}

func formatHelpMessage(cmds map[string]*Command) string {
	time.Sleep(2 * time.Second)
	sb := strings.Builder{}
	for _, c := range cmds {
		sb.WriteString(fmt.Sprintf("%s - %s \n", c.Name, c.Description))
	}
	return sb.String()
}
