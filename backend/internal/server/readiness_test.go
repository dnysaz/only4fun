package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/secure"

	"github.com/gin-gonic/gin"
)

func TestReadinessHandlerReturnsReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := secure.Init("only4.fun-ready-test-secret"); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := database.Init(filepath.Join(root, "data", "only4.fun.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: filepath.Join(root, "data", "only4.fun.db")},
		Build: config.BuildConfig{
			SitesDir: filepath.Join(root, "sites"),
			CacheDir: filepath.Join(root, "cache"),
		},
	}
	r := gin.New()
	r.GET("/ready", ReadinessHandler(cfg, db, builder.NewWorker(nil, nil, nil, nil)))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestReadinessHandlerRejectsMissingWorker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ready", ReadinessHandler(&config.Config{}, nil, nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
