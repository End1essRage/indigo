package service

import (
	b "github.com/end1essrage/indigo-core/bot"
	s "github.com/end1essrage/indigo-core/storage"
)

type Service struct {
	ChatMemberService
	bot     Bot
	storage s.Storage
}

type ChatMemberService interface {
	HandleBotAdm()
}

type Bot interface {
	SendMessage(chatId int64, text string) error
	SendKeyboard(chatId int64, text string, mesh b.MeshInlineKeyboard) error
}

func NewService(bot Bot, storage s.Storage) *Service {
	return &Service{bot: bot, storage: storage}
}
