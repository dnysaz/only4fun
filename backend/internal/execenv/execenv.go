// Package execenv builds a minimal environment for processes that run
// user-controlled source (git hooks, npm install/build scripts, SSR servers).
//
// Never use os.Environ() here: the server process holds secrets such as
// SESSION_SECRET, GitHub OAuth credentials and the Cloudflare API token.
package execenv

import (
	"os"
	"path/filepath"
)

// Base returns the only variables that untrusted child processes need.
// Callers can append explicit, validated overrides after this slice.
func Base(home string) []string {
	if home == "" {
		home = filepath.Join(os.TempDir(), "only4.fun-home")
	}

	return []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + home,
		"TMPDIR=" + os.TempDir(),
		"NPM_CONFIG_CACHE=" + filepath.Join(home, ".npm"),
		"npm_config_update_notifier=false",
		"CI=1",
	}
}

// EnsureHome creates the HOME directory used by Base so npm can write its cache.
func EnsureHome(home string) error {
	if home == "" {
		home = filepath.Join(os.TempDir(), "only4.fun-home")
	}
	return os.MkdirAll(filepath.Join(home, ".npm"), 0o700)
}
