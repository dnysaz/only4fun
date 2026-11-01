package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
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
