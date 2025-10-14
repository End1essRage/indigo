package lua_modules

import (
	h "github.com/end1essrage/indigo-core/lua/helpers"
	"github.com/end1essrage/indigo-core/secret"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type CoreModule struct {
	secret *secret.SecretsOperator
}

func NewCore(secret *secret.SecretsOperator) *CoreModule {
	return &CoreModule{secret: secret}
}

func (m *CoreModule) applyLog(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		logrus.Warnf("[LUA] %s", msg)
		return 0
	}))
}

func (m *CoreModule) applyEncode(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		dataTable := L.CheckTable(1)

		data, err := h.LuaTableToJSON(dataTable)
		if err != nil {
			L.Push(lua.LString("{}"))
			return 1
		}
		L.Push(lua.LString(string(data)))

		return 1
	}))
}

func (m *CoreModule) applyDecode(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		dataTable := L.CheckString(1)

		data, err := h.JsonToLuaTable(L, []byte(dataTable))
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(data)

		return 1
	}))
}

func (m *CoreModule) applySecrets(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		name := L.ToString(1)
		sec := m.secret.RevealSecret(name)

		L.Push(lua.LString(sec))
		return 1
	}))
}
