package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type LuaContext struct {
	MessageText string
	ChatId      int64
	FromId      int64
	FromName    string
}

// Lua engine wrapper
type LuaEngine struct {
	bot Bot
}

// интерфейс описывает какие ручки торчат в lua
type Bot interface {
	SendMessage(chatId int64, msg string) error
}

func NewLuaEngine(b Bot) *LuaEngine {
	return &LuaEngine{bot: b}
}

func (le *LuaEngine) RegisterFunctions(L *lua.LState) {

	//Логирование
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logrus.Warnf("[LUA] %s", msg)
		return 0
	}))

	//Отправка сообщения
	L.SetGlobal("send_message", L.NewFunction(func(L *lua.LState) int {
		logrus.Infof("send_message")

		chatID := L.ToInt64(1)
		text := L.ToString(2)

		logrus.Infof("send_message ChatId: %v  text: %s", chatID, text)

		if err := le.bot.SendMessage(chatID, text); err != nil {
			logrus.Errorf("Error sending message: %v", err)
		}
		return 0
	}))
}

func (le *LuaEngine) ExecuteScript(scriptPath string, lContext LuaContext) error {
	logrus.Infof("ExecuteScript path:%s", scriptPath)
	L := lua.NewState()
	defer L.Close()

	le.RegisterFunctions(L)

	// Прокидываем контекст
	ctx := L.NewTable()

	L.SetField(ctx, "chat_id", lua.LNumber(lContext.ChatId))
	L.SetField(ctx, "text", lua.LString(lContext.MessageText))
	user := L.NewTable()
	L.SetField(user, "id", lua.LNumber(lContext.FromId))
	L.SetField(user, "from_name", lua.LString(lContext.FromName))
	L.SetField(ctx, "user", user)

	L.SetGlobal("ctx", ctx)

	// Run script
	if err := L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("lua error: %v", err)
	}

	return nil
}
