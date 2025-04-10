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

//при использовании в личке chat.Id == From.Id

type Buffer interface {
	GetString(key string) string
	SetString(key string, val string) error
	Exists(key string) bool
}

type Server struct {
	le         *l.LuaEngine
	bot        *b.TgBot
	config     *c.Config
	api        *api.API
	formWorker *FormWorker
	stopping   bool
	handling   bool
	stopped    chan struct{}
	mu         sync.Mutex
}

func NewServer(le *l.LuaEngine, bot *b.TgBot, config *c.Config, buffer Buffer) *Server {
	s := &Server{
		le:         le,
		bot:        bot,
		config:     config,
		formWorker: NewFormWorker(bot, buffer, config, le),
		stopped:    make(chan struct{}),
	}
	if s.config.HTTP != nil {
		s.api = api.New(s.le, s.config.HTTP)
	}
	return s
}

func (s *Server) Start(updates tgbotapi.UpdatesChannel) {
	go func() {
		for update := range updates {
			s.HandleUpdate(&update)
		}
	}()

	if s.config.HTTP != nil {
		go func() {
			if err := s.api.Start(); err != nil {
				logrus.Fatalf("Failed to start API: %v", err)
			}
		}()
	}
}

func (s *Server) HandleUpdate(update *tgbotapi.Update) {
	s.mu.Lock()
	if s.stopping {
		s.mu.Unlock()
		s.stopped <- struct{}{}
		return
	}
	s.handling = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.handling = false
		s.mu.Unlock()
	}()

	// Формы
	if s.formWorker.HasActiveForm(update) {
		s.formWorker.HandleInput(update)
		return
	}

	// Кнопки
	if update.CallbackQuery != nil {
		s.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Команды
	if update.Message != nil && update.Message.IsCommand() {
		s.handleCommand(update)
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	s.stopping = true
	handling := s.handling
	s.mu.Unlock()

	if s.api != nil {
		s.api.Stop()
	}

	if handling {
		select {
		case <-s.stopped:
		case <-time.After(5 * time.Second):
		}
	}
}

func (s *Server) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// формируем контекст
	lCtx := m.FromCallbackQueryToLuaContext(query)

	//удаляем сообщение с клавиатурой
	s.bot.DeleteMsg(lCtx.ChatId, query.Message.MessageID)

	//если есть скрипт запускаем
	if lCtx.CbData.Script != "" {
		if err := s.le.ExecuteScript(lCtx.CbData.Script, lCtx); err != nil {
			logrus.Errorf("Callback script error: %v", err)
		}
	}
}

func (s *Server) handleCommand(upd *tgbotapi.Update) {
	chatId := upd.Message.Chat.ID

	if upd.Message.Command() == "help" {
		s.bot.SendMessage(chatId, s.formatHelpMessage())
		return
	}

	//ищем команду
	cmd := s.config.Commands[upd.Message.Command()]
	if cmd == nil {
		s.bot.SendMessage(chatId, "Unknown command: "+upd.Message.Command())
		return
	}

	// Выполняем скрипт
	if cmd.Script != nil && *cmd.Script != "" {
		ctx := m.FromTgUpdateToLuaContext(upd)
		if err := s.le.ExecuteScript(*cmd.Script, ctx); err != nil {
			logrus.Errorf("Command script error: %v", err)
		}
	}

	// Генерируем клавиатуру
	if cmd.Keyboard != nil && *cmd.Keyboard != "" {
		kb := s.config.Keyboards[*cmd.Keyboard]
		if kb == nil {
			logrus.Errorf("keyboard '%s' not found", *cmd.Keyboard)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			b.CreateInlineKeyboard(b.ParseInlineKeyboard(kb))...,
		)

		msg := tgbotapi.NewMessage(chatId, *kb.Message)
		msg.ReplyMarkup = &keyboard

		s.bot.Send(msg)
	}

	// Запускаем форму
	if cmd.Form != nil && *cmd.Form != "" {
		if err := s.formWorker.StartForm(*cmd.Form, upd.Message.From.ID, upd); err != nil {
			logrus.Errorf("Form start error: %v", err)
			s.bot.SendMessage(chatId, "Failed to start form: "+err.Error())
		}
	}

	// Шлем ответ
	if cmd.Reply != nil && *cmd.Reply != "" {
		s.bot.SendMessage(chatId, *cmd.Reply)
	}
}

func (s *Server) formatHelpMessage() string {
	sb := strings.Builder{}
	sb.WriteString("Available commands:\n")
	for _, cmd := range s.config.Commands {
		sb.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name, cmd.Description))
	}
	return sb.String()
}
