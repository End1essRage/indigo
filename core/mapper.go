package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func FromTgUpdateToLuaContext(update tgbotapi.Update) LuaContext {
	c := LuaContext{}
	c.ChatId = update.Message.Chat.ID
	c.FromId = update.Message.From.ID
	c.FromName = update.Message.From.UserName
	c.MessageText = update.Message.Text
	return c
}
