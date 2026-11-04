package server

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger menambah logging utama untuk setiap request: method, path, status, durasi.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		if status >= 400 {
			Metrics.Errors++
		}
		Metrics.Requests++
		Metrics.LastReq = start

		log.Printf("REQ %s %s %d %dms", method, path, status, duration)
	}
}
