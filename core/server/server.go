package server

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/end1essrage/indigo-core/api"
	b "github.com/end1essrage/indigo-core/bot"
	c "github.com/end1essrage/indigo-core/config"
	l "github.com/end1essrage/indigo-core/lua"
	m "github.com/end1essrage/indigo-core/mapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

//работа с хранилищем

//http - эндпоинты

type Server struct {
	le       *l.LuaEngine
	bot      *b.TgBot
	config   *c.Config
	stopping bool
	handling bool
	stopped  chan struct{}
	mu       sync.Mutex
	api      *api.API
}

func NewServer(le *l.LuaEngine, bot *b.TgBot, config *c.Config) *Server {
	return &Server{le: le, bot: bot, config: config, stopped: make(chan struct{})}

}

func (s *Server) Start(updates tgbotapi.UpdatesChannel) {
	go func() {
		for update := range updates {
			s.HandleUpdate(&update)
		}
	}()

	if s.config.HTTP != nil {
		s.api = api.New(s.bot, s.le, s.config)
		go func() {
			if err := s.api.Start(); err != nil {
				panic(err)
			}
		}()
	}

}

func (h *Server) HandleUpdate(update *tgbotapi.Update) {
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
		lCtx := m.FromCallbackQueryToLuaContext(update.CallbackQuery)

		//удаляем сообщение с клавиатурой
		h.bot.DeleteMsg(lCtx.ChatId, update.CallbackQuery.Message.MessageID)

		logrus.Infof("cbdata script is %s", lCtx.CbData.Script)
		logrus.Infof("cbdata data is %s", lCtx.CbData.Data)

		if lCtx.CbData.Script != "" {
			scriptPath := fmt.Sprintf("scripts/%s", lCtx.CbData.Script)
			if err := h.le.ExecuteScript(scriptPath, lCtx); err != nil {
				logrus.Errorf("Error executing script: %v", err)
			}
		} else {
			h.bot.SendMessage(lCtx.ChatId, fmt.Sprintf("no script, custom data is %s", lCtx.CbData.Data))
		}

		return
	}

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

func (h *Server) Stop() {
	h.mu.Lock()
	h.stopping = true
	handling := h.handling
	h.mu.Unlock()

	if h.api != nil {
		h.api.Stop()
	}

	if handling {
		select {
		case <-h.stopped:
		case <-time.After(5 * time.Second): // Таймаут на случай блокировки
		}
	}

}

func (h *Server) handleCommand(upd *tgbotapi.Update, cmd *c.Command) {
	//запуск скрипта
	if cmd.Script != nil && *cmd.Script != "" {
		scriptPath := fmt.Sprintf("scripts/%s", *cmd.Script)
		if err := h.le.ExecuteScript(scriptPath, m.FromTgUpdateToLuaContext(upd)); err != nil {
			logrus.Errorf("Error executing script: %v", err)
		}
		logrus.Info("скрипт выполнен")
	}

	//обработка сообщения Reply
	if cmd.Reply != nil && *cmd.Reply != "" {
		h.bot.SendMessage(upd.Message.Chat.ID, *cmd.Reply)
	}

	// обработка клавиатуры
	if cmd.Keyboard != nil && *cmd.Keyboard != "" {
		// ищем по имени в map
		kb := h.config.Keyboards[*cmd.Keyboard]
		if kb == nil {
			logrus.Errorf("не удалось найти клавиутуру с именем : %s", *cmd.Keyboard)
			return
		}

		// обрабатываем клавиатуру из конфига
		if kb.Script != nil && *kb.Script != "" {
			scriptPath := fmt.Sprintf("scripts/%s", *kb.Script)
			if err := h.le.ExecuteScript(scriptPath, m.FromTgUpdateToLuaContext(upd)); err != nil {
				logrus.Errorf("Error executing script: %v", err)
			}
		} else {
			rMessage := *kb.Message
			kbMesh := b.ParseInlineKeyboard(kb)

			keyboard := b.CreateInlineKeyboard(kbMesh)

			replyMessage := tgbotapi.NewMessage(upd.Message.Chat.ID, rMessage)
			replyMessage.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}

			h.bot.Send(replyMessage)
		}
	}

}

func formatHelpMessage(cmds map[string]*c.Command) string {
	sb := strings.Builder{}
	for _, c := range cmds {
		sb.WriteString(fmt.Sprintf("%s - %s \n", c.Name, c.Description))
	}
	return sb.String()
}
