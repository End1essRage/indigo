package server

import (
	"sync"
	"time"

	"github.com/end1essrage/indigo-core/api"
	b "github.com/end1essrage/indigo-core/bot"
	"github.com/end1essrage/indigo-core/config"
	c "github.com/end1essrage/indigo-core/config"
	h "github.com/end1essrage/indigo-core/handler"
	"github.com/end1essrage/indigo-core/interceptor"
	modules "github.com/end1essrage/indigo-core/interceptor/modules"
	l "github.com/end1essrage/indigo-core/lua"
	"github.com/end1essrage/indigo-core/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

//при использовании в личке chat.Id == From.Id

type Buffer interface {
	GetString(key string) string
	SetString(key string, val string) error
	Exists(key string) bool
}

type Server struct {
	le           *l.LuaEngine
	bot          *b.TgBot
	config       *c.Config
	api          *api.API
	formWorker   *h.FormWorker
	service      *service.Service
	interceptors map[c.AffectMode][]interceptor.Interceptor
	stopping     bool
	handling     bool
	stopped      chan struct{}
	mu           sync.Mutex
}

func NewServer(le *l.LuaEngine, bot *b.TgBot, config *c.Config, buffer Buffer, service *service.Service) *Server {
	s := &Server{
		le:           le,
		bot:          bot,
		config:       config,
		service:      service,
		formWorker:   h.NewFormWorker(bot, buffer, config, le),
		stopped:      make(chan struct{}),
		interceptors: registerInterceptors(config.Interceptors),
	}
	if s.config.HTTP != nil {
		s.api = api.New(s.le, s.config.HTTP)
	}
	return s
}

func registerInterceptors(inters []config.Interceptor) map[c.AffectMode][]interceptor.Interceptor {
	result := make(map[c.AffectMode][]interceptor.Interceptor)
	for _, inter := range inters {
		//массив перехватчиков
		arr := make([]interceptor.Interceptor, 0)

		for _, s := range inter.Scripts {
			arr = append(arr, interceptor.Script(s))
		}

		for _, f := range inter.Modules {
			switch {
			case f == string(c.TRACK_USER):
				arr = append(arr, modules.TrackUser())
			default:
				logrus.Errorf("не существует модуля %s", f)
			}
		}

		result[inter.Affects] = arr
	}

	return result
}

func (s *Server) Start(updates tgbotapi.UpdatesChannel) {
	go func() {
		for update := range updates {
			s.HandleUpdate(&update)
		}
	}()

	if s.config.HTTP != nil {
		go func() {
			if err := s.api.Start(); err != nil {
				logrus.Fatalf("Failed to start API: %v", err)
			}
		}()
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	s.stopping = true
	handling := s.handling
	s.mu.Unlock()

	if s.api != nil {
		s.api.Stop()
	}

	if handling {
		select {
		case <-s.stopped:
		case <-time.After(5 * time.Second):
		}
	}
}
