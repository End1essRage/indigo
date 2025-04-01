package main

import (
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type MeshKeyboard struct {
	Rows [][]MeshButton
}

type MeshButton struct {
	Name         string
	Text         string
	CustomCbData string
	Script       string
}

type CbData struct {
	Script *string `json:"script,omitempty"`
	Data   *string `json:"data,omitempty"`
}

func (b MeshButton) formatCbData() string {
	data := CbData{Data: &b.CustomCbData, Script: &b.Script}
	body, err := json.Marshal(data)
	if err != nil {
		logrus.Error("ошибка сериализации")
	}
	return fmt.Sprintf("%s", body)
}

func createKeyboard(mesh MeshKeyboard) [][]tgbotapi.InlineKeyboardButton {
	keyboard := make([][]tgbotapi.InlineKeyboardButton, 0)
	//проходимся по блоку Buttons по каждому Row
	for _, r := range mesh.Rows {
		row := make([]tgbotapi.InlineKeyboardButton, 0)
		//проходимся по кнопкам внутри Row
		for _, b := range r {
			//заполняем CallBackData
			btn := tgbotapi.NewInlineKeyboardButtonData(b.Text, b.formatCbData())
			row = append(row, btn)
		}
		keyboard = append(keyboard, row)
	}

	return keyboard
}
