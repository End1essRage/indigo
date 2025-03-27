package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(b *tgbotapi.BotAPI) *TgBot {
	return &TgBot{bot: b}
}

func (t *TgBot) SendMessage(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	_, err := t.bot.Send(msg)

	return err
}

func (t *TgBot) Send(msg tgbotapi.MessageConfig) error {
	_, err := t.bot.Send(msg)

	return err
}
