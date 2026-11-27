package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestUnverifiedWebhookDoesNotCreateSecret(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)
	userID := store.createUser(t, "alice")
	projectID := store.createProject(t, userID, "https://github.com/alice/site", "site-webhook", nil)
	// Simulasikan project lama yang belum punya secret webhook.
	if _, err := store.db.DB().Exec(`UPDATE projects SET webhook_secret = '' WHERE id = ?`, projectID); err != nil {
		t.Fatal(err)
	}

	body := `{"ref":"refs/heads/main","repository":{"html_url":"https://github.com/alice/site"}}`
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	handler.HandleGitHubWebhook(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid signature status = %d, want 401", rec.Code)
	}
	var stored string
	if err := store.db.DB().Get(&stored, `SELECT webhook_secret FROM projects WHERE id = ?`, projectID); err != nil {
		t.Fatal(err)
	}
	if stored != "" {
		t.Fatal("unverified webhook created a webhook secret")
	}
}

func TestValidWebhookSignature(t *testing.T) {
	body := []byte(`{"ref":"refs/heads/main"}`)
	valid := sign("s3cret", body)

	cases := []struct {
		name     string
		secret   string
		header   string
		expected bool
	}{
		{"signature cocok", "s3cret", valid, true},
		{"secret salah", "lain", valid, false},
		{"body diubah", "s3cret", sign("s3cret", []byte("{}")), false},
		{"tanpa prefix", "s3cret", hex.EncodeToString([]byte("abc")), false},
		{"header kosong", "s3cret", "", false},
		{"secret kosong", "", valid, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validWebhookSignature(tc.secret, body, tc.header); got != tc.expected {
				t.Fatalf("validWebhookSignature = %v, ingin %v", got, tc.expected)
			}
		})
	}
}
