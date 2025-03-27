package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type LuaContext struct {
	MessageText string
	CbData      LuaCbData
	ChatId      int64
	FromId      int64
	FromName    string
}

type LuaCbData struct {
	Handler string
}

func FromTgUpdateToLuaContext(update *tgbotapi.Update) LuaContext {
	c := LuaContext{}
	c.ChatId = update.Message.Chat.ID
	c.FromId = update.Message.From.ID
	c.FromName = update.Message.From.UserName
	c.MessageText = update.Message.Text
	return c
}

func FromCallbackQueryToLuaContext(cb *tgbotapi.CallbackQuery) LuaContext {
	c := LuaContext{}
	c.ChatId = cb.Message.Chat.ID
	c.FromId = cb.From.ID
	c.FromName = cb.From.UserName
	c.CbData = FromCallbackDataToLuaCbData(cb.Data)
	return c
}

func FromCallbackDataToLuaCbData(data string) LuaCbData {
	res := LuaCbData{}
	res.Handler = data
	logrus.Warn(res.Handler)
	return res
}
