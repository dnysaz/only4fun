package server

import (
	"net/http"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
)

type Server struct {
	cfg            *config.Config
	db             *database.Database
	srv            *http.Server
	worker         *builder.Worker
	processManager *ProcessManager
}

func New(cfg *config.Config, db *database.Database, worker *builder.Worker) (*Server, *ProcessManager) {
	// Satu instance dipakai bersama router dan signal handler.
	pm := NewProcessManager()
	r := NewRouter(cfg, db, worker, pm)

	return &Server{
		cfg:            cfg,
		db:             db,
		srv:            newHTTPServer(cfg.Server.Addr, r),
		worker:         worker,
		processManager: pm,
	}, pm
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
}

func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *Server) GetProcessManager() *ProcessManager {
	return s.processManager
}
