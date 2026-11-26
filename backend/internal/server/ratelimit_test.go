package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestContext(userID string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "10.0.0.1:12345"
	if userID != "" {
		c.Set("userID", userID)
	}
	return c
}

func TestRateLimitKeyPrefersUserID(t *testing.T) {
	if got := rateLimitKey(newTestContext("user-uuid-1")); got != "user:user-uuid-1" {
		t.Fatalf("rateLimitKey = %q, ingin %q", got, "user:user-uuid-1")
	}
	if got := rateLimitKey(newTestContext("")); got != "ip:10.0.0.1" {
		t.Fatalf("rateLimitKey tanpa user = %q, ingin %q", got, "ip:10.0.0.1")
	}
}

func TestRateLimitBlocksAfterLimitAndIsolatesUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", func(c *gin.Context) {
		if user := c.Query("user"); user != "" {
			c.Set("userID", user)
		}
	}, RateLimit(2, time.Minute), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	call := func(user string) *httptest.ResponseRecorder {
		target := "/x"
		if user != "" {
			target += "?user=" + user
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, target, nil)
		req.RemoteAddr = "10.0.0.2:4321"
		r.ServeHTTP(w, req)
		return w
	}

	for i := 0; i < 2; i++ {
		if w := call("u1"); w.Code != http.StatusOK {
			t.Fatalf("request ke-%d user u1 = %d, ingin 200", i+1, w.Code)
		}
	}
	blocked := call("u1")
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("request ke-3 user u1 = %d, ingin 429", blocked.Code)
	}
	if blocked.Header().Get("Retry-After") == "" {
		t.Fatalf("header Retry-After tidak dikirim saat limit tercapai")
	}

	// User lain di IP yang sama harus punya kuota sendiri.
	if w := call("u2"); w.Code != http.StatusOK {
		t.Fatalf("request user u2 = %d, ingin 200 (kuota terpisah per user)", w.Code)
	}
}
