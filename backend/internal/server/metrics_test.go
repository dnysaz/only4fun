package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"

	"github.com/gin-gonic/gin"
)

func TestMetricsAllowedOnlyFromLoopback(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "::1"} {
		if !metricsAllowed(ip) {
			t.Errorf("metricsAllowed(%q) = false, want true", ip)
		}
	}
	for _, ip := range []string{"10.0.0.5", "203.0.113.9", "", "not-an-ip"} {
		if metricsAllowed(ip) {
			t.Errorf("metricsAllowed(%q) = true, want false", ip)
		}
	}
}

func TestMetricsConcurrentRecording(t *testing.T) {
	var m metrics
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				m.record(j%5 == 0, time.Now())
			}
		}(i)
	}
	wg.Wait()

	requests, errors, last := m.snapshot()
	if requests != 800 {
		t.Fatalf("requests = %d, want 800", requests)
	}
	if errors != 160 {
		t.Fatalf("errors = %d, want 160", errors)
	}
	if last.IsZero() {
		t.Fatal("last request time was not recorded")
	}
}

func TestMetricsHandlerIncludesBuildAndDiskStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	cfg := &config.Config{Build: config.BuildConfig{
		SitesDir: filepath.Join(root, "sites"),
		CacheDir: filepath.Join(root, "cache"),
	}}
	worker := builder.NewWorker(nil, nil, nil, nil)
	if err := worker.Enqueue(builder.Job{ProjectID: "p1"}); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/metrics", MetricsHandler(cfg, worker))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec.Body.String()
	for _, want := range []string{
		"build_queued 1",
		"build_active 0",
		"build_total 0",
		"disk_free_sites_bytes ",
		"disk_free_cache_bytes ",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics missing %q:\n%s", want, body)
		}
	}
}

func TestMetricTextIsSingleLine(t *testing.T) {
	if got := metricText("boom\nwith\tdetails"); got != "boom with details" {
		t.Fatalf("metricText = %q", got)
	}
}

func TestDiskFreeBytesUsesNearestExistingDir(t *testing.T) {
	root := t.TempDir()
	free, err := diskFreeBytes(filepath.Join(root, "not-created", "sites"))
	if err != nil {
		t.Fatalf("diskFreeBytes: %v", err)
	}
	if free == 0 {
		t.Fatal("diskFreeBytes = 0, want available space")
	}
}
