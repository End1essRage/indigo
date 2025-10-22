package server

import (
	"fmt"
	"strings"

	b "github.com/end1essrage/indigo-core/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// обработка запроса на авторизацию как админа
func (s *Server) handleAdm(upd *tgbotapi.Update) {
	//проверили что чат личный
	if !upd.Message.Chat.IsPrivate() {
		s.bot.SendMessage(upd.Message.Chat.ID, "Нельзя вызывать админ меню не в личном чате")
	}

	//проверили что пользователь админ
	userId := upd.Message.From.ID
	f := s.isAdmin(userId)
	if !f {
		s.bot.SendMessage(userId, "отказано в доступе")
	}

	//генерируем клавиатуру
	s.bot.SendKeyboard(userId, "меню:", GetAdminMenu())

	//TODO как обработать нажатие именно по этой клавиатуре(расширить конфиг кнопок? и изменить обрабокту колюэков с тг)
	//тут надо создать меню с кнопками, которые будут запускать разные штуки

	//Одна из кнопок будет запускать форму, форма может собирать данные как раньшше, отличается только шаги с клавиатурой и выходной скрипт

	//для кейса с каналами необходимо чтоб он создал кнопку с настрйокой каналов по нажатию кнопки запускалась форма
	//после выполнения формы надо будет запустить гошный код
}

func GetAdminMenu() b.MeshInlineKeyboard {
	var keyboard b.MeshInlineKeyboard

	// Первый ряд
	row1 := []b.MeshInlineButton{
		{
			Text:         "Каналы",
			Script:       "0",
			CustomCbData: "channels",
		},
	}
	keyboard.Rows = append(keyboard.Rows, row1)

	return keyboard
}

func (s *Server) isAdmin(userId int64) bool {
	return userId == s.config.Bot.AdminId
}

func (s *Server) admHandleCallbackQuery(chatId int64, data string) {
	if data == "channels" {
		s.bot.SendMessage(chatId, s.admFormatChannelsList())
	}
}

func (s *Server) admFormatChannelsList() string {
	items, err := s.service.GetChannels()
	if err != nil {
		return "ошибка получения каналов " + err.Error()
	}

	sb := strings.Builder{}
	sb.WriteString("Подключенные каналы: (название - код для скриптов)\n")
	for _, channel := range items {
		sb.WriteString(fmt.Sprintf("%s - %s", channel["title"], channel["code"]))
	}

	return sb.String()
}
