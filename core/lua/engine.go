package lua

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

// Lua engine wrapper
type LuaEngine struct {
	bot      Bot
	cache    Cache
	http     HttpClient
	storage  Storage
	BasePath string
	scripts  map[string][]byte
}

func NewLuaEngine(b Bot, c Cache, h HttpClient, s Storage, path string) *LuaEngine {
	engine := &LuaEngine{bot: b, cache: c, http: h, storage: s, BasePath: path}
	buffer, err := LoadScripts(path)
	if err != nil {
		logrus.Fatalf("ошибка загрузки скриптов %v", err)
	}

	engine.scripts = buffer
	return engine
}

func LoadScripts(p string) (map[string][]byte, error) {
	buffer := make(map[string][]byte)

	dir, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		info, err := f.Info()
		if err != nil {
			return nil, err
		}

		if _, exists := buffer[info.Name()]; exists {
			return nil, fmt.Errorf("name conflict between file and directory: %s", info.Name())
		}

		sPath := filepath.Join(p, info.Name())
		//рекурсивно обрабатываем
		if info.IsDir() {
			logrus.Info("зашел в подпапку")
			innerData, err := LoadScripts(sPath)
			if err != nil {
				return nil, err
			}
			//заполняем
			for k, v := range innerData {
				buffer[filepath.Join(info.Name(), k)] = v
			}

			return buffer, nil
		}

		//пропускаем все что не луа
		shards := strings.Split(info.Name(), ".")
		if shards[len(shards)-1] != "lua" {
			logrus.Warnf("найден файл неправильного формата %s", info.Name())
			continue
		}

		data, err := os.ReadFile(sPath)
		if err != nil {
			return nil, err
		}

		buffer[info.Name()] = data
	}

	return buffer, nil
}

func (le *LuaEngine) ExecuteScripts(scriptPaths []string, lContext LuaContext) error {

	L := NewStateBuilder(le).
		WithModule(&CacheModule{cache: le.cache}).
		WithModule(&BotModule{bot: le.bot}).
		WithModule(&HttpModule{client: le.http}).
		WithModule(&StorageModule{storage: le.storage}).
		Build()
	defer L.Close()

	//заполняем контекст
	setLuaContext(L, &lContext)

	// Выполняем скрипт
	for _, scriptPath := range scriptPaths {
		if _, ok := le.scripts[scriptPath]; ok {
			if err := L.DoString(string(le.scripts[scriptPath])); err != nil {
				return fmt.Errorf("lua error: %v", err)
			}
		} else {
			//TODO убрать это так как обновления скрипта не будет происходить, либо надо реализовать отслеживание
			//try file
			script, err := os.ReadFile(filepath.Join(le.BasePath, scriptPath))
			if err != nil {
				return fmt.Errorf("error readinq script from file: %v", err)
			}

			if err := L.DoString(string(script)); err != nil {
				return fmt.Errorf("lua error: %v", err)
			}

			// сохраняем если все норм
			le.scripts[scriptPath] = script
		}
	}

	return nil
}

func (le *LuaEngine) ExecuteScript(scriptPath string, lContext LuaContext) error {
	logrus.Infof("ExecuteScript path:%s", scriptPath)

	L := NewStateBuilder(le).
		WithModule(&CacheModule{cache: le.cache}).
		WithModule(&BotModule{bot: le.bot}).
		WithModule(&HttpModule{client: le.http}).
		WithModule(&StorageModule{storage: le.storage}).
		Build()
	defer L.Close()

	//заполняем контекст
	setLuaContext(L, &lContext)

	// Выполняем скрипт
	if _, ok := le.scripts[scriptPath]; ok {
		if err := L.DoString(string(le.scripts[scriptPath])); err != nil {
			return fmt.Errorf("lua error: %v", err)
		}
	} else {
		//try file
		script, err := os.ReadFile(filepath.Join(le.BasePath, scriptPath))
		if err != nil {
			return fmt.Errorf("error readinq script from file: %v", err)
		}

		if err := L.DoString(string(script)); err != nil {
			return fmt.Errorf("lua error: %v", err)
		}

		le.scripts[scriptPath] = script
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
