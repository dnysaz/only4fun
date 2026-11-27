package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func callAuthMiddleware(t *testing.T, apiBase, sessionID string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store := newTestStore(t)
	sessions := database.NewSessionStore(store.db)
	userID := store.createUser(t, "alice")
	if _, err := sessions.Create(userID, sessionID, strings.TrimSuffix(sessionID, "-session"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create session: %v", err)
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	c.Request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	AuthMiddlewareWithGitHubAPI(sessions, apiBase)(c)
	return rec
}

func TestAuthMiddlewareRevokedTokenEndsSession(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	rec := callAuthMiddleware(t, upstream.URL, "middleware-revoked-session")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("middleware = %d, ingin 401: %s", rec.Code, rec.Body.String())
	}
}

func TestIsHTTPSRequiresTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	check := func(trusted bool) bool {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")
		c.Set("trustedProxy", trusted)
		return isHTTPS(c)
	}

	if check(false) {
		t.Error("forwarded proto from an untrusted peer must be ignored")
	}
	if !check(true) {
		t.Error("forwarded proto from a trusted proxy must be honored")
	}
}

func TestAuthMiddlewareAllowsRateLimitedToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer upstream.Close()

	rec := callAuthMiddleware(t, upstream.URL, "middleware-ratelimited-session")
	if rec.Code != http.StatusOK {
		t.Fatalf("middleware 403 = %d, ingin 200: %s", rec.Code, rec.Body.String())
	}
}

func TestLogoutAllSessions(t *testing.T) {
	store := newTestStore(t)
	sessions := database.NewSessionStore(store.db)
	userID := store.createUser(t, "alice")
	otherUserID := store.createUser(t, "bob")
	for _, sessionID := range []string{"alice-one", "alice-two"} {
		if _, err := sessions.Create(userID, sessionID, "", time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := sessions.Create(otherUserID, "bob-one", "", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	handler, err := NewAuthHandler("id", "secret", "http://localhost/callback", "http://localhost", store.users, sessions)
	if err != nil {
		t.Fatal(err)
	}

	call := func(sessionID, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
		handler.Logout(c)
		return rec
	}

	if rec := call("alice-one", `{"all":true}`); rec.Code != http.StatusOK {
		t.Fatalf("logout all = %d: %s", rec.Code, rec.Body.String())
	}
	for _, sessionID := range []string{"alice-one", "alice-two"} {
		if _, err := sessions.GetBySessionID(sessionID); err == nil {
			t.Fatalf("session %q should be deleted by logout all", sessionID)
		}
	}
	if _, err := sessions.GetBySessionID("bob-one"); err != nil {
		t.Fatalf("another user's session was deleted: %v", err)
	}
}

func TestLogoutOnlyCurrentSessionByDefault(t *testing.T) {
	store := newTestStore(t)
	sessions := database.NewSessionStore(store.db)
	userID := store.createUser(t, "alice")
	for _, sessionID := range []string{"current", "other"} {
		if _, err := sessions.Create(userID, sessionID, "", time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	handler, err := NewAuthHandler("id", "secret", "http://localhost/callback", "http://localhost", store.users, sessions)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	c.Request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "current"})
	handler.Logout(c)

	if _, err := sessions.GetBySessionID("current"); err == nil {
		t.Fatal("current session should be deleted")
	}
	if _, err := sessions.GetBySessionID("other"); err != nil {
		t.Fatalf("other session should remain: %v", err)
	}
}
