package service

import (
	"context"

	b "github.com/end1essrage/indigo-core/bot"
	"github.com/end1essrage/indigo-core/storage"
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
	Get(ctx context.Context, collection string, count int, query string) ([]storage.Entity, error)
	GetById(ctx context.Context, collection string, id string) (storage.Entity, error)

	Create(ctx context.Context, collection string, entity storage.Entity) (string, error)

	UpdateById(ctx context.Context, collection string, id string, entity storage.Entity) error
	Update(ctx context.Context, collection string, query string, entity storage.Entity) (int, error)

	DeleteById(ctx context.Context, collection string, id string) error
	Delete(ctx context.Context, collection string, query string) (int, error)
}

func NewService(bot Bot, storage Storage) *Service {
	return &Service{bot: bot, storage: storage}
}
