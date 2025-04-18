package interceptor

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Interceptor struct {
	useFunc func(upd *tgbotapi.Update) error
}

func New(f func(upd *tgbotapi.Update) error) Interceptor {
	return Interceptor{useFunc: f}
}

func (i Interceptor) Use(upd *tgbotapi.Update) error {
	return i.useFunc(upd)
}

func Script(scriptPath string) Interceptor {
	useFunc := func(upd *tgbotapi.Update) error {
		logrus.Warningf("script run")
		return nil
	}

	return Interceptor{useFunc: useFunc}
}
