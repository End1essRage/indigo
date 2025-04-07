package lua

import (
	"fmt"

	b "github.com/end1essrage/indigo-core/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type LuaContext struct {
	MessageText string
	CbData      LuaCbData
	ChatId      int64
	FromId      int64
	FromName    string
}

type LuaCbData struct {
	Script string
	Data   string
}

type Module interface {
	Apply(le *LuaEngine, L *lua.LState)
}

// LuaStateBuilder для конфигурации стейта под конкретный скрипт
type LuaStateBuilder struct {
	modules []Module
	le      *LuaEngine
}

func NewStateBuilder(engine *LuaEngine) *LuaStateBuilder {
	return &LuaStateBuilder{
		le: engine,
	}
}

func (b *LuaStateBuilder) WithModule(m Module) *LuaStateBuilder {
	b.modules = append(b.modules, m)
	return b
}

func (b *LuaStateBuilder) Build() *lua.LState {
	L := lua.NewState()

	// Базовые модули
	base := CoreModule{}
	base.Apply(b.le, L)

	// Кастомные модули
	for _, module := range b.modules {
		module.Apply(b.le, L)
	}

	return L
}

// Lua engine wrapper
type LuaEngine struct {
	bot     Bot
	cache   Cache
	modules []Module
}

func (le *LuaEngine) RegisterFunctions(L *lua.LState) {
	for _, module := range le.modules {
		module.Apply(le, L)
	}
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
}

//работа с хранилищем

//http клиент

func NewLuaEngine(b Bot, c Cache) *LuaEngine {
	return &LuaEngine{bot: b, cache: c}
}

func (le *LuaEngine) ExecuteScript(scriptPath string, lContext LuaContext) error {
	logrus.Infof("ExecuteScript path:%s", scriptPath)
	L := NewStateBuilder(le).WithModule(&CacheModule{}).WithModule(&BotModule{}).Build()
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

// функция для конвертации Lua таблицы в MeshKeyboard
func fromLuaTableToMeshInlineKeyboard(lt *lua.LTable) b.MeshInlineKeyboard {
	var mesh b.MeshInlineKeyboard

	lt.ForEach(func(key lua.LValue, value lua.LValue) {
		if key.String() == "Rows" {
			if rows, ok := value.(*lua.LTable); ok {
				rows.ForEach(func(rowKey lua.LValue, rowValue lua.LValue) {
					if row, ok := rowValue.(*lua.LTable); ok {
						var meshRow []b.MeshInlineButton
						row.ForEach(func(btnKey lua.LValue, btnValue lua.LValue) {
							if btn, ok := btnValue.(*lua.LTable); ok {
								var meshBtn b.MeshInlineButton
								btn.ForEach(func(fieldKey lua.LValue, fieldValue lua.LValue) {
									switch fieldKey.String() {
									case "Text":
										meshBtn.Text = fieldValue.String()
									case "Script":
										meshBtn.Script = fieldValue.String()
									case "Data":
										meshBtn.CustomCbData = fieldValue.String()
									case "Name":
										meshBtn.Name = fieldValue.String()
									}
								})
								meshRow = append(meshRow, meshBtn)
							}
						})
						mesh.Rows = append(mesh.Rows, meshRow)
					}
				})
			}
		}
	})

	return mesh
}

// Реализации модулей
type CoreModule struct{}
type CacheModule struct{}
type BotModule struct{}

func (m *CoreModule) Apply(le *LuaEngine, L *lua.LState) {
	//Логирование
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logrus.Warnf("[LUA] %s", msg)
		return 0
	}))
}

func (m *CacheModule) Apply(le *LuaEngine, L *lua.LState) {
	// Кэш
	L.SetGlobal("cache_get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := le.cache.GetString(key)
		if value == "" {
			L.Push(lua.LNil)
		} else {
			L.Push(lua.LString(value))
		}
		return 0
	}))

	L.SetGlobal("cache_set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		le.cache.SetString(key, value)
		return 0
	}))
}

func (m *BotModule) Apply(le *LuaEngine, L *lua.LState) {
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
		mesh := fromLuaTableToMeshInlineKeyboard(meshTable)

		// Создаем и отправляем сообщение
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: b.CreateInlineKeyboard(mesh),
		}

		if err := le.bot.Send(msg); err != nil {
			logrus.Errorf("Error sending keyboard: %v", err)
			L.Push(lua.LString("send failed"))
			return 1
		}

		return 0
	}))
}
