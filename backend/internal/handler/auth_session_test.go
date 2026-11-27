package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// callGetMe menjalankan GetMe dengan cookie sesi seperti browser.
func callGetMe(t *testing.T, h *AuthHandler, sessionID, userID string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if sessionID != "" {
		c.Request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	}
	c.Set("userID", userID)
	h.GetMe(c)
	return rec
}

func TestGetMeRevokedGitHubTokenEndsSession(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer revoked-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	store := newTestStore(t)
	sessions := database.NewSessionStore(store.db)
	userID := store.createUser(t, "alice")
	if _, err := sessions.Create(userID, "revoked-session", "revoked-token", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create session: %v", err)
	}

	handler, err := NewAuthHandler("id", "secret", "http://localhost/callback", "http://localhost", store.users, sessions)
	if err != nil {
		t.Fatalf("NewAuthHandler: %v", err)
	}
	handler.githubAPIBase = upstream.URL

	rec := callGetMe(t, handler, "revoked-session", userID)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetMe = %d, ingin 401: %s", rec.Code, rec.Body.String())
	}
	if _, err := sessions.GetBySessionID("revoked-session"); err == nil {
		t.Fatal("session dengan token dicabut masih tersimpan")
	}
}

func TestGetMeValidGitHubTokenAndDevSession(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	store := newTestStore(t)
	sessions := database.NewSessionStore(store.db)
	userID := store.createUser(t, "alice")
	if _, err := sessions.Create(userID, "good-session", "good-token", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := sessions.Create(userID, "dev-session", "", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create dev session: %v", err)
	}

	handler, err := NewAuthHandler("id", "secret", "http://localhost/callback", "http://localhost", store.users, sessions)
	if err != nil {
		t.Fatalf("NewAuthHandler: %v", err)
	}
	handler.githubAPIBase = upstream.URL

	for _, sessionID := range []string{"good-session", "dev-session"} {
		rec := callGetMe(t, handler, sessionID, userID)
		if rec.Code != http.StatusOK {
			t.Fatalf("GetMe(%q) = %d, ingin 200: %s", sessionID, rec.Code, rec.Body.String())
		}
	}
}
