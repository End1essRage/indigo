package service

import (
	b "github.com/end1essrage/indigo-core/bot"
)

type Service struct {
	ChatMemberService
	bot     Bot
	storage Storage
}

type ChatMemberService interface {
	HandleBotAdm()
}

type Bot interface {
	SendMessage(chatId int64, text string) error
	SendChannelMessage( /*channel string, */ text string, mesh *b.MeshInlineKeyboard) error
	SendKeyboard(chatId int64, text string, mesh b.MeshInlineKeyboard) error
}

type Storage interface {
	Save(entityType string, data interface{}) (string, error)
	Load(entityType string, id string, result interface{}) error
	LoadArray(docFolder, docPath string) ([]interface{}, error)
}

func NewService(bot Bot, storage Storage) *Service {
	return &Service{bot: bot, storage: storage}
}
