package lua

import (
	"encoding/json"

	b "github.com/end1essrage/indigo-core/bot"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

// Core
type CoreModule struct{}

func (m *CoreModule) Apply(L *lua.LState) {
	//Логирование
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logrus.Warnf("[LUA] %s", msg)
		return 0
	}))
}

// Хранилище
type Storage interface {
	Save(entityType string, id string, data interface{}) error
	Load(entityType string, id string, result interface{}) error
	//Exists(ctx context.Context, entityType EntityType, id string) (bool, error)
	//Delete(ctx context.Context, entityType EntityType, id string) error
}

type StorageModule struct{ storage Storage }

func (m *StorageModule) Apply(L *lua.LState) {
	L.SetGlobal("storage_save", L.NewFunction(func(L *lua.LState) int {
		entityType := L.CheckString(1)
		id := L.CheckString(2)
		dataTable := L.CheckTable(3)

		data, err := luaTableToJSON(L, dataTable)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if err := m.storage.Save(entityType, id, jsonData); err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("storage_load", L.NewFunction(func(L *lua.LState) int {
		entityType := L.CheckString(1)
		id := L.CheckString(2)

		var result map[string]interface{}
		if err := m.storage.Load(entityType, id, &result); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := convertToLuaTable(L, result)
		L.Push(tbl)
		return 1
	}))
}

// Кэш
type Cache interface {
	GetString(key string) string
	SetString(key string, val string)
}

type CacheModule struct{ cache Cache }

func (m *CacheModule) Apply(L *lua.LState) {
	L.SetGlobal("cache_get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := m.cache.GetString(key)
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
		m.cache.SetString(key, value)
		return 0
	}))
}

// Тг апи
type Bot interface {
	SendMessage(chatId int64, text string) error
	SendKeyboard(chatId int64, text string, mesh b.MeshInlineKeyboard) error
}
type BotModule struct{ bot Bot }

func (m *BotModule) Apply(L *lua.LState) {
	//Отправка сообщения
	L.SetGlobal("send_message", L.NewFunction(func(L *lua.LState) int {
		logrus.Infof("send_message")

		chatID := L.ToInt64(1)
		text := L.ToString(2)

		logrus.Infof("send_message ChatId: %v  text: %s", chatID, text)

		if err := m.bot.SendMessage(chatID, text); err != nil {
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
		mesh := b.FromLuaTableToMeshInlineKeyboard(meshTable)

		if err := m.bot.SendKeyboard(chatID, text, mesh); err != nil {
			logrus.Errorf("Error sending keyboard: %v", err)
			L.Push(lua.LString("send failed"))
			return 1
		}

		return 0
	}))
}

// работа с http
type HttpClient interface {
	Get(url string, headers map[string]string) ([]byte, int, error)
	Post(url string, body []byte, headers map[string]string) ([]byte, int, error)
	Fetch(method, url string, body []byte, headers map[string]string) ([]byte, int, error)
}

type HttpModule struct{ client HttpClient }

func (m *HttpModule) Apply(L *lua.LState) {
	// GET запрос
	L.SetGlobal("http_get", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		headersTable := L.OptTable(2, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		body, status, err := m.client.Get(url, headers)
		return pushHttpResponse(L, body, status, err)
	}))

	// POST запрос
	L.SetGlobal("http_post", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		body := L.CheckString(2)
		headersTable := L.OptTable(3, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		respBody, status, err := m.client.Post(url, []byte(body), headers)
		return pushHttpResponse(L, respBody, status, err)
	}))

	// запрос
	L.SetGlobal("http_do", L.NewFunction(func(L *lua.LState) int {
		method := L.CheckString(1)
		url := L.CheckString(2)
		body := L.CheckString(3)
		headersTable := L.OptTable(4, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		respBody, status, err := m.client.Fetch(method, url, []byte(body), headers)
		return pushHttpResponse(L, respBody, status, err)
	}))
}

func pushHttpResponse(L *lua.LState, body []byte, status int, err error) int {
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	tbl := L.NewTable()
	tbl.RawSetString("status", lua.LNumber(status))
	tbl.RawSetString("body", lua.LString(string(body)))
	L.Push(tbl)
	return 1
}
