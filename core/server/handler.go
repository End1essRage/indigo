package server

import (
	"fmt"
	"strings"
	"sync"

	b "github.com/end1essrage/indigo-core/bot"
	c "github.com/end1essrage/indigo-core/config"
	m "github.com/end1essrage/indigo-core/mapper"
	"github.com/end1essrage/indigo-core/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

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

	//перехватчики
	wg := &sync.WaitGroup{}

	if len(s.interceptors) > 0 {
		for k, v := range s.interceptors {
			switch k {
			case c.AffectMode_All:
				for _, f := range v {
					wg.Add(1)
					go func() {
						defer wg.Done()
						f.Use(update)
					}()
				}
			case c.AffectMode_Text:
				if update.Message.IsCommand() || update.CallbackQuery != nil {
					continue
				}
				for _, f := range v {
					wg.Add(1)
					go func() {
						defer wg.Done()
						f.Use(update)
					}()
				}
			}
		}
	}

	wg.Wait()

	// Кнопки
	if update.CallbackQuery != nil {
		s.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// обработка добавления/удаления из чата
	if update.MyChatMember != nil {
		s.handleChatMember(update.MyChatMember)
	}

	// Команды
	if update.Message != nil && update.Message.IsCommand() {
		s.handleCommand(update)
	}
}

// обработка добавления/удаления из чата
func (s *Server) handleChatMember(upd *tgbotapi.ChatMemberUpdated) {
	if upd.NewChatMember.Status != "administrator" {
		return
	}

	go s.service.HandleBotAdd(service.BotAdminRequest{FromId: upd.From.ID, ChannelId: upd.Chat.ID, Title: upd.Chat.Title})

}

// обработка запроса на авторизацию как админа
func handleAdm() {

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

	// атвообработка команды хелп,мб стоит дать возможность оверрайдить
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
