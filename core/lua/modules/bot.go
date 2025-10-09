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

func (m *BotModule) applySendKeyboard(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		chatID := L.CheckInt64(1)
		text := L.CheckString(2)
		meshTable := L.CheckTable(3) // Принимаем таблицу вместо JSON строки

		// Конвертируем Lua таблицу в структуру Go
		mesh := b.FromLuaTableToMeshInlineKeyboard(meshTable)

		if err := m.bot.SendKeyboard(chatID, text, mesh); err != nil {
			logrus.Errorf("Error sending keyboard: %v", err)
			L.Push(lua.LString("send failed"))
			return 1
		}

		return 0
	}))
}

type BotModule struct{ bot Bot }

func NewBot(bot Bot) *BotModule {
	return &BotModule{bot: bot}
}
