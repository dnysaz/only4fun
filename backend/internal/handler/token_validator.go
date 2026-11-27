package handler

import (
	"crypto/sha256"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// githubTokenCheckTTL membatasi berapa lama hasil pengecekan token dipakai
// ulang. Middleware dipanggil setiap request API, jadi GitHub tidak boleh
// dihubungi tiap kali; 60 detik cukup singkat untuk mencabut sesi yang tokennya
// revoked tanpa memperlambat semua request.
const githubTokenCheckTTL = 60 * time.Second

type githubTokenCheck struct {
	valid     bool
	checkedAt time.Time
}

var (
	githubTokenChecks     sync.Map
	githubTokenHTTPClient = &http.Client{Timeout: 8 * time.Second}
)

// githubTokenValid mengembalikan false hanya bila GitHub menolak token dengan
// 401. 403 diperlakukan sebagai valid karena GitHub memakainya untuk secondary
// rate limit; gangguan jaringan juga tidak boleh mengeluarkan user.
func githubTokenValid(token, apiBase string) bool {
	// Sesi demo/dev sengaja dibuat tanpa token GitHub.
	if token == "" {
		return true
	}

	// Simpan hash, bukan token mentah, sebagai kunci cache.
	key := sha256.Sum256([]byte(token))
	if cached, ok := githubTokenChecks.Load(key); ok {
		check := cached.(githubTokenCheck)
		if time.Since(check.checkedAt) < githubTokenCheckTTL {
			return check.valid
		}
	}

	valid := checkGitHubToken(token, apiBase)
	githubTokenChecks.Store(key, githubTokenCheck{valid: valid, checkedAt: time.Now()})
	return valid
}

func checkGitHubToken(token, apiBase string) bool {
	base := strings.TrimRight(apiBase, "/")
	if base == "" {
		base = defaultGitHubAPIBase
	}

	req, err := http.NewRequest(http.MethodGet, base+"/user", nil)
	if err != nil {
		return true
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := githubTokenHTTPClient.Do(req)
	if err != nil {
		// GitHub sedang tidak bisa dijangkau: jangan log out user karena gangguan jaringan.
		return true
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.StatusCode != http.StatusUnauthorized
}
