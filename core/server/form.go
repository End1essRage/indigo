package server

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	b "github.com/end1essrage/indigo-core/bot"
	c "github.com/end1essrage/indigo-core/config"
	l "github.com/end1essrage/indigo-core/lua"
	m "github.com/end1essrage/indigo-core/mapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

//Добавить описание
//Возможность сбрасывать форму и канселить контекст дергая очистку
//создать буффер с методами для генерации ключей форм, так в едином месте можно будет увидеть все ключи с которыми работаем
//в основе буфера будет инмемори

//сейчас все крепится на юзер айди, формы в глобал чатах стоит запретить

//Надо улучшить механиз удаления сообщений, записывая их в буфер, стоит создать временные сообщения и сообщения этапа
//например ошибки валидации надо удалять после переввода пользователем, вопросы наверное тоже стоит удалять автоматически

type FormWorker struct {
	bot    *b.TgBot
	buffer Buffer
	config *c.Config
	le     *l.LuaEngine
}

func NewFormWorker(bot *b.TgBot, buffer Buffer, config *c.Config, le *l.LuaEngine) *FormWorker {
	return &FormWorker{
		bot:    bot,
		buffer: buffer,
		config: config,
		le:     le,
	}
}

func (fw *FormWorker) HasActiveForm(upd *tgbotapi.Update) bool {
	if upd.Message != nil {
		return fw.buffer.Exists(fw.formKey(upd.Message.From.ID))
	}
	if upd.CallbackQuery != nil {
		return fw.buffer.Exists(fw.formKey(upd.CallbackQuery.From.ID))
	}
	return false
}

func (fw *FormWorker) StartForm(formName string, userID int64, upd *tgbotapi.Update) error {
	form := fw.config.Forms[formName]
	if form == nil {
		return fmt.Errorf("form '%s' not found", formName)
	}

	fw.buffer.SetString(fw.formKey(userID), formName)
	fw.buffer.SetString(fw.progressKey(userID), "0")
	return fw.sendFormStep(userID, 0, upd)
}

func (fw *FormWorker) HandleInput(upd *tgbotapi.Update) {
	var userID int64
	var input string

	switch {
	case upd.Message != nil:
		userID = upd.Message.From.ID
		input = upd.Message.Text
	case upd.CallbackQuery != nil:
		userID = upd.CallbackQuery.From.ID
		input = m.FromCallbackDataToLuaCbData(upd.CallbackQuery.Data).Data

		fw.bot.DeleteMsg(upd.CallbackQuery.Message.Chat.ID, upd.CallbackQuery.Message.MessageID)
	default:
		return
	}

	formName := fw.buffer.GetString(fw.formKey(userID))
	if formName == "" {
		return
	}

	progress, _ := strconv.Atoi(fw.buffer.GetString(fw.progressKey(userID)))
	form := fw.config.Forms[formName]

	if progress >= len(form.Stages) {
		fw.clearFormData(userID)
		return
	}

	currentStep := form.Stages[progress]

	//ожидалось нажатие кнопки но его не рпоизошло
	if currentStep.Keyboard != nil && upd.CallbackQuery == nil {
		fw.sendValidationError(userID)
		return
	}

	if upd.Message != nil && input != "" && currentStep.Validation != nil {
		if !fw.validateInput(*currentStep.Validation, input) {
			fw.sendValidationError(userID)
			return
		}
	}

	fw.saveFormData(userID, currentStep.Field, input)

	if progress < len(form.Stages)-1 {
		fw.buffer.SetString(fw.progressKey(userID), strconv.Itoa(progress+1))
		fw.sendFormStep(userID, progress+1, upd)
	} else {
		fw.completeForm(userID, form, upd)
	}
}

func (fw *FormWorker) sendFormStep(userID int64, stepIndex int, upd *tgbotapi.Update) error {
	formName := fw.buffer.GetString(fw.formKey(userID))
	form := fw.config.Forms[formName]
	step := form.Stages[stepIndex]

	msg := tgbotapi.NewMessage(userID, step.Message)

	// Handle keyboard
	if step.Keyboard != nil && *step.Keyboard != "" {
		kb := fw.config.Keyboards[*step.Keyboard]
		if kb == nil {
			return fmt.Errorf("keyboard '%s' not found", *step.Keyboard)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			b.CreateInlineKeyboard(b.ParseInlineKeyboard(kb))...,
		)
		msg.ReplyMarkup = &keyboard
	}

	// Execute step script
	if step.Script != nil && *step.Script != "" {
		ctx := m.FromTgUpdateToLuaContext(upd)
		ctx.FormData = fw.collectFormData(userID)
		if err := fw.le.ExecuteScript(*step.Script, ctx); err != nil {
			logrus.Errorf("Form step script error: %v", err)
		}
	}

	err := fw.bot.Send(msg)
	return err
}

func (fw *FormWorker) completeForm(userID int64, form *c.Form, upd *tgbotapi.Update) {
	data := fw.collectFormData(userID)

	// Execute completion script
	if form.Script != "" {
		var ctx l.LuaContext

		if upd.CallbackQuery != nil {
			ctx = m.FromCallbackQueryToLuaContext(upd.CallbackQuery)
		} else {
			ctx = m.FromTgUpdateToLuaContext(upd)
		}

		ctx.FormData = data
		if err := fw.le.ExecuteScript(form.Script, ctx); err != nil {
			logrus.Errorf("Form completion script error: %v", err)
		}
	}

	fw.clearFormData(userID)
}

func (fw *FormWorker) validateInput(rules map[string]any, value string) bool {
	if rules == nil {
		return true
	}

	switch rules["type"].(string) {
	case "string":
		if min, ok := rules["min_length"].(int); ok {
			return len(value) >= min
		}
	case "email":
		return strings.Contains(value, "@") && strings.Contains(value, ".")
	case "number":
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case "regex":
		if pattern, ok := rules["pattern"].(string); ok {
			matched, _ := regexp.MatchString(pattern, value)
			return matched
		}
	}
	return true
}

func (fw *FormWorker) sendValidationError(userID int64) {
	msg := tgbotapi.NewMessage(userID, "Validation error")
	fw.bot.Send(msg)
}

// Helper methods
func (fw *FormWorker) formKey(userID int64) string {
	return fmt.Sprintf("form:%d", userID)
}

func (fw *FormWorker) progressKey(userID int64) string {
	return fmt.Sprintf("form_progress:%d", userID)
}

func (fw *FormWorker) dataKey(userID int64, field string) string {
	return fmt.Sprintf("form_data:%d:%s", userID, field)
}

func (fw *FormWorker) saveFormData(userID int64, field, value string) {
	fw.buffer.SetString(fw.dataKey(userID, field), value)
}

func (fw *FormWorker) collectFormData(userID int64) map[string]interface{} {
	data := make(map[string]interface{})
	formName := fw.buffer.GetString(fw.formKey(userID))
	form := fw.config.Forms[formName]

	for _, stage := range form.Stages {
		if val := fw.buffer.GetString(fw.dataKey(userID, stage.Field)); val != "" {
			data[stage.Field] = val
		}
	}
	return data
}

func (fw *FormWorker) clearFormData(userID int64) {
	fw.buffer.SetString(fw.formKey(userID), "")
	fw.buffer.SetString(fw.progressKey(userID), "")
}
