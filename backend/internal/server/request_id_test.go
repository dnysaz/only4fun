package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"only4.fun/backend/internal/id"

	"github.com/gin-gonic/gin"
)

func TestRequestIDIsGeneratedAndEchoed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("requestID"))
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !id.Valid(rec.Body.String()) {
		t.Fatalf("request ID = %q, want UUID v7", rec.Body.String())
	}
	if got := rec.Header().Get(requestIDHeader); got != rec.Body.String() {
		t.Fatalf("header request ID = %q, body = %q", got, rec.Body.String())
	}
}

func TestRequestIDKeepsSafeClientValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, c.GetString("requestID")) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(requestIDHeader, "trace-123")
	r.ServeHTTP(rec, req)
	if rec.Body.String() != "trace-123" {
		t.Fatalf("request ID = %q, want trace-123", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(requestIDHeader, "bad\nheader")
	r.ServeHTTP(rec, req)
	if !id.Valid(rec.Body.String()) {
		t.Fatalf("unsafe request ID was accepted: %q", rec.Body.String())
	}
}
