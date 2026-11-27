package server

import (
	"log"
	"time"

	"only4.fun/backend/internal/proxytrust"

	"github.com/gin-gonic/gin"
)

// TrustedProxyMiddleware menandai apakah peer langsung berasal dari proxy yang
// dipercaya. Handler memakai penanda ini sebelum mempercayai header forwarding.
func TrustedProxyMiddleware(trust *proxytrust.Trust) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("trustedProxy", trust.ContainsRemote(c.Request.RemoteAddr))
		c.Next()
	}
}

// RequestLogger menambah logging utama untuk setiap request: method, path, status, durasi.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		Metrics.record(status >= 400, start)

		log.Printf("REQ %s %s %d %dms", method, path, status, duration)
	}
}
