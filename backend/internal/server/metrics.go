package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"

	"github.com/gin-gonic/gin"
)

// metrics menyimpan metrik dasar untuk endpoint /metrics. Counter memakai
// atomic agar request yang berjalan bersamaan tidak saling menimpa.
type metrics struct {
	requests    atomic.Int64
	errors      atomic.Int64
	lastReqNano atomic.Int64
}

func (m *metrics) record(failed bool, at time.Time) {
	if failed {
		m.errors.Add(1)
	}
	m.requests.Add(1)
	m.lastReqNano.Store(at.UnixNano())
}

func (m *metrics) snapshot() (requests int64, errors int64, last time.Time) {
	if nano := m.lastReqNano.Load(); nano > 0 {
		last = time.Unix(0, nano)
	}
	return m.requests.Load(), m.errors.Load(), last
}

// Metrics menyimpan metrik proses server.
var Metrics = &metrics{}

// metricsAllowed membatasi /metrics untuk pemantauan lokal. Scraper di VPS
// atau sidecar memakai loopback; publik tidak boleh melihat angka internal.
func metricsAllowed(clientIP string) bool {
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	return ip != nil && ip.IsLoopback()
}

// nearestExistingDir memakai parent terdekat yang sudah ada, supaya endpoint
// metrics tidak perlu membuat folder hanya untuk membaca statfs.
func nearestExistingDir(path string) string {
	for path != "" && path != "." {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return "."
}

func diskFreeBytes(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(nearestExistingDir(path), &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}

func metricText(value string) string {
	fields := strings.Fields(strings.ReplaceAll(value, "\n", " "))
	return strings.Join(fields, " ")
}

// MetricsHandler menampilkan metrik dasar dalam format teks sederhana.
func MetricsHandler(cfg *config.Config, worker *builder.Worker) gin.HandlerFunc {
	return func(c *gin.Context) {
		requests, errors, lastReq := Metrics.snapshot()
		stats := builder.WorkerStats{}
		if worker != nil {
			stats = worker.Stats()
		}
		var sitesFree, cacheFree uint64
		if cfg != nil {
			sitesFree, _ = diskFreeBytes(cfg.Build.SitesDir)
			cacheFree, _ = diskFreeBytes(cfg.Build.CacheDir)
		}

		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(
			"only4fun_metrics\n"+
				"requests_total %d\n"+
				"errors_total %d\n"+
				"last_request_time %d\n"+
				"build_queued %d\n"+
				"build_active %d\n"+
				"build_total %d\n"+
				"build_succeeded %d\n"+
				"build_failed %d\n"+
				"build_last_duration_ms %d\n"+
				"build_last_failure %s\n"+
				"disk_free_sites_bytes %d\n"+
				"disk_free_cache_bytes %d\n",
			requests,
			errors,
			lastReq.Unix(),
			stats.Queued,
			stats.Active,
			stats.Total,
			stats.Succeeded,
			stats.Failed,
			stats.LastMS,
			metricText(stats.LastFailure),
			sitesFree,
			cacheFree,
		))
	}
}
