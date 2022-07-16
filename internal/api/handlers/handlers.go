package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"github.com/sergeysynergy/gophermart/internal/gophermart"
)

const (
	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextPlain       = "text/plain"
	LogLvlDebug                = "[DEBUG]"
	LogLvlInfo                 = "[INFO]"
	LogLvlWarning              = "[WARNING]"
	LogLvlError                = "[ERROR]"
	LogLvlFatal                = "[FATAL]"
)

type handler struct {
	r  chi.Router
	gm *gophermart.GopherMart
}

type Option func(*handler)

func New(gm *gophermart.GopherMart, opts ...Option) *handler {
	h := &handler{
		r:  chi.NewRouter(),
		gm: gm,
	}
	// применяем в цикле каждую опцию
	for _, opt := range opts {
		opt(h) // *Handler как аргумент
	}

	// Общая для всех роутеров миделвара
	h.r.Use(middleware.Compress(3, "gzip"))
	h.r.Use(middleware.RequestID)
	h.r.Use(middleware.RealIP)
	h.r.Use(middleware.Logger)
	h.r.Use(middleware.Recoverer)

	// Зададим роуты
	h.setRoutes()

	// Вернём измененный экземпляр Handler
	return h
}

func (h *handler) GetRouter() chi.Router {
	return h.r
}

func (h *handler) log(r *http.Request, lvl, msg string) {
	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		reqID = "[" + reqID + "] "
	}
	url := fmt.Sprintf(`"%s %s%s%s"`, r.Method, "http://", r.Host, r.URL)
	log.Printf("%s%s %s %s", reqID, lvl, url, msg)
}
