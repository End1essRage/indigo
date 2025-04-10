package lua

import (
	"encoding/json"
	"fmt"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

// convertToLuaTable рекурсивно конвертирует Go-значения в Lua-структуры
func convertToLuaTable(L *lua.LState, value interface{}) *lua.LTable {
	tbl := L.NewTable()

	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			tbl.RawSetString(key, convertValue(L, val))
		}
	case []interface{}:
		for i, elem := range v {
			tbl.RawSetInt(i+1, convertValue(L, elem)) // В Lua индексы с 1
		}
	default:
		// Обрабатываем простые типы
		tbl.RawSetString("value", convertValue(L, value))
	}
	return tbl
}

// convertValue обрабатывает отдельные значения
func convertValue(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		return convertToLuaTable(L, v)
	case map[string]interface{}:
		return convertToLuaTable(L, v)
	case json.Number:
		if num, err := v.Float64(); err == nil {
			return lua.LNumber(num)
		}
		return lua.LString(v)
	default:
		// Попытка обработки через рефлексию для сложных типов
		return convertReflectedValue(L, v)
	}
}

// convertReflectedValue обрабатывает нестандартные типы
func convertReflectedValue(L *lua.LState, value interface{}) lua.LValue {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		tbl := L.NewTable()
		for i := 0; i < rv.Len(); i++ {
			tbl.RawSetInt(i+1, convertValue(L, rv.Index(i).Interface()))
		}
		return tbl
	case reflect.Map:
		tbl := L.NewTable()
		for _, key := range rv.MapKeys() {
			strKey := fmt.Sprintf("%v", key.Interface())
			tbl.RawSetString(strKey, convertValue(L, rv.MapIndex(key).Interface()))
		}
		return tbl
	case reflect.Struct:
		tbl := L.NewTable()
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if tag := field.Tag.Get("json"); tag != "" {
				tbl.RawSetString(tag, convertValue(L, rv.Field(i).Interface()))
			} else {
				tbl.RawSetString(field.Name, convertValue(L, rv.Field(i).Interface()))
			}
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", value))
	}
}

func luaTableToJSON(L *lua.LState, tbl *lua.LTable) ([]byte, error) {
	goValue := convertLuaValue(L, tbl)
	return json.Marshal(goValue)
}

func convertLuaValue(L *lua.LState, value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // объект
			result := make(map[string]interface{})
			v.ForEach(func(k, val lua.LValue) {
				result[k.String()] = convertLuaValue(L, val)
			})
			return result
		} else { // массив
			result := make([]interface{}, maxn)
			for i := 1; i <= maxn; i++ {
				result[i-1] = convertLuaValue(L, v.RawGetInt(i))
			}
			return result
		}
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	default:
		return nil
	}
}
