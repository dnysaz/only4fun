package handler

import (
	"net/http"
	"strings"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memverifikasi session cookie dan memastikan token GitHub
// milik session masih valid. Pengecekan token memakai cache singkat supaya
// tidak memanggil GitHub pada setiap request API.
func AuthMiddleware(sessionStore *database.SessionStore) gin.HandlerFunc {
	return AuthMiddlewareWithGitHubAPI(sessionStore, defaultGitHubAPIBase)
}

// AuthMiddlewareWithGitHubAPI dipakai router dan test (API base diganti server
// lokal saat test).
func AuthMiddlewareWithGitHubAPI(sessionStore *database.SessionStore, apiBase string) gin.HandlerFunc {
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

		if !githubTokenValid(session.Token, apiBase) {
			_ = sessionStore.DeleteByID(session.ID)
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(sessionCookieName, "", -1, "/", "", isHTTPS(c), true)
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		c.Set("userID", session.UserID)
		c.Next()
	}
}
