package lua_modules

import (
	"context"
	"encoding/json"

	h "github.com/end1essrage/indigo-core/lua/helpers"
	"github.com/end1essrage/indigo-core/storage"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type Storage interface {
	Get(ctx context.Context, collection string, count int, query storage.QueryNode) ([]storage.Entity, error)
	GetIds(ctx context.Context, collection string, count int, query storage.QueryNode) ([]string, error)
	GetOne(ctx context.Context, collection string, query storage.QueryNode) (storage.Entity, error)
	GetById(ctx context.Context, collection string, id string) (storage.Entity, error)

	Create(ctx context.Context, collection string, entity storage.Entity) (string, error)

	UpdateById(ctx context.Context, collection string, id string, entity storage.Entity) error
	Update(ctx context.Context, collection string, query storage.QueryNode, entity storage.Entity) (int, error)

	DeleteById(ctx context.Context, collection string, id string) error
	Delete(ctx context.Context, collection string, query storage.QueryNode) (int, error)
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
		return 2
	}))
}

func (m *StorageModule) applyStorageGetById(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		id := L.CheckString(2)

		result, err := m.storage.GetById(context.TODO(), collection, id)
		if err != nil {
			// Возвращаем пустую таблицу при ошибках
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Если файл отсутствовал или пустой
		if result == nil {
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			return 2
		}

		logrus.Debugf("get by id result %+v", result)
		tbl := h.ConvertToLuaTable(L, result)
		logrus.Debugf("get by id result table %+v", tbl)

		L.Push(tbl)
		L.Push(lua.LNil)
		return 2
	}))
}

// storage_get(collection, count, query)
func (m *StorageModule) applyStorageGet(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		count := L.CheckInt(2)

		var query storage.QueryNode
		if L.GetTop() >= 3 {
			query = checkQueryNode(L, 3)
		}

		results, err := m.storage.Get(context.TODO(), collection, count, query)
		if err != nil {
			L.Push(L.NewTable())
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Конвертируем результаты в Lua таблицу
		tbl := L.NewTable()
		for i, entity := range results {
			tbl.RawSetInt(i+1, h.ConvertToLuaTable(L, entity))
		}
		L.Push(tbl)
		return 1
	}))
}

// storage_get_one(collection, query)
func (m *StorageModule) applyStorageGetOne(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)

		var query storage.QueryNode
		if L.GetTop() >= 2 {
			query = checkQueryNode(L, 2)
		}

		result, err := m.storage.GetOne(context.TODO(), collection, query)
		if err != nil {
			logrus.Errorf("[ENGINE] error GetOne %s", err.Error())
			L.Push(L.NewTable())
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if result == nil {
			logrus.Debug("не найдено")
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			return 2
		}

		L.Push(h.ConvertToLuaTable(L, result))
		L.Push(lua.LNil)
		return 2
	}))
}

// storage_get_ids(collection, count, query)
func (m *StorageModule) applyStorageGetIds(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		count := L.CheckInt(2)

		var query storage.QueryNode
		if L.GetTop() >= 3 {
			query = checkQueryNode(L, 3)
		}

		ids, err := m.storage.GetIds(context.TODO(), collection, count, query)
		if err != nil {
			L.Push(L.NewTable())
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Конвертируем ID в Lua таблицу
		tbl := L.NewTable()
		for i, id := range ids {
			tbl.RawSetInt(i+1, lua.LString(id))
		}
		L.Push(tbl)
		return 1
	}))
}

// storage_update(collection, query, data)
func (m *StorageModule) applyStorageUpdate(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		query := checkQueryNode(L, 2)
		dataTable := L.CheckTable(3)

		data, err := h.LuaTableToJSON(L, dataTable)
		if err != nil {
			L.Push(lua.LNumber(0))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		var jsonData storage.Entity
		if err := json.Unmarshal(data, &jsonData); err != nil {
			L.Push(lua.LNumber(0))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		count, err := m.storage.Update(context.TODO(), collection, query, jsonData)
		if err != nil {
			L.Push(lua.LNumber(0))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNumber(count))
		return 1
	}))
}

// storage_update_by_id(collection, id, data)
func (m *StorageModule) applyStorageUpdateById(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		id := L.CheckString(2)
		dataTable := L.CheckTable(3)

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

		err = m.storage.UpdateById(context.TODO(), collection, id, jsonData)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))
}

// storage_delete(collection, query)
func (m *StorageModule) applyStorageDelete(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		query := checkQueryNode(L, 2)

		count, err := m.storage.Delete(context.TODO(), collection, query)
		if err != nil {
			L.Push(lua.LNumber(0))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNumber(count))
		return 1
	}))
}

// storage_delete_by_id(collection, id)
func (m *StorageModule) applyStorageDeleteById(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		collection := L.CheckString(1)
		id := L.CheckString(2)

		err := m.storage.DeleteById(context.TODO(), collection, id)
		if err != nil {
			logrus.Errorf("[ENGINE] %s", err.Error())

			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		L.Push(lua.LString(""))
		return 2
	}))
}

type StorageModule struct{ storage Storage }

func NewStorage(storage Storage) *StorageModule {
	return &StorageModule{storage: storage}
}
