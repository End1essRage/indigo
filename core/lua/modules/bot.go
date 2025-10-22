package lua_modules

import (
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	b "github.com/end1essrage/indigo-core/bot"
)

type Bot interface {
	SendMessage(chatId int64, text string) error
	SendKeyboard(chatId int64, text string, mesh b.MeshInlineKeyboard) error
}

//можно ли схлопнуть в один метод?

func (m *BotModule) applySendMessage(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		chatID := L.ToInt64(1)
		text := L.ToString(2)

		logrus.Infof("send_message ChatId: %v  text: %s", chatID, text)

		if err := m.bot.SendMessage(chatID, text); err != nil {
			logrus.Errorf("Error sending message: %v", err)
		}
		return 0
	}))
}

func (m *BotModule) applySend(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		chatID := L.CheckInt64(1)
		text := L.CheckString(2)

		if L.GetTop() >= 3 {
			meshTable := L.CheckTable(3)
			mesh := b.FromLuaTableToMeshInlineKeyboard(meshTable)

			if err := m.bot.SendKeyboard(chatID, text, mesh); err != nil {
				logrus.Errorf("Error sending keyboard: %v", err)
				L.Push(lua.LString("send failed"))
				return 1
			}
		} else {
			if err := m.bot.SendMessage(chatID, text); err != nil {
				logrus.Errorf("Error sending message: %v", err)
				L.Push(lua.LString("send failed"))
				return 1
			}
		}

		L.Push(lua.LNil)
		return 1
	}))
}

func (m *BotModule) applySendChannel(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		chanCode := L.CheckString(1)
		text := L.CheckString(2)

		chatID, err := m.service.GetChannelId(chanCode)
		if err != nil {
			logrus.Errorf("ошибка получения айди канала по коду %s", err.Error())
			L.Push(lua.LString("send failed"))
			return 1
		}

		if L.GetTop() >= 3 {
			meshTable := L.CheckTable(3)
			mesh := b.FromLuaTableToMeshInlineKeyboard(meshTable)

			if err := m.bot.SendKeyboard(chatID, text, mesh); err != nil {
				logrus.Errorf("Error sending keyboard: %v", err)
				L.Push(lua.LString("send failed"))
				return 1
			}
		} else {
			if err := m.bot.SendMessage(chatID, text); err != nil {
				logrus.Errorf("Error sending message: %v", err)
				L.Push(lua.LString("send failed"))
				return 1
			}
		}

		L.Push(lua.LNil)
		return 1
	}))
}

type BotModule struct {
	bot     Bot
	service Service
}

func NewBot(bot Bot, service Service) *BotModule {
	return &BotModule{bot: bot, service: service}
}
