package server

import (
	"net/http"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
)

type Server struct {
	cfg           *config.Config
	db            *database.Database
	srv           *http.Server
	worker        *builder.Worker
	processManager *ProcessManager
}

func New(cfg *config.Config, db *database.Database, worker *builder.Worker) (*Server, *ProcessManager) {
	r := NewRouter(cfg, db, worker)
	pm := NewProcessManager()

	return &Server{
		cfg:           cfg,
		db:            db,
		srv:           &http.Server{Addr: cfg.Server.Addr, Handler: r},
		worker:        worker,
		processManager: pm,
	}, pm
}

func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *Server) GetProcessManager() *ProcessManager {
	return s.processManager
}
