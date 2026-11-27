package execenv

import (
	"strings"
	"testing"
)

func TestBaseDoesNotLeakServerSecrets(t *testing.T) {
	t.Setenv("SESSION_SECRET", "test-session-secret-value")
	t.Setenv("GITHUB_CLIENT_SECRET", "test-github-secret-value")
	t.Setenv("CLOUDFLARE_API_KEY", "test-cloudflare-token-value")
	t.Setenv("DATABASE_PATH", "/secret/only4.fun.db")

	env := strings.Join(Base("/tmp/only4-fun-home"), "\n")

	for _, needle := range []string{
		"SESSION_SECRET=",
		"GITHUB_CLIENT_SECRET=",
		"CLOUDFLARE_API_KEY=",
		"DATABASE_PATH=",
		"test-session-secret-value",
		"test-github-secret-value",
		"test-cloudflare-token-value",
	} {
		if strings.Contains(env, needle) {
			t.Errorf("Base() leaked %q", needle)
		}
	}

	for _, key := range []string{"PATH=", "HOME=", "TMPDIR=", "NPM_CONFIG_CACHE=", "CI=1"} {
		if !strings.Contains(env, key) {
			t.Errorf("Base() is missing %q", key)
		}
	}
}
