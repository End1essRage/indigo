package lua

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	b "github.com/end1essrage/indigo-core/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

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

// работа с http
type HttpClient interface {
	Get(url string, headers map[string]string) ([]byte, int, error)
	Post(url string, body []byte, headers map[string]string) ([]byte, int, error)
	Fetch(method, url string, body []byte, headers map[string]string) ([]byte, int, error)
}

type Storage interface {
	Save(ctx context.Context, entityType string, id string, data interface{}) error
	Load(ctx context.Context, entityType string, id string, result interface{}) error
	//Exists(ctx context.Context, entityType EntityType, id string) (bool, error)
	//Delete(ctx context.Context, entityType EntityType, id string) error
}

type CoreModule struct{}
type CacheModule struct{}
type BotModule struct{}
type HttpModule struct{}
type StorageModule struct{}

func (m *CoreModule) Apply(le *LuaEngine, L *lua.LState) {
	//Логирование
	L.SetGlobal("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logrus.Warnf("[LUA] %s", msg)
		return 0
	}))
}

func (m *StorageModule) Apply(le *LuaEngine, L *lua.LState) {
	L.SetGlobal("storage_save", L.NewFunction(func(L *lua.LState) int {
		entityType := L.CheckString(1)
		id := L.CheckString(2)
		dataTable := L.CheckTable(3)

		data, err := LuaTableToJSON(L, dataTable)
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

		if err := le.storage.Save(context.Background(), entityType, id, jsonData); err != nil {
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
		if err := le.storage.Load(context.Background(), entityType, id, &result); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := convertToLuaTable(L, result)
		L.Push(tbl)
		return 1
	}))
}

// convertToLuaTable рекурсивно конвертирует Go-значения в Lua-структуры
func convertToLuaTable(L *lua.LState, value interface{}) *lua.LTable {
	tbl := L.NewTable()

	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			tbl.RawSetString(key, convertValue(L, val))
		}
	case []interface{}:
		for i, elem := range v {
			tbl.RawSetInt(i+1, convertValue(L, elem)) // В Lua индексы с 1
		}
	default:
		// Обрабатываем простые типы
		tbl.RawSetString("value", convertValue(L, value))
	}
	return tbl
}

// convertValue обрабатывает отдельные значения
func convertValue(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		return convertToLuaTable(L, v)
	case map[string]interface{}:
		return convertToLuaTable(L, v)
	case json.Number:
		if num, err := v.Float64(); err == nil {
			return lua.LNumber(num)
		}
		return lua.LString(v)
	default:
		// Попытка обработки через рефлексию для сложных типов
		return convertReflectedValue(L, v)
	}
}

// convertReflectedValue обрабатывает нестандартные типы
func convertReflectedValue(L *lua.LState, value interface{}) lua.LValue {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		tbl := L.NewTable()
		for i := 0; i < rv.Len(); i++ {
			tbl.RawSetInt(i+1, convertValue(L, rv.Index(i).Interface()))
		}
		return tbl
	case reflect.Map:
		tbl := L.NewTable()
		for _, key := range rv.MapKeys() {
			strKey := fmt.Sprintf("%v", key.Interface())
			tbl.RawSetString(strKey, convertValue(L, rv.MapIndex(key).Interface()))
		}
		return tbl
	case reflect.Struct:
		tbl := L.NewTable()
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if tag := field.Tag.Get("json"); tag != "" {
				tbl.RawSetString(tag, convertValue(L, rv.Field(i).Interface()))
			} else {
				tbl.RawSetString(field.Name, convertValue(L, rv.Field(i).Interface()))
			}
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", value))
	}
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

func LuaTableToJSON(L *lua.LState, tbl *lua.LTable) ([]byte, error) {
	goValue := convertLuaValue(L, tbl)
	return json.Marshal(goValue)
}

func convertLuaValue(L *lua.LState, value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // объект
			result := make(map[string]interface{})
			v.ForEach(func(k, val lua.LValue) {
				result[k.String()] = convertLuaValue(L, val)
			})
			return result
		} else { // массив
			result := make([]interface{}, maxn)
			for i := 1; i <= maxn; i++ {
				result[i-1] = convertLuaValue(L, v.RawGetInt(i))
			}
			return result
		}
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	default:
		return nil
	}
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

func (m *HttpModule) Apply(le *LuaEngine, L *lua.LState) {
	// GET запрос
	L.SetGlobal("http_get", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		headersTable := L.OptTable(2, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		body, status, err := le.http.Get(url, headers)
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

		respBody, status, err := le.http.Post(url, []byte(body), headers)
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

		respBody, status, err := le.http.Fetch(method, url, []byte(body), headers)
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
