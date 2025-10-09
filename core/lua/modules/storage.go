package lua_modules

import (
	"context"
	"encoding/json"

	h "github.com/end1essrage/indigo-core/lua/helpers"
	"github.com/end1essrage/indigo-core/storage"
	lua "github.com/yuin/gopher-lua"
)

type Storage interface {
	GetById(ctx context.Context, collection string, id string) (storage.Entity, error)
	Create(ctx context.Context, collection string, entity storage.Entity) (string, error)
}

func (m *StorageModule) applyStorageCreate(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		dataTable := L.CheckTable(2)

		data, err := h.LuaTableToJSON(L, dataTable)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		var jsonData storage.Entity
		if err := json.Unmarshal(data, &jsonData); err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		id, err := m.storage.Create(context.TODO(), collection, jsonData)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		L.Push(lua.LString(id))
		return 1
	}))
}

func (m *StorageModule) applyStorageGetById(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		id := L.CheckString(2)

		result, err := m.storage.GetById(context.TODO(), collection, id)
		if err != nil {
			// Возвращаем пустую таблицу при ошибках
			L.Push(L.NewTable())
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Если файл отсутствовал или пустой
		if result == nil {
			L.Push(L.NewTable())
			return 1
		}

		tbl := h.ConvertToLuaTable(L, result)
		L.Push(tbl)
		return 1
	}))
}

type StorageModule struct{ storage Storage }

func NewStorage(storage Storage) *StorageModule {
	return &StorageModule{storage: storage}
}
