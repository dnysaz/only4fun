package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/secure"

	"github.com/gin-gonic/gin"
)

func TestAuditMiddlewareRecordsSensitiveRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := secure.Init("only4.fun-audit-test-secret"); err != nil {
		t.Fatal(err)
	}
	db, err := database.Init(filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	users := database.NewUserStore(db)
	user, err := users.Create("gh-audit-server", "audit-server-user", "", "")
	if err != nil {
		t.Fatal(err)
	}
	audit := database.NewAuditStore(db)

	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	r.Use(func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, AuditMiddleware(audit))
	r.DELETE("/api/projects/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/project-1", nil)
	req.RemoteAddr = "203.0.113.9:4567"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}

	logs, err := audit.ListByUser(user.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(logs))
	}
	if logs[0].Action != "project.delete" || logs[0].Target != "project-1" {
		t.Fatalf("unexpected entry: %+v", logs[0])
	}
	if logs[0].IP != "203.0.113.9" {
		t.Fatalf("IP = %q, want 203.0.113.9", logs[0].IP)
	}
}

func TestAuditActionUsesFullPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		method, pattern, path, action string
	}{
		{"POST", "/api/projects/upload", "/api/projects/upload", "project.upload"},
		{"PUT", "/api/projects/:id/domain", "/api/projects/p1/domain", "project.domain.update"},
		{"DELETE", "/api/projects/:id/deployments", "/api/projects/p1/deployments", "deployment.history.clear"},
		{"POST", "/api/projects/:id/deployments/:deploymentID/rollback", "/api/projects/p1/deployments/d1/rollback", "deployment.rollback"},
		{"POST", "/api/auth/logout", "/api/auth/logout?all=true", "auth.logout_all"},
	}

	for _, tc := range cases {
		var got string
		r := gin.New()
		r.Handle(tc.method, tc.pattern, func(c *gin.Context) {
			got, _ = auditAction(c)
			c.Status(http.StatusOK)
		})
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(tc.method, tc.path, nil))
		if got != tc.action {
			t.Errorf("%s %s action = %q, want %q", tc.method, tc.path, got, tc.action)
		}
	}
}
