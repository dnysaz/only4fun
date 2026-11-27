package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// checkWritableDir memastikan folder siap dipakai build. CreateTemp dipakai
// supaya pengecekan juga menangkas izin tulis yang berubah saat runtime.
func checkWritableDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".ready-*")
	if err != nil {
		return err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Remove(name)
}

// ReadinessHandler berbeda dari /health: /health hanya menyatakan proses
// hidup, sedangkan /ready dipakai orchestrator sebelum mengalihkan traffic.
func ReadinessHandler(cfg *config.Config, db *database.Database, worker *builder.Worker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil || db == nil || worker == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "missing runtime dependencies"})
			return
		}
		if err := db.DB().Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "database unavailable"})
			return
		}
		for _, dir := range []string{filepath.Dir(cfg.Database.Path), cfg.Build.SitesDir, cfg.Build.CacheDir} {
			if err := checkWritableDir(dir); err != nil {
				log.Printf("readiness check failed for %s: %v", dir, err)
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "storage unavailable"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
