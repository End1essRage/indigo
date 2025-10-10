package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type CbData struct {
	Script *string `json:"script,omitempty"`
	Data   *string `json:"data,omitempty"`
}

type TgBot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(b *tgbotapi.BotAPI) *TgBot {
	return &TgBot{bot: b}
}

func (t *TgBot) SendMessage(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)

	_, err := t.bot.Send(msg)

	return err
}

func (t *TgBot) Send(msg tgbotapi.MessageConfig) error {
	_, err := t.bot.Send(msg)

	return err
}

// автоматически если много кнопок реализовать переключалку через сервис
func (t *TgBot) SendKeyboard(chatId int64, text string, mesh MeshInlineKeyboard) error {
	msg := tgbotapi.NewMessage(chatId, text)
	logrus.Infof("%+v", mesh)
	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: CreateInlineKeyboard(mesh),
	}

	if err := t.Send(msg); err != nil {
		logrus.Errorf("Error sending keyboard: %v", err)
		return err
	}

	return nil
}

func (t *TgBot) DeleteMsg(chatId int64, msgId int) error {
	d := tgbotapi.NewDeleteMessage(chatId, msgId)
	_, err := t.bot.Send(d)

	return err
}
