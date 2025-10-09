package lua_modules

import (
	"github.com/end1essrage/indigo-core/storage"
	lua "github.com/yuin/gopher-lua"
)

func (m *StorageModule) applyQueryBuilder(L *lua.LState) {
	// Создаем метатаблицу для QueryNode
	mt := L.NewTypeMetatable("QueryNode")
	L.SetGlobal("QueryNode", mt)

	// Методы для QueryNode
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"and":      queryNodeAnd,
		"or":       queryNodeOr,
		"toString": queryNodeToString,
	}))

	// Глобальная функция для создания условий
	L.SetGlobal("query_condition", L.NewFunction(createCondition))

	// Глобальная функция для создания сложных запросов
	L.SetGlobal("query_and", L.NewFunction(createAndQuery))
	L.SetGlobal("query_or", L.NewFunction(createOrQuery))
}

// Создание условия: query_condition("field", "=", value)
func createCondition(L *lua.LState) int {
	field := L.CheckString(1)
	operator := L.CheckString(2)
	value := L.Get(3)

	// Конвертируем Lua значение в Go
	var goValue interface{}
	switch v := value.(type) {
	case *lua.LNilType:
		goValue = nil
	case lua.LBool:
		goValue = bool(v)
	case lua.LNumber:
		goValue = float64(v)
	case lua.LString:
		goValue = string(v)
	default:
		L.ArgError(3, "unsupported value type")
		return 0
	}

	condition := &storage.Condition{
		Field:    field,
		Operator: operator,
		Value:    goValue,
	}

	// Создаем userdata для условия
	ud := L.NewUserData()
	ud.Value = condition
	L.SetMetatable(ud, L.GetTypeMetatable("QueryNode"))
	L.Push(ud)
	return 1
}

// Создание AND запроса: query_and(left, right)
func createAndQuery(L *lua.LState) int {
	left := checkQueryNode(L, 1)
	right := checkQueryNode(L, 2)

	binaryOp := &storage.BinaryOp{
		Left:     left,
		Right:    right,
		Operator: "AND",
	}

	ud := L.NewUserData()
	ud.Value = binaryOp
	L.SetMetatable(ud, L.GetTypeMetatable("QueryNode"))
	L.Push(ud)
	return 1
}

// Создание OR запроса: query_or(left, right)
func createOrQuery(L *lua.LState) int {
	left := checkQueryNode(L, 1)
	right := checkQueryNode(L, 2)

	binaryOp := &storage.BinaryOp{
		Left:     left,
		Right:    right,
		Operator: "OR",
	}

	ud := L.NewUserData()
	ud.Value = binaryOp
	L.SetMetatable(ud, L.GetTypeMetatable("QueryNode"))
	L.Push(ud)
	return 1
}

// Метод для QueryNode:and(other)
func queryNodeAnd(L *lua.LState) int {
	self := checkQueryNode(L, 1)
	other := checkQueryNode(L, 2)

	binaryOp := &storage.BinaryOp{
		Left:     self,
		Right:    other,
		Operator: "AND",
	}

	ud := L.NewUserData()
	ud.Value = binaryOp
	L.SetMetatable(ud, L.GetTypeMetatable("QueryNode"))
	L.Push(ud)
	return 1
}

// Метод для QueryNode:or(other)
func queryNodeOr(L *lua.LState) int {
	self := checkQueryNode(L, 1)
	other := checkQueryNode(L, 2)

	binaryOp := &storage.BinaryOp{
		Left:     self,
		Right:    other,
		Operator: "OR",
	}

	ud := L.NewUserData()
	ud.Value = binaryOp
	L.SetMetatable(ud, L.GetTypeMetatable("QueryNode"))
	L.Push(ud)
	return 1
}

// Метод для QueryNode:toString()
func queryNodeToString(L *lua.LState) int {
	node := checkQueryNode(L, 1)
	L.Push(lua.LString(node.ToString()))
	return 1
}

// Вспомогательная функция для извлечения QueryNode из userdata
func checkQueryNode(L *lua.LState, n int) storage.QueryNode {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(storage.QueryNode); ok {
		return v
	}
	L.ArgError(n, "QueryNode expected")
	return nil
}
