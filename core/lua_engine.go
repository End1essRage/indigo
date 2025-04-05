package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

// Lua engine wrapper
type LuaEngine struct {
	bot   Bot
	cache Cache
}

// интерфейс описывает какие ручки торчат в lua
type Bot interface {
	SendMessage(chatId int64, text string) error
	Send(msg tgbotapi.MessageConfig) error
}

// работа с кэшом
type Cache interface {
	GetString(key string) string
	SetString(key string, val string)
	Exist(key string) bool
}

func NewLuaEngine(b Bot, c Cache) *LuaEngine {
	return &LuaEngine{bot: b, cache: c}
}

// можно через паттерн билдер сделать возможность подключения луа/го модулей к своему боту, чтоб не тянуть кучу функционала
// но это лучше скорее всего делать через cli или как-то чтоб и не раздувался бинарник, мб кодогенеративано подключать модули рядом
func (le *LuaEngine) RegisterFunctions(L *lua.LState) {

	//Логирование
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logrus.Warnf("[LUA] %s", msg)
		return 0
	}))

	// Кэш
	L.SetGlobal("cache_get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := le.cache.GetString(key)
		if value == "" {
			L.Push(lua.LNil)
		} else {
			L.Push(lua.LString(value))
		}
		return 1
	}))

	L.SetGlobal("cache_set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		le.cache.SetString(key, value)
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

	//отправка клавиатуры
	L.SetGlobal("send_keyboard", L.NewFunction(func(L *lua.LState) int {
		chatID := L.CheckInt64(1)
		text := L.CheckString(2)
		meshTable := L.CheckTable(3) // Принимаем таблицу вместо JSON строки

		// Конвертируем Lua таблицу в структуру Go
		mesh := FromLuaTableToMeshInlineKeyboard(L, meshTable)

		// Создаем и отправляем сообщение
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: createInlineKeyboard(mesh),
		}

		if err := le.bot.Send(msg); err != nil {
			logrus.Errorf("Error sending keyboard: %v", err)
			L.Push(lua.LString("send failed"))
			return 1
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
	L.SetField(ctx, "cb_data", lua.LString(lContext.CbData.Data))
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
