package lua_modules

import (
	lua "github.com/yuin/gopher-lua"
)

type Service interface {
	GetChannelId(code string) (int64, error)
}

// Core
func (m *CoreModule) Apply(L *lua.LState) {
	//(msg: string)
	m.applyLog(L, "log")

	//(data: table) -> (json: string)
	m.applyEncode(L, "json_encode")

	//(json: string) -> (data: table)
	m.applyDecode(L, "json_decode")

	//(name: string) -> (secret: string)
	m.applySecrets(L, "reveal")
}

// Bot
func (m *BotModule) Apply(L *lua.LState) {
	//(chatId: int64, msg: string)
	//m.applySendMessage(L, "send_message")

	//(chatId: int64, msg: string, keyboard: table) -> err?
	m.applySend(L, "send")

	//(chan_code: string, msg: string) -> err?
	m.applySendChannel(L, "send_chan")
}

// Storage
func (m *StorageModule) Apply(L *lua.LState) {
	/*
		query_condition("field", "=", value)
		query_and(left, right)
		query_or(left, right)

		local query = query_condition("field", "=", value):and(query_condition("field2", "<", value2))
		query:or(query_condition("field3", ">", value3))
	*/
	m.applyQueryBuilder(L)

	// storage_get(collection, count, query)
	m.applyStorageGet(L, "storage_get")
	//(collection: string, id: string) -> (data: table, err?)
	m.applyStorageGetById(L, "storage_get_by_id")
	// storage_get_one(collection, query)
	m.applyStorageGetOne(L, "storage_get_one")
	// (collection, count, query)
	m.applyStorageGetIds(L, "storage_get_ids")

	//(collection: string, data: table) -> (ok: bool, id: string)
	m.applyStorageCreate(L, "storage_create")

	// storage_update(collection, query, data)
	m.applyStorageUpdate(L, "storage_update")
	// (collection, id, data)
	m.applyStorageUpdateById(L, "storage_update_by_id")

	// (collection, query)
	m.applyStorageDelete(L, "storage_delete")
	// (collection, id)
	m.applyStorageDeleteById(L, "storage_delete_by_id")
}

// Cache
func (m *CacheModule) Apply(L *lua.LState) {
	//(key: string) -> (value: string?)
	m.applyGet(L, "cache_get")

	//(key: string, value: string)
	m.applySet(L, "cache_set")
}

// Http
func (m *HttpModule) Apply(L *lua.LState) {
	// (url: string, headers: table) -> (resp: table?, err?)
	m.applyGet(L, "http_get")

	// (url: string, headers: table, body: string) -> (resp: table?, err?)
	m.applyPost(L, "http_post")

	// (method: string, url: string, body: string, headers: table) -> (resp: table?, err?)
	m.applyRequest(L, "http_do")
}
