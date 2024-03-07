package server

import (
	"net/http"
	"time"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
)

type Server struct {
	server *http.Server
}

func New(handler http.Handler, url string) *Server {
	
	httpServer := &http.Server{
		ReadTimeout: readTimeout,
		WriteTimeout: writeTimeout,
		Handler: handler,
		Addr: url,
	}

	server := &Server {
		server: httpServer,
	}

	return server
}

func (s *Server) Start() error {

	err := s.server.ListenAndServe()

	return err
}
