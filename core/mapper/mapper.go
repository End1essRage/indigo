package mapper

import (
	"encoding/json"

	b "github.com/end1essrage/indigo-core/bot"
	l "github.com/end1essrage/indigo-core/lua"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func FromTgUpdateToLuaContext(update *tgbotapi.Update) l.LuaContext {
	c := l.LuaContext{}
	c.ChatId = update.Message.Chat.ID
	c.FromId = update.Message.From.ID
	c.FromName = update.Message.From.UserName
	c.MessageText = update.Message.Text
	return c
}

func FromCallbackQueryToLuaContext(cb *tgbotapi.CallbackQuery) l.LuaContext {
	c := l.LuaContext{}
	c.ChatId = cb.Message.Chat.ID
	c.FromId = cb.From.ID
	c.FromName = cb.From.UserName
	c.CbData = FromCallbackDataToLuaCbData(cb.Data)
	return c
}

func FromCallbackDataToLuaCbData(data string) l.LuaCbData {
	res := l.LuaCbData{}
	d := b.CbData{}
	if err := json.Unmarshal([]byte(data), &d); err != nil {
		logrus.Error("ошибка десериализции")
	}

	res.Script = *d.Script
	res.Data = *d.Data

	return res
}
