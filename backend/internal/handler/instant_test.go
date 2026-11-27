package handler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRandomInstantSubdomain(t *testing.T) {
	for i := 0; i < 100; i++ {
		sub, err := randomInstantSubdomain()
		if err != nil {
			t.Fatalf("randomInstantSubdomain: %v", err)
		}
		if len(sub) != instantSubdomainLength {
			t.Fatalf("got %q with length %d, want %d", sub, len(sub), instantSubdomainLength)
		}
		for _, c := range sub {
			if !strings.ContainsRune(instantAlphabet, c) {
				t.Fatalf("got invalid character %q in %q", c, sub)
			}
		}
	}
}

func TestCheckDirSize(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("1234567890"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkDirSize(dir, 5); err == nil {
		t.Fatal("oversized directory was accepted")
	}
	if err := checkDirSize(dir, 100); err != nil {
		t.Fatalf("small directory rejected: %v", err)
	}
}

func resetInstantReservationsForTest() {
	instantReservationMu.Lock()
	defer instantReservationMu.Unlock()
	instantReservations = make(map[string]instantAddressReservation)
}

func TestInstantAddressReservationIsUniqueAndClaimableOnce(t *testing.T) {
	resetInstantReservationsForTest()

	const subdomain = "abc123xy"
	token, ok := tryReserveInstantAddress(subdomain)
	if !ok || token == "" {
		t.Fatal("first reservation should succeed")
	}
	if _, ok := tryReserveInstantAddress(subdomain); ok {
		t.Fatal("the same address must not be reserved twice")
	}
	if !isInstantAddressReserved(subdomain) {
		t.Fatal("reserved address should be reported as reserved")
	}
	if claimInstantAddress(subdomain, "wrong-token") {
		t.Fatal("a different token must not claim the address")
	}
	if !claimInstantAddress(subdomain, token) {
		t.Fatal("the matching token should claim the address")
	}
	if isInstantAddressReserved(subdomain) {
		t.Fatal("claimed address should no longer be reserved")
	}
	if claimInstantAddress(subdomain, token) {
		t.Fatal("the same reservation must not be claimed twice")
	}
}

func TestInstantAddressReservationExpires(t *testing.T) {
	resetInstantReservationsForTest()

	const subdomain = "expired1"
	token, ok := tryReserveInstantAddress(subdomain)
	if !ok {
		t.Fatal("reservation should succeed")
	}

	instantReservationMu.Lock()
	reservation := instantReservations[subdomain]
	reservation.expires = time.Now().Add(-time.Second)
	instantReservations[subdomain] = reservation
	instantReservationMu.Unlock()

	if isInstantAddressReserved(subdomain) {
		t.Fatal("expired reservation should be ignored")
	}
	if claimInstantAddress(subdomain, token) {
		t.Fatal("expired reservation must not be claimable")
	}
	if _, ok := tryReserveInstantAddress(subdomain); !ok {
		t.Fatal("expired address should be reservable again")
	}
}

func TestNewAvailableAddressNeverRepeats(t *testing.T) {
	resetInstantReservationsForTest()

	h := &InstantHandler{domain: "only4.fun"}
	seen := make(map[string]bool)
	for i := 0; i < 24; i++ {
		subdomain, token, err := h.newAvailableAddress()
		if err != nil {
			t.Fatalf("newAvailableAddress: %v", err)
		}
		if seen[subdomain] {
			t.Fatalf("duplicate preview address %q", subdomain)
		}
		seen[subdomain] = true
		if !claimInstantAddress(subdomain, token) {
			t.Fatalf("could not claim preview address %q", subdomain)
		}
	}
}

func TestValidateStaticSite(t *testing.T) {
	write := func(dir, rel, body string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("plain static site", func(t *testing.T) {
		dir := t.TempDir()
		write(dir, "index.html", "<h1>hi</h1>")
		write(dir, "style.css", "body{}")
		write(dir, "app.js", "console.log(1)")
		if err := validateStaticSite(dir); err != nil {
			t.Fatalf("want static site accepted, got %v", err)
		}
	})

	t.Run("missing index", func(t *testing.T) {
		dir := t.TempDir()
		write(dir, "readme.txt", "hello")
		err := validateStaticSite(dir)
		if err == nil || errors.Is(err, errFrameworkProject) {
			t.Fatalf("want a missing-index error, got %v", err)
		}
	})

	cases := map[string]string{
		"package.json":                   `{"name":"app"}`,
		"server.js":                      `require("http")`,
		"src/App.jsx":                    `export default () => null`,
		"dist/server.js":                 `require("http")`,
		"_next/static/chunk.js":          `// built output`,
		"assets/index-a1b2c3d4.js":       `// hashed bundle`,
		"static/js/main.1a2b3c.chunk.js": `// chunk`,
	}
	for name, body := range cases {
		t.Run("reject "+name, func(t *testing.T) {
			dir := t.TempDir()
			write(dir, "index.html", "<h1>hi</h1>")
			write(dir, name, body)
			err := validateStaticSite(dir)
			if !errors.Is(err, errFrameworkProject) {
				t.Fatalf("want framework error for %s, got %v", name, err)
			}
		})
	}
}
