package server

import (
	"log"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// auditAction memetakan route dengan efek samping ke nama aksi dan target.
// Target sengaja hanya id, bukan nilai env/token, supaya log audit aman dibaca.
func auditAction(c *gin.Context) (action, target string) {
	target = c.Param("id")
	if deploymentID := c.Param("deploymentID"); deploymentID != "" {
		target = target + "/" + deploymentID
	}

	switch c.FullPath() {
	case "/api/projects":
		if c.Request.Method == "POST" {
			return "project.create", target
		}
	case "/api/projects/upload":
		if c.Request.Method == "POST" {
			return "project.upload", target
		}
	case "/api/projects/:id":
		if c.Request.Method == "DELETE" {
			return "project.delete", target
		}
	case "/api/projects/:id/env":
		switch c.Request.Method {
		case "POST", "PUT", "PATCH":
			return "project.env.update", target
		}
	case "/api/projects/:id/deploy":
		if c.Request.Method == "POST" {
			return "project.deploy", target
		}
	case "/api/projects/:id/domain":
		if c.Request.Method == "PUT" {
			return "project.domain.update", target
		}
	case "/api/projects/:id/deployments":
		if c.Request.Method == "DELETE" {
			return "deployment.history.clear", target
		}
	case "/api/projects/:id/deployments/:deploymentID/rollback":
		if c.Request.Method == "POST" {
			return "deployment.rollback", target
		}
	case "/api/auth/logout":
		if c.Request.Method == "POST" {
			if c.Query("all") == "true" {
				return "auth.logout_all", "all-sessions"
			}
			return "auth.logout", "current-session"
		}
	}
	return "", ""
}

// AuditMiddleware mencatat aksi sensitif setelah handler selesai. Request yang
// gagal tetap dicatat (beserta statusnya) karena percobaan itu penting untuk
// investigasi. Kegagalan DB hanya masuk log; respons ke user tidak berubah.
func AuditMiddleware(store *database.AuditStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if store == nil {
			return
		}
		action, target := auditAction(c)
		if action == "" {
			return
		}
		userID := c.GetString("userID")
		if err := store.Record(userID, action, target, c.ClientIP(), c.Writer.Status()); err != nil {
			log.Printf("failed to write audit log %s: %v", action, err)
		}
	}
}
