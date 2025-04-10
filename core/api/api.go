package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/end1essrage/indigo-core/config"
	"github.com/end1essrage/indigo-core/lua"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type API struct {
	router   *chi.Mux
	server   *http.Server
	le       *lua.LuaEngine
	config   *config.HTTPConfig
	stopping bool
	handling bool
	stopped  chan struct{}
	mu       sync.Mutex
}

func New(le *lua.LuaEngine, cfg *config.HTTPConfig) *API {
	r := chi.NewRouter()

	return &API{
		router:  r,
		le:      le,
		config:  cfg,
		stopped: make(chan struct{}),
	}
}

func (a *API) Start() error {
	// Регистрируем обработчики из конфига
	a.registerHandlers()

	a.server = &http.Server{
		Addr:    a.config.Address,
		Handler: a.router,
	}

	go func() {
		logrus.Infof("Starting API server on %s", a.config.Address)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("API server error: %v", err)
		}
	}()

	return nil
}

func (a *API) registerHandlers() {
	for _, endpoint := range a.config.Endpoints {
		scheme := a.findScheme(endpoint.Scheme)
		if scheme == nil {
			logrus.Errorf("Scheme %s not found for endpoint %s", endpoint.Scheme, endpoint.Path)
			continue
		}

		handler := a.createEndpointHandler(endpoint, scheme)

		switch strings.ToUpper(endpoint.Method) {
		case "GET":
			a.router.Get(endpoint.Path, handler)
		case "POST":
			a.router.Post(endpoint.Path, handler)
			// Добавить другие методы при необходимости
		}
	}
}

func (a *API) createEndpointHandler(endpoint config.Endpoint, scheme *config.Scheme) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		if a.stopping {
			a.mu.Unlock()
			a.stopped <- struct{}{}
			return
		}
		a.handling = true
		a.mu.Unlock()

		defer func() {
			a.mu.Lock()
			a.handling = false
			a.mu.Unlock()
		}()

		// Собираем данные по схеме
		data, err := a.collectRequestData(r, scheme)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Создаем Lua контекст
		ctx := lua.LuaContext{
			RequestData: data,
			Headers:     r.Header,
		}

		// Выполняем скрипт
		if err := a.le.ExecuteScript(endpoint.Script, ctx); err != nil {
			logrus.Errorf("Error executing script: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (a *API) collectRequestData(r *http.Request, scheme *config.Scheme) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	for _, field := range scheme.Fields {
		var value interface{}
		var err error

		switch field.Source {
		case "query":
			value = r.URL.Query().Get(field.Name)
		case "header":
			value = r.Header.Get(field.Name)
		case "body":
			value, err = a.parseBodyField(r, field)
		default:
			value = r.URL.Query().Get(field.Name)
		}

		if err != nil {
			return nil, err
		}

		if field.Required && value == nil {
			return nil, fmt.Errorf("missing required field: %s", field.Name)
		}

		data[field.Name] = value
	}

	return data, nil
}

func (a *API) parseBodyField(r *http.Request, field config.Field) (interface{}, error) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	value, exists := body[field.Name]
	if !exists && field.Required {
		return nil, fmt.Errorf("missing body field: %s", field.Name)
	}

	return value, nil
}

func (a *API) findScheme(name string) *config.Scheme {
	for _, s := range a.config.Schemes {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

func (a *API) Stop() {
	a.mu.Lock()
	a.stopping = true
	handling := a.handling
	a.mu.Unlock()

	if handling {
		select {
		case <-a.stopped:
		case <-time.After(5 * time.Second): // Таймаут на случай блокировки
		}
	}

	if a.server != nil {
		if err := a.server.Shutdown(context.Background()); err != nil {
			logrus.Errorf("API server shutdown error: %v", err)
		}
	}
}
