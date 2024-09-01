package server

import (
	"net/http"
	"log/slog"
	"context"
	"time"
	"fmt"
)

type Server struct {
	server *http.Server
}

type Config struct {
	Url          string        `yaml:"url"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

func New(handler http.Handler, cfg *Config) *Server {
	return &Server{
		server: &http.Server{
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			Handler:      handler,
			Addr:         cfg.Url,
		},
	}
}

func (s *Server) Start(logger *slog.Logger) error {
	var err error

	logger.Info(fmt.Sprintf("Server started at %s", s.server.Addr))

	go func() {
		err = s.server.ListenAndServe()
	}()

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}