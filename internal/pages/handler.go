package pages

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"net/http"
)

type handler struct {
	ctx context.Context
	log *zap.SugaredLogger
}

func NewHandler(ctx context.Context, log *zap.SugaredLogger) *handler {
	return &handler{ctx: ctx, log: log}
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/main", h.MainPage)

	return r
}

func (h *handler) MainPage(writer http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFiles("resources/index.html")
	if err != nil {
		h.log.Error(fmt.Errorf("MainPage failed during template parse: %w", err))
		http.Error(writer, "Error setting the file size", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(writer, nil)
	if err != nil {
		h.log.Error(fmt.Errorf("MainPage failed during template execute: %w", err))
		http.Error(writer, "Error setting the file size", http.StatusInternalServerError)
		return
	}
}
