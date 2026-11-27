package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTrustedOriginGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/instant", TrustedOriginGuard([]string{"http://localhost:5173", "https://only4.fun"}), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	call := func(origin, host string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/instant", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if host != "" {
			req.Host = host
		}
		r.ServeHTTP(w, req)
		return w.Code
	}

	cases := []struct {
		name   string
		origin string
		host   string
		want   int
	}{
		{"allowed frontend", "http://localhost:5173", "localhost:8080", http.StatusOK},
		{"allowed production", "https://only4.fun", "api.only4.fun", http.StatusOK},
		{"host shortcut removed", "http://localhost:8080", "localhost:8080", http.StatusForbidden},
		{"no origin (non-browser)", "", "localhost:8080", http.StatusOK},
		{"evil origin", "https://evil.example", "localhost:8080", http.StatusForbidden},
		{"attacker host pointed at backend", "http://attacker.example", "attacker.example", http.StatusForbidden},
		{"null origin", "null", "localhost:8080", http.StatusForbidden},
		{"lookalike host", "http://localhost:5173.evil.example", "localhost:8080", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := call(tc.origin, tc.host); got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRateLimitByIPUsesTrustedProxyHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := r.SetTrustedProxies([]string{"10.0.0.0/8"}); err != nil {
		t.Fatal(err)
	}
	r.RemoteIPHeaders = []string{"CF-Connecting-IP", "X-Forwarded-For"}
	r.POST("/instant", RateLimitByIP(1, time.Minute), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	call := func(peer, clientIP string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/instant", nil)
		req.RemoteAddr = peer
		if clientIP != "" {
			req.Header.Set("CF-Connecting-IP", clientIP)
		}
		r.ServeHTTP(w, req)
		return w.Code
	}

	if got := call("10.0.0.1:1111", "203.0.113.10"); got != http.StatusOK {
		t.Fatalf("first trusted request = %d, want 200", got)
	}
	if got := call("10.0.0.1:1111", "203.0.113.11"); got != http.StatusOK {
		t.Fatalf("second client should get a fresh bucket, got %d", got)
	}
	if got := call("10.0.0.1:1111", "203.0.113.10"); got != http.StatusTooManyRequests {
		t.Fatalf("same client should be rate limited, got %d", got)
	}
	if got := call("192.0.2.9:2222", "203.0.113.10"); got != http.StatusOK {
		t.Fatalf("spoofed header from untrusted peer must be ignored, got %d", got)
	}
}

func TestRateLimitByIPIgnoresForgedCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/instant", RateLimitByIP(1, time.Minute), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	call := func(cookie string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/instant", nil)
		req.RemoteAddr = "10.9.9.9:1234"
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: "session_id", Value: cookie})
		}
		r.ServeHTTP(w, req)
		return w.Code
	}

	if got := call("first-session"); got != http.StatusOK {
		t.Fatalf("first request = %d, want 200", got)
	}
	// Cookie berbeda tidak boleh membuka kuota baru pada endpoint publik.
	if got := call("another-session"); got != http.StatusTooManyRequests {
		t.Fatalf("rotated cookie request = %d, want 429", got)
	}
}
