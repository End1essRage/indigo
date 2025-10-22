package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/end1essrage/indigo-core/storage"
	s "github.com/end1essrage/indigo-core/storage"
	"github.com/sirupsen/logrus"
)

const channelAdmCollection = "channelAdm"
const channelKey = "channel_"

type BotAdminRequest struct {
	FromId    int64
	ChannelId int64
	Title     string
}

type ChannelInfo struct {
	ChanId int64
	Title  string
	Code   string
}

func (s *Service) HandleBotAdd(req BotAdminRequest) {
	code := genCode()
	//сохранить канал с каким-то кодо
	entity := storage.NewEntity()
	entity["chan_id"] = req.ChannelId
	entity["title"] = req.Title
	entity["code"] = code

	id, err := s.storage.Create(context.TODO(), channelAdmCollection, entity)
	if err != nil {
		logrus.Errorf("ошибка создания записи")
	}

	logrus.Infof("id документа %s", id)

	//отправить сообщение админу в личку тому кто добавил
	s.bot.SendMessage(req.FromId, fmt.Sprintf("Бот добавлен в канал %s с кодом %s ", req.Title, code))

	//Далее через админ меню - в виде клавиатуры(куда то надо засунуть в интернал ее опеределенние)
	//можно выбрать настройку каналов, получить список каналов в виде кнопок
	//при выборе канала появляется форма с одним шагом - ввести новое кодовое название канала

	//либо изменить генерацию кода для канала и позволить админам лишь просматривать его код
	//в таком случае админ меню содержит кнопки при нажатии кнопки с каналами мы просто получаем и выводим их список

	//как можно улучшить работу для админа, по сути нам нужен форм воркер который в конце не выполнит скрипт, а запустит го код и соберет данные формы
	//может позволить форм воркеру создаваться с колюэком в виде го функции?

	// проще всего сделать через форм воркер может отдельную реализацию или отдельный тип форм

	//считать заданный код и присвоить сохраненному каналу

	//отправлять в канал можно в луа по коду
}

func (s *Service) GetChannels() ([]s.Entity, error) {
	items, err := s.storage.Get(context.TODO(), channelAdmCollection, 0, nil)
	if err != nil {
		logrus.Errorf("ошибка получения каналов")
		return nil, err
	}

	return items, nil
}

func (s *Service) GetChannelId(code string) (int64, error) {
	//Проверить в кэше
	id := s.cache.GetString(channelKey + code)
	if id != "" {
		return strconv.ParseInt(id, 10, 64)
	}
	//достать с базы
	item, err := s.storage.GetOne(context.TODO(), channelAdmCollection, &storage.Condition{Field: "code", Operator: "=", Value: code})
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			logrus.Error("не найден канал по коду")
			return 0, err
		}
		return 0, err
	}

	chanId, ok := item["chan_id"]
	if !ok {
		return 0, fmt.Errorf("нет поля chan_id")
	}
	//обновитю кэш и вернуть
	s.cache.SetString(channelKey+code, strconv.FormatInt(chanId.(int64), 10))
	return chanId.(int64), nil
}

func (s *Service) HandleBotRemove(req BotAdminRequest) {
	//удалить из базы
	s.bot.SendMessage(req.FromId, fmt.Sprintf("Бот был удален из канала %s", req.Title))
}

func genCode() string {
	letters := "abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 3)
	for i := range bytes {
		// Генерируем случайное число от 0 до 25
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		// Выбираем соответствующий символ из набора
		bytes[i] = letters[num.Int64()]
	}

	return string(bytes)
}
