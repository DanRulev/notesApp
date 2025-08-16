package server

import (
	"context"
	"net/http"
	"noteApp/internal/config"
)

type Server struct {
	server http.Server
}

func NewServer(cfg config.ServerCfg, handler http.Handler) *Server {
	return &Server{
		server: http.Server{
			Addr:           cfg.Host + ":" + cfg.Port,
			Handler:        handler,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			IdleTimeout:    cfg.IdleTimeout,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
