package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServerShutdown(t *testing.T) {
	srv := &Server{srv: newHTTPServer("127.0.0.1:0", http.NotFoundHandler())}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.srv.Serve(listener) }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("Serve after Shutdown = %v, want http.ErrServerClosed", err)
		}
	case <-ctx.Done():
		t.Fatal("Serve did not stop after Shutdown")
	}
}

func TestHTTPServerHasTimeouts(t *testing.T) {
	srv := newHTTPServer(":8080", http.NotFoundHandler())

	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Fatalf("HTTP server timeouts are not configured: %+v", srv)
	}
	if srv.ReadHeaderTimeout > 10*time.Second {
		t.Errorf("ReadHeaderTimeout = %s, want a short Slowloris guard", srv.ReadHeaderTimeout)
	}
}
