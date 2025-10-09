package lua

import (
	"net/http"

	m "github.com/end1essrage/indigo-core/lua/modules"
	lua "github.com/yuin/gopher-lua"
)

type LuaContext struct {
	RequestData map[string]interface{}
	FormData    map[string]interface{}
	Headers     http.Header
	MessageText string
	CbData      LuaCbData
	ChatId      int64
	FromId      int64
	FromName    string
}

type LuaCbData struct {
	Script string
	Data   string
}

type Module interface {
	Apply(L *lua.LState)
}

// LuaStateBuilder для конфигурации стейта под конкретный скрипт
type LuaStateBuilder struct {
	modules []Module
	le      *LuaEngine
}

func NewStateBuilder(engine *LuaEngine) *LuaStateBuilder {
	return &LuaStateBuilder{
		le: engine,
	}
}

func (b *LuaStateBuilder) WithModule(m Module) *LuaStateBuilder {
	b.modules = append(b.modules, m)
	return b
}

func (b *LuaStateBuilder) Build() *lua.LState {
	L := lua.NewState()

	// Базовые модули
	base := m.NewCore(b.le.Secret)
	base.Apply(L)

	// Кастомные модули
	for _, module := range b.modules {
		module.Apply(L)
	}

	return L
}
