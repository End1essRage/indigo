package service

import (
	"context"
	"fmt"

	"github.com/end1essrage/indigo-core/storage"
	"github.com/sirupsen/logrus"
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
	entity := storage.NewEntity()
	entity["ChanId"] = req.ChannelId
	entity["Title"] = req.Title
	entity["Code"] = code

	id, err := s.storage.Create(context.TODO(), "channelAdm", entity)
	if err != nil {
		logrus.Errorf("ошибка создания записи")
	}

	logrus.Infof("id документа %s", id)

	//отправить сообщение админу в личку тому кто добавил
	s.bot.SendMessage(req.FromId, fmt.Sprintf("Бот добавлен в канал %s с кодом %s для смены кода перейдите в админ меню", req.Title, code))

	//считать заданный код и присвоить сохраненному каналу

	//отправлять в канал можно в луа по коду
}

func (s *Service) HandleBotRemove(req BotAdminRequest) {
	//удалить из базы
	s.bot.SendMessage(req.FromId, fmt.Sprintf("Бот был удален из канала %s", req.Title))
}
