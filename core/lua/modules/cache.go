package lua_modules

import lua "github.com/yuin/gopher-lua"

type Cache interface {
	GetString(key string) string
	SetString(key string, val string) error
}

func (m *CacheModule) applyGet(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)

		value := m.cache.GetString(key)
		if value == "" {
			L.Push(lua.LNil)
		} else {
			L.Push(lua.LString(value))
		}

		return 0
	}))
}

func (m *CacheModule) applySet(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)

		m.cache.SetString(key, value)
		return 0
	}))
}

type CacheModule struct{ cache Cache }

func NewCache(cache Cache) *CacheModule {
	return &CacheModule{cache: cache}
}
