package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"test_task/internal/file"
	"test_task/internal/pages"
)

type Server struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mux    *chi.Mux
	*http.Server
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func NewRestServer(ctx context.Context, addr string, logger *zap.SugaredLogger) *Server {
	service := file.NewService()
	mux := chi.NewRouter()

	httpSrv := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	serv := Server{
		ctx:    ctx,
		mux:    mux,
		Server: &httpSrv,
		logger: logger,
	}
	mux.With(serv.recoverer).Mount("/", pages.NewHandler(ctx, logger).Routes())
	mux.With(serv.recoverer).Mount("/api/file", file.NewHandler(ctx, logger, service).Routes())
	return &serv
}

func (s *Server) recoverer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Something going wrong..."))
				s.logger.Error("panic occurred:", err)
			}
		}()
		handler.ServeHTTP(writer, request)
	})
}
