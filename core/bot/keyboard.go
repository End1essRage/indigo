package bot

import (
	"encoding/json"
	"fmt"

	c "github.com/end1essrage/indigo-core/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type MeshReplyKeyboard struct {
	Rows [][]MeshReplyButton
}

type MeshReplyButton struct {
	Text string
}

type MeshInlineKeyboard struct {
	Rows [][]MeshInlineButton
}

type MeshInlineButton struct {
	Name         string
	Text         string
	CustomCbData string
	Script       string
}

func (b MeshInlineButton) formatCbData() string {
	data := CbData{Data: &b.CustomCbData, Script: &b.Script}
	body, err := json.Marshal(data)
	if err != nil {
		logrus.Error("ошибка сериализации")
	}
	return fmt.Sprintf("%s", body)
}

func CreateInlineKeyboard(mesh MeshInlineKeyboard) [][]tgbotapi.InlineKeyboardButton {
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

func ParseInlineKeyboard(kb *c.Keyboard) MeshInlineKeyboard {
	kbMesh := MeshInlineKeyboard{}

	//проходимся по блоку Buttons по каждому Row
	for _, r := range *kb.Buttons {
		row := make([]MeshInlineButton, 0)
		//проходимся по кнопкам внутри Row
		for _, b := range r.Row {
			btn := MeshInlineButton{Text: b.Text}
			//заполняем CallBackData
			if b.Script != nil {
				btn.Script = *b.Script
			}
			if b.Data != nil {
				btn.CustomCbData = *b.Data
			}

			row = append(row, btn)
		}
		kbMesh.Rows = append(kbMesh.Rows, row)
	}

	return kbMesh
}
