package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics menyimpan metrik dasar untuk endpoint /metrics.
var Metrics = struct {
	Requests int
	Errors   int
	LastReq  time.Time
}{
	Requests: 0,
	Errors:   0,
	LastReq:  time.Time{},
}

// MetricsHandler menampilkan metrik dasar dalam format teks sederhana.
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(
			"only4fun_metrics\nrequests_total %d\nerrors_total %d\nlast_request_time %d\n",
			Metrics.Requests,
			Metrics.Errors,
			Metrics.LastReq.Unix(),
		))
	}
}
