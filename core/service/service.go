package service

import (
	b "github.com/end1essrage/indigo-core/bot"
	c "github.com/end1essrage/indigo-core/cache"
	s "github.com/end1essrage/indigo-core/storage"
)

type Service struct {
	ChatMemberService
	bot     Bot
	storage s.Storage
	cache   c.Cache
}

type ChatMemberService interface {
	HandleBotAdm()
	GetChannels() ([]s.Entity, error)
	GetChannelId(code string) (int64, error)
}

type Bot interface {
	SendMessage(chatId int64, text string) error
	SendKeyboard(chatId int64, text string, mesh b.MeshInlineKeyboard) error
}

func NewService(bot Bot, storage s.Storage, cache c.Cache) *Service {
	return &Service{bot: bot, storage: storage, cache: cache}
}
