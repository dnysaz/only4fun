package handler

import (
	"net/http"
	"strings"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memverifikasi session cookie.
func AuthMiddleware(sessionStore *database.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_id")
		if err != nil || cookie == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				cookie = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if cookie == "" {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		session, err := sessionStore.GetBySessionID(cookie)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		c.Set("userID", session.UserID)
		c.Next()
	}
}
