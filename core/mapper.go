package main

import (
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type LuaContext struct {
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

func FromTgUpdateToLuaContext(update *tgbotapi.Update) LuaContext {
	c := LuaContext{}
	c.ChatId = update.Message.Chat.ID
	c.FromId = update.Message.From.ID
	c.FromName = update.Message.From.UserName
	c.MessageText = update.Message.Text
	return c
}

func FromCallbackQueryToLuaContext(cb *tgbotapi.CallbackQuery) LuaContext {
	c := LuaContext{}
	c.ChatId = cb.Message.Chat.ID
	c.FromId = cb.From.ID
	c.FromName = cb.From.UserName
	c.CbData = FromCallbackDataToLuaCbData(cb.Data)
	return c
}

func FromCallbackDataToLuaCbData(data string) LuaCbData {
	res := LuaCbData{}
	d := CbData{}
	if err := json.Unmarshal([]byte(data), &d); err != nil {
		logrus.Error("ошибка десериализции")
	}

	res.Script = *d.Script
	res.Data = *d.Data

	return res
}

// функция для конвертации Lua таблицы в MeshKeyboard
func FromLuaTableToMeshInlineKeyboard(L *lua.LState, lt *lua.LTable) MeshInlineKeyboard {
	var mesh MeshInlineKeyboard

	lt.ForEach(func(key lua.LValue, value lua.LValue) {
		if key.String() == "Rows" {
			if rows, ok := value.(*lua.LTable); ok {
				rows.ForEach(func(rowKey lua.LValue, rowValue lua.LValue) {
					if row, ok := rowValue.(*lua.LTable); ok {
						var meshRow []MeshInlineButton
						row.ForEach(func(btnKey lua.LValue, btnValue lua.LValue) {
							if btn, ok := btnValue.(*lua.LTable); ok {
								var meshBtn MeshInlineButton
								btn.ForEach(func(fieldKey lua.LValue, fieldValue lua.LValue) {
									switch fieldKey.String() {
									case "Text":
										meshBtn.Text = fieldValue.String()
									case "Script":
										meshBtn.Script = fieldValue.String()
									case "Data":
										meshBtn.CustomCbData = fieldValue.String()
									case "Name":
										meshBtn.Name = fieldValue.String()
									}
								})
								meshRow = append(meshRow, meshBtn)
							}
						})
						mesh.Rows = append(mesh.Rows, meshRow)
					}
				})
			}
		}
	})

	return mesh
}
