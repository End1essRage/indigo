package lua

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type LuaContext struct {
	RequestData map[string]interface{}
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
	Apply(le *LuaEngine, L *lua.LState)
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
	base := CoreModule{}
	base.Apply(b.le, L)

	// Кастомные модули
	for _, module := range b.modules {
		module.Apply(b.le, L)
	}

	return L
}

// Lua engine wrapper
type LuaEngine struct {
	bot     Bot
	cache   Cache
	http    HttpClient
	storage Storage
}

func NewLuaEngine(b Bot, c Cache, h HttpClient, s Storage) *LuaEngine {
	return &LuaEngine{bot: b, cache: c, http: h, storage: s}
}

func (le *LuaEngine) ExecuteScript(scriptPath string, lContext LuaContext) error {
	logrus.Infof("ExecuteScript path:%s", scriptPath)
	L := NewStateBuilder(le).WithModule(&CacheModule{}).WithModule(&BotModule{}).WithModule(&HttpModule{}).WithModule(&StorageModule{}).Build()
	defer L.Close()

	// Прокидываем контекст
	ctx := L.NewTable()

	L.SetField(ctx, "chat_id", lua.LNumber(lContext.ChatId))
	L.SetField(ctx, "text", lua.LString(lContext.MessageText))
	L.SetField(ctx, "cb_data", lua.LString(lContext.CbData.Data))
	user := L.NewTable()
	L.SetField(user, "id", lua.LNumber(lContext.FromId))
	L.SetField(user, "from_name", lua.LString(lContext.FromName))
	L.SetField(ctx, "user", user)

	L.SetGlobal("ctx", ctx)

	// Run script
	if err := L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("lua error: %v", err)
	}

	return nil
}
