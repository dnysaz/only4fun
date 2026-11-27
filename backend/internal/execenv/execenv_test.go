package execenv

import (
	"strings"
	"testing"
)

func TestWithUserEnvCannotOverrideInternalValues(t *testing.T) {
	t.Setenv("PATH", "/server/bin")
	base := Base("/tmp/only4-fun-home")
	env := strings.Join(WithUserEnv(base, map[string]string{
		"PATH":           "/tmp/evil",
		"HOME":           "/tmp/evil",
		"LD_PRELOAD":     "/tmp/evil.so",
		"NODE_OPTIONS":   "--require /tmp/evil.js",
		"HOST":           "0.0.0.0",
		"HOSTNAME":       "0.0.0.0",
		"BIND":           "0.0.0.0",
		"LISTEN_ADDRESS": "0.0.0.0",
		"NODE_ENV":       "development",
		"SUPABASE_URL":   "https://example.supabase.co",
	}), "\n")

	if !strings.Contains(env, "PATH=/server/bin") {
		t.Errorf("user PATH replaced the sandbox PATH: %s", env)
	}
	if !strings.Contains(env, "HOME=/tmp/only4-fun-home") {
		t.Errorf("user HOME replaced the sandbox HOME: %s", env)
	}
	if !strings.Contains(env, "SUPABASE_URL=https://example.supabase.co") {
		t.Errorf("valid user env var was dropped: %s", env)
	}
	for _, needle := range []string{"/tmp/evil", "LD_PRELOAD=", "NODE_OPTIONS=", "0.0.0.0"} {
		if strings.Contains(env, needle) {
			t.Errorf("WithUserEnv leaked %q", needle)
		}
	}
}

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
