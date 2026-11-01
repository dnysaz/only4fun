package handler

import "testing"

// TestSafeNextPath memastikan ?next= hanya boleh mengarah ke path lokal,
// sehingga tidak bisa dipakai untuk open redirect setelah login.
func TestSafeNextPath(t *testing.T) {
	valid := map[string]string{
		"/dashboard":        "/dashboard",
		"/projects/abc-123": "/projects/abc-123",
		"/docs?tab=api":     "/docs?tab=api",
	}
	for input, want := range valid {
		if got := safeNextPath(input); got != want {
			t.Errorf("safeNextPath(%q) = %q, ingin %q", input, got, want)
		}
	}

	invalid := []string{
		"",
		"dashboard",
		"//evil.com",
		"https://evil.com",
		"/\\evil.com",
		"/ok\r\nSet-Cookie: x=1",
	}
	for _, input := range invalid {
		if got := safeNextPath(input); got != "" {
			t.Errorf("safeNextPath(%q) = %q, ingin kosong", input, got)
		}
	}
}
