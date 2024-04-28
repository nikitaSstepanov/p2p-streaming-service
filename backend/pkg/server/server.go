package server

import (
	"net/http"
	"log/slog"
	"context"
	"time"
	"fmt"
)

const (
	readTimeout  = 10 * time.Minute
	writeTimeout = 10 * time.Minute
)

type Server struct {
	server *http.Server
}

func New(handler http.Handler, url string) *Server {
	return &Server{
		server: &http.Server{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			Handler:      handler,
			Addr:         url,
		},
	}
}

func (s *Server) Start() error {
	var err error

	slog.Info(fmt.Sprintf("Server started at %s", s.server.Addr))

	go func() {
		err = s.server.ListenAndServe()
	}()

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
