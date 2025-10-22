package lua

import (
	"context"
	"fmt"
	"time"

	"github.com/end1essrage/indigo-core/helpers"
	h "github.com/end1essrage/indigo-core/lua/helpers"
	m "github.com/end1essrage/indigo-core/lua/modules"
	"github.com/end1essrage/indigo-core/secret"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

// Lua engine wrapper
type LuaEngine struct {
	bot      m.Bot
	cache    m.Cache
	service  m.Service
	http     m.HttpClient
	storage  m.Storage
	BasePath string
	Secret   *secret.SecretsOperator
	scripts  map[string][]byte
}

func NewLuaEngine(b m.Bot, c m.Cache, h m.HttpClient, s m.Storage, path string, sec *secret.SecretsOperator, svc m.Service) *LuaEngine {
	engine := &LuaEngine{bot: b, cache: c, http: h, storage: s, BasePath: path, Secret: sec, service: svc}
	spy, err := helpers.NewScripts(path)
	if err != nil {
		logrus.Fatalf("ошибка загрузки скриптов %v", err)
	}

	engine.scripts = spy.Data
	return engine
}

func (le *LuaEngine) ExecuteScript(scriptPath string, lContext LuaContext) error {
	logrus.Infof("ExecuteScript path:%s", scriptPath)

	L := NewStateBuilder(le).
		WithModule(m.NewCache(le.cache)).
		WithModule(m.NewBot(le.bot, le.service)).
		WithModule(m.NewHttp(le.http)).
		WithModule(m.NewStorage(le.storage)).
		Build()

	defer L.Close()

	//заполняем контекст
	setLuaContext(L, &lContext)

	//ограничиваем по времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	L.SetContext(ctx)

	// Выполняем скрипт
	if _, ok := le.scripts[scriptPath]; ok {
		if err := L.DoString(string(le.scripts[scriptPath])); err != nil {
			return fmt.Errorf("lua error: %v", err)
		}
	} else {
		return fmt.Errorf("script didnt found %s", scriptPath)
	}

	return nil
}

func setLuaContext(L *lua.LState, lContext *LuaContext) {
	data := L.NewTable()

	// Базовые поля
	L.SetField(data, "chat_id", lua.LNumber(lContext.ChatId))
	L.SetField(data, "text", lua.LString(lContext.MessageText))

	// Обработка callback данных
	cbData := L.NewTable()
	L.SetField(cbData, "script", lua.LString(lContext.CbData.Script))
	L.SetField(cbData, "data", lua.LString(lContext.CbData.Data))
	L.SetField(data, "cb_data", cbData)

	// Прокидываем form_data как Lua таблицу
	if lContext.FormData != nil {
		formDataTable := convertMapToLuaTable(L, lContext.FormData)
		L.SetField(data, "form_data", formDataTable)
	}

	// Прокидываем request_data
	if lContext.RequestData != nil {
		reqDataTable := convertMapToLuaTable(L, lContext.RequestData)
		L.SetField(data, "req_data", reqDataTable)
	}

	// Информация о пользователе
	user := L.NewTable()
	L.SetField(user, "id", lua.LNumber(lContext.FromId))
	L.SetField(user, "name", lua.LString(lContext.FromName))
	L.SetField(data, "user", user)

	// Устанавливаем глобальную переменную ctx
	L.SetGlobal("ctx", data)
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
			L.SetField(tbl, k, h.ConvertToLuaTable(L, value))
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
					L.RawSetInt(arr, i+1, h.ConvertToLuaTable(L, elem))
				}
			}
			L.SetField(tbl, k, arr)
		}
	}
	return tbl
}
