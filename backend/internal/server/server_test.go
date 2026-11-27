package server

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPServerHasTimeouts(t *testing.T) {
	srv := newHTTPServer(":8080", http.NotFoundHandler())

	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Fatalf("HTTP server timeouts are not configured: %+v", srv)
	}
	if srv.ReadHeaderTimeout > 10*time.Second {
		t.Errorf("ReadHeaderTimeout = %s, want a short Slowloris guard", srv.ReadHeaderTimeout)
	}
}
