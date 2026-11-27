package server

import (
	"encoding/json"
	"log"
	"regexp"
	"time"

	"only4.fun/backend/internal/id"
	"only4.fun/backend/internal/proxytrust"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

// RequestID menerima ID dari proxy/front-end bila formatnya aman, kalau tidak
// membuat UUID baru. ID ikut dikembalikan di response agar mudah dilacak.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if !requestIDPattern.MatchString(requestID) {
			requestID = id.New()
		}
		c.Set("requestID", requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

// TrustedProxyMiddleware menandai apakah peer langsung berasal dari proxy yang
// dipercaya. Handler memakai penanda ini sebelum mempercayai header forwarding.
func TrustedProxyMiddleware(trust *proxytrust.Trust) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("trustedProxy", trust.ContainsRemote(c.Request.RemoteAddr))
		c.Next()
	}
}

// RequestLogger menulis satu baris JSON per request. Format terstruktur lebih
// mudah dicari di log aggregator dan menyertakan request ID untuk korelasi.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestPath := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		Metrics.record(status >= 400, start)

		entry := map[string]any{
			"time":        start.Format(time.RFC3339Nano),
			"request_id":  c.GetString("requestID"),
			"method":      method,
			"path":        requestPath,
			"status":      status,
			"duration_ms": time.Since(start).Milliseconds(),
			"ip":          c.ClientIP(),
		}
		if errs := c.Errors.String(); errs != "" {
			entry["errors"] = errs
		}
		encoded, err := json.Marshal(entry)
		if err != nil {
			log.Printf("request id=%s %s %s %d", c.GetString("requestID"), method, requestPath, status)
			return
		}
		log.Printf("%s", encoded)
	}
}
