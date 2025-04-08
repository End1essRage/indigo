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

type Cache interface {
	GetString(key string) string
	SetString(key string, val string)
	Exists(key string) bool
}

type Server struct {
	le         *l.LuaEngine
	bot        *b.TgBot
	config     *c.Config
	cache      Cache
	formWorker *FormWorker
	stopping   bool
	handling   bool
	stopped    chan struct{}
	mu         sync.Mutex
	api        *api.API
}

func NewServer(le *l.LuaEngine, bot *b.TgBot, config *c.Config, cache Cache) *Server {
	return &Server{
		le:         le,
		bot:        bot,
		config:     config,
		cache:      cache,
		formWorker: NewFormWorker(bot, cache, config, le),
		stopped:    make(chan struct{}),
	}
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

	// Handle forms first
	if s.formWorker.HasActiveForm(update) {
		s.formWorker.HandleInput(update)
		return
	}

	// Handle callback queries
	if update.CallbackQuery != nil {
		s.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Handle commands
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
	lCtx := m.FromCallbackQueryToLuaContext(query)
	s.bot.DeleteMsg(lCtx.ChatId, query.Message.MessageID)

	if lCtx.CbData.Script != "" {
		if err := s.le.ExecuteScript(lCtx.CbData.Script, lCtx); err != nil {
			logrus.Errorf("Callback script error: %v", err)
		}
	}
}

func (s *Server) handleCommand(upd *tgbotapi.Update) {
	// Handle help command
	if upd.Message.Command() == "help" {
		s.bot.SendMessage(upd.Message.Chat.ID, s.formatHelpMessage())
		return
	}

	cmd := s.config.Commands[upd.Message.Command()]
	if cmd == nil {
		s.bot.SendMessage(upd.Message.Chat.ID, "Unknown command: "+upd.Message.Command())
		return
	}

	// Execute command script
	if cmd.Script != nil && *cmd.Script != "" {
		ctx := m.FromTgUpdateToLuaContext(upd)
		if err := s.le.ExecuteScript(*cmd.Script, ctx); err != nil {
			logrus.Errorf("Command script error: %v", err)
		}
	}

	// Handle form start
	if cmd.Form != nil && *cmd.Form != "" {
		if err := s.formWorker.StartForm(*cmd.Form, upd.Message.From.ID, upd); err != nil {
			logrus.Errorf("Form start error: %v", err)
			s.bot.SendMessage(upd.Message.Chat.ID, "Failed to start form: "+err.Error())
		}
	}

	// Send reply if specified
	if cmd.Reply != nil && *cmd.Reply != "" {
		s.bot.SendMessage(upd.Message.Chat.ID, *cmd.Reply)
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
