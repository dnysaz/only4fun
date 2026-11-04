package server

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu        sync.Mutex
	hits      map[string][]time.Time
	lastSweep time.Time
	limit     int
	window    time.Duration
}

// RateLimit membatasi jumlah request per user/session (fallback: IP) dalam satu window.
// Dipakai untuk endpoint yang memicu kerja berat seperti deploy dan pembuatan project.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := &rateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}

	return func(c *gin.Context) {
		key := rateLimitKey(c)

		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)

		// Bersihkan entri kedaluwarsa minimal sekali per window supaya map tidak
		// tumbuh tanpa batas untuk setiap IP/session yang pernah request.
		if now.Sub(rl.lastSweep) >= rl.window {
			for k, times := range rl.hits {
				kept := times[:0]
				for _, t := range times {
					if t.After(cutoff) {
						kept = append(kept, t)
					}
				}
				if len(kept) == 0 {
					delete(rl.hits, k)
				} else {
					rl.hits[k] = kept
				}
			}
			rl.lastSweep = now
		}

		recent := rl.hits[key][:0]
		for _, t := range rl.hits[key] {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) >= rl.limit {
			rl.hits[key] = recent
			rl.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, please try again later"})
			return
		}
		rl.hits[key] = append(recent, now)
		rl.mu.Unlock()

		c.Next()
	}
}

func rateLimitKey(c *gin.Context) string {
	// Prioritas: session cookie > user ID (auth) > IP.
	// Ini mencegah pengguna yang login berbagi kuota dengan pengguna lain
	// di IP yang sama (mis. kantor, Wi-Fi publik).
	if cookie, err := c.Cookie("session_id"); err == nil && cookie != "" {
		return "session:" + cookie
	}
	if id, ok := c.Get("userID"); ok {
		if userID, ok := id.(string); ok && userID != "" {
			return "user:" + userID
		}
	}
	return "ip:" + c.ClientIP()
}
