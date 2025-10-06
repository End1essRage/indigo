package service

import (
	"fmt"
)

type BotAdminRequest struct {
	FromId    int64
	ChannelId int64
	Title     string
}

type ChannelInfo struct {
	ChanId int64
	Title  string
	Code   string
}

func (s *Service) HandleBotAdd(req BotAdminRequest) {
	code := req.Title[0:3]
	//сохранить канал с каким-то кодо
	s.storage.Save("channelAdm", ChannelInfo{
		ChanId: req.ChannelId,
		Title:  req.Title,
		Code:   code})

	//отправить сообщение админу в личку тому кто добавил
	s.bot.SendMessage(req.FromId, fmt.Sprintf("Бот добавлен в канал %s с кодом %s для смены кода перейдите в админ меню", req.Title, code))

	//считать заданный код и присвоить сохраненному каналу

	//отправлять в канал можно в луа по коду
}

func (s *Service) HandleBotRemove(req BotAdminRequest) {
	//удалить из базы
	s.bot.SendMessage(req.FromId, fmt.Sprintf("Бот был удален из канала %s", req.Title))
}
