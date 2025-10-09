package lua_modules

import (
	lua "github.com/yuin/gopher-lua"
)

// Core
func (m *CoreModule) Apply(L *lua.LState) {
	//(msg: string)
	m.applyLog(L, "log")

	//(data: table) -> (json: string)
	m.applyEncode(L, "json_encode")

	//(name: string) -> (secret: string)
	m.applySecrets(L, "reveal")
}

// Bot
func (m *BotModule) Apply(L *lua.LState) {
	//(chatId: int64, msg: string)
	m.applySendMessage(L, "send_message")

	//(chatId: int64, msg: string, keyboard: table) -> err?
	m.applySendKeyboard(L, "send_keyboard")
}

// Storage
func (m *StorageModule) Apply(L *lua.LState) {
	//(collection: string, data: table) -> (ok: bool, id: string)
	m.applyStorageCreate(L, "storage_save")

	//(collection: string, id: string) -> (data: table, err?)
	m.applyStorageGetById(L, "storage_load")
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
