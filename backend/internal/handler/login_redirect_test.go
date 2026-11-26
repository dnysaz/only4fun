package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestAuthLoginRouter(t *testing.T, store *testStore) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler, err := NewAuthHandler(
		"client-id",
		"client-secret",
		"http://localhost:8080/auth/github/callback",
		"http://localhost:3000",
		store.users,
		nil,
	)
	if err != nil {
		t.Fatalf("NewAuthHandler: %v", err)
	}

	r := gin.New()
	r.GET("/auth/github/login", handler.GetGitHubLoginURL)
	return r
}

// nextCookieValue mengembalikan nilai cookie tujuan seperti yang dibaca handler
// (gin menulis nilai cookie dalam bentuk URL-encoded).
func nextCookieValue(res *http.Response) string {
	for _, cookie := range res.Cookies() {
		if cookie.Name == oauthNextCookie {
			value, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				return cookie.Value
			}
			return value
		}
	}
	return ""
}

// TestLoginStoresNextPath memastikan halaman tujuan tersimpan agar pengguna
// kembali ke halaman yang tadi terkunci, bukan selalu ke dashboard.
func TestLoginStoresNextPath(t *testing.T) {
	store := newTestStore(t)
	r := newTestAuthLoginRouter(t, store)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/login?next=%2Fprojects%2Fabc%3Ftab%3Dlogs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("login = %d, ingin 302 ke GitHub", rec.Code)
	}
	if got := nextCookieValue(rec.Result()); got != "/projects/abc?tab=logs" {
		t.Errorf("cookie %s = %q, ingin %q", oauthNextCookie, got, "/projects/abc?tab=logs")
	}
	if rec.Result().Cookies()[0].Name != oauthStateCookie {
		t.Logf("cookie pertama: %s", rec.Result().Cookies()[0].Name)
	}
}

// TestLoginRejectsExternalNext memastikan cookie tujuan tidak diisi untuk
// URL luar supaya tidak ada open redirect.
func TestLoginRejectsExternalNext(t *testing.T) {
	store := newTestStore(t)
	r := newTestAuthLoginRouter(t, store)

	for _, next := range []string{"https://evil.com", "//evil.com", "evil.com"} {
		req := httptest.NewRequest(http.MethodGet, "/auth/github/login?next="+next, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("login next=%q = %d, ingin 302", next, rec.Code)
		}
		if got := nextCookieValue(rec.Result()); got != "" {
			t.Errorf("next=%q menyimpan cookie %q, ingin kosong", next, got)
		}
	}
}

// TestLoginWithoutNext memastikan tidak ada cookie tujuan saat next tidak diisi.
func TestLoginWithoutNext(t *testing.T) {
	store := newTestStore(t)
	r := newTestAuthLoginRouter(t, store)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := nextCookieValue(rec.Result()); got != "" {
		t.Errorf("cookie oauth_next = %q, ingin kosong", got)
	}
}
