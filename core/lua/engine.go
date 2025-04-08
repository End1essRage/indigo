package lua

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type LuaContext struct {
	RequestData map[string]interface{}
	FormData    map[string]interface{}
	Headers     http.Header
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
	http    HttpClient
	storage Storage
}

func NewLuaEngine(b Bot, c Cache, h HttpClient, s Storage) *LuaEngine {
	return &LuaEngine{bot: b, cache: c, http: h, storage: s}
}

func (le *LuaEngine) ExecuteScript(scriptPath string, lContext LuaContext) error {
	logrus.Infof("ExecuteScript path:%s", scriptPath)
	L := NewStateBuilder(le).
		WithModule(&CacheModule{}).
		WithModule(&BotModule{}).
		WithModule(&HttpModule{}).
		WithModule(&StorageModule{}).
		Build()
	defer L.Close()

	// Создаем таблицу для контекста
	ctx := L.NewTable()

	// Базовые поля
	L.SetField(ctx, "chat_id", lua.LNumber(lContext.ChatId))
	L.SetField(ctx, "text", lua.LString(lContext.MessageText))

	// Обработка callback данных
	cbData := L.NewTable()
	L.SetField(cbData, "script", lua.LString(lContext.CbData.Script))
	L.SetField(cbData, "data", lua.LString(lContext.CbData.Data))
	L.SetField(ctx, "cb_data", cbData)

	// Прокидываем form_data как Lua таблицу
	if lContext.FormData != nil {
		formDataTable := convertMapToLuaTable(L, lContext.FormData)
		L.SetField(ctx, "form_data", formDataTable)
	}

	// Прокидываем request_data
	if lContext.RequestData != nil {
		reqDataTable := convertMapToLuaTable(L, lContext.RequestData)
		L.SetField(ctx, "req_data", reqDataTable)
	}

	// Информация о пользователе
	user := L.NewTable()
	L.SetField(user, "id", lua.LNumber(lContext.FromId))
	L.SetField(user, "name", lua.LString(lContext.FromName))
	L.SetField(ctx, "user", user)

	// Устанавливаем глобальную переменную ctx
	L.SetGlobal("ctx", ctx)

	// Выполняем скрипт
	if err := L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("lua error: %v", err)
	}

	return nil
}

// Функция для конвертации map[string]interface{} в Lua таблицу
func convertMapToLuaTable(L *lua.LState, data map[string]interface{}) *lua.LTable {
	tbl := L.NewTable()
	for k, v := range data {
		switch value := v.(type) {
		case string:
			L.SetField(tbl, k, lua.LString(value))
		case int, int64, float64:
			L.SetField(tbl, k, lua.LNumber(value.(float64)))
		case bool:
			L.SetField(tbl, k, lua.LBool(value))
		case map[string]interface{}:
			L.SetField(tbl, k, convertToLuaTable(L, value))
		case []interface{}:
			arr := L.NewTable()
			for i, item := range value {
				switch elem := item.(type) {
				case string:
					L.RawSetInt(arr, i+1, lua.LString(elem))
				case int, int64, float64:
					L.RawSetInt(arr, i+1, lua.LNumber(elem.(float64)))
				case bool:
					L.RawSetInt(arr, i+1, lua.LBool(elem))
				case map[string]interface{}:
					L.RawSetInt(arr, i+1, convertToLuaTable(L, elem))
				}
			}
			L.SetField(tbl, k, arr)
		}
	}
	return tbl
}
