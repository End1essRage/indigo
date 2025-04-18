package interceptor

import (
	i "github.com/end1essrage/indigo-core/interceptor"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func TrackUser() i.Interceptor {
	useFunc := func(upd *tgbotapi.Update) error {
		logrus.Warningf("track user")
		return nil
	}

	return i.New(useFunc)
}
