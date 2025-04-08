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

type FormWorker struct {
	bot    *b.TgBot
	cache  Cache
	config *c.Config
	le     *l.LuaEngine
}

func NewFormWorker(bot *b.TgBot, cache Cache, config *c.Config, le *l.LuaEngine) *FormWorker {
	return &FormWorker{
		bot:    bot,
		cache:  cache,
		config: config,
		le:     le,
	}
}

func (fw *FormWorker) HasActiveForm(upd *tgbotapi.Update) bool {
	if upd.Message != nil {
		return fw.cache.Exists(fw.formKey(upd.Message.From.ID))
	}
	if upd.CallbackQuery != nil {
		return fw.cache.Exists(fw.formKey(upd.CallbackQuery.From.ID))
	}
	return false
}

func (fw *FormWorker) StartForm(formName string, userID int64, upd *tgbotapi.Update) error {
	form := fw.config.Forms[formName]
	if form == nil {
		return fmt.Errorf("form '%s' not found", formName)
	}

	fw.cache.SetString(fw.formKey(userID), formName)
	fw.cache.SetString(fw.progressKey(userID), "0")
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

	formName := fw.cache.GetString(fw.formKey(userID))
	if formName == "" {
		return
	}

	progress, _ := strconv.Atoi(fw.cache.GetString(fw.progressKey(userID)))
	form := fw.config.Forms[formName]

	if progress >= len(form.Stages) {
		fw.clearFormData(userID)
		return
	}

	currentStep := form.Stages[progress]

	if upd.Message != nil && input != "" {
		if !fw.validateInput(*currentStep.Validation, input) {
			fw.sendValidationError(userID)
			return
		}
	}

	fw.saveFormData(userID, currentStep.Field, input)

	if progress < len(form.Stages)-1 {
		fw.cache.SetString(fw.progressKey(userID), strconv.Itoa(progress+1))
		fw.sendFormStep(userID, progress+1, upd)
	} else {
		fw.completeForm(userID, form, upd)
	}
}

func (fw *FormWorker) sendFormStep(userID int64, stepIndex int, upd *tgbotapi.Update) error {
	formName := fw.cache.GetString(fw.formKey(userID))
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
	fw.cache.SetString(fw.dataKey(userID, field), value)
}

func (fw *FormWorker) collectFormData(userID int64) map[string]interface{} {
	data := make(map[string]interface{})
	formName := fw.cache.GetString(fw.formKey(userID))
	form := fw.config.Forms[formName]

	for _, stage := range form.Stages {
		if val := fw.cache.GetString(fw.dataKey(userID, stage.Field)); val != "" {
			data[stage.Field] = val
		}
	}
	return data
}

func (fw *FormWorker) clearFormData(userID int64) {
	fw.cache.SetString(fw.formKey(userID), "")
	fw.cache.SetString(fw.progressKey(userID), "")
}
