package server

import (
	"net/http"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
)

type Server struct {
	cfg *config.Config
	db  *database.Database
	srv *http.Server
}

func New(cfg *config.Config, db *database.Database, worker *builder.Worker) *Server {
	router := NewRouter(cfg, db, worker)

	return &Server{
		cfg: cfg,
		db:  db,
		srv: &http.Server{
			Addr:         cfg.Server.Addr,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}
