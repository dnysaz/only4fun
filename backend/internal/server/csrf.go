package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// TrustedOriginGuard melindungi endpoint yang menerima form/multipart.
//
// Request lintas origin dengan Content-Type application/json sudah diblokir
// oleh preflight CORS, tetapi multipart/form-data adalah "simple request" yang
// dikirim browser tanpa preflight. middleware ini menolak Origin yang tidak
// ada dalam daftar origin frontend yang diizinkan. Request non-browser (tanpa
// header Origin, mis. curl) tetap diterima.
func TrustedOriginGuard(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, raw := range allowedOrigins {
		if origin := normalizeOrigin(raw); origin != "" {
			allowed[origin] = true
		}
	}

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}

		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cross-origin requests are not allowed"})
			return
		}
		normalized := strings.ToLower(u.Scheme + "://" + u.Host)
		// Origin aplikasi itu sendiri juga valid bila request datang langsung
		// dari halaman backend (mis. form di domain yang sama).
		if allowed[normalized] || strings.EqualFold(u.Host, c.Request.Host) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cross-origin requests are not allowed"})
	}
}

func normalizeOrigin(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Scheme + "://" + u.Host)
}
