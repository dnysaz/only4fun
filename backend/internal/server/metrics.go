package server

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

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

// MetricsHandler menampilkan metrik dasar dalam format teks sederhana.
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requests, errors, lastReq := Metrics.snapshot()
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(
			"only4fun_metrics\nrequests_total %d\nerrors_total %d\nlast_request_time %d\n",
			requests,
			errors,
			lastReq.Unix(),
		))
	}
}
