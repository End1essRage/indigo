package lua

import (
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

type CoreModule struct{}
type CacheModule struct{}
type BotModule struct{}
type HttpModule struct{}

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
