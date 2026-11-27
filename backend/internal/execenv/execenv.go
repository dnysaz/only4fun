// Package execenv builds a minimal environment for processes that run
// user-controlled source (git hooks, npm install/build scripts, SSR servers).
//
// Never use os.Environ() here: the server process holds secrets such as
// SESSION_SECRET, GitHub OAuth credentials and the Cloudflare API token.
package execenv

import (
	"os"
	"path/filepath"
	"strings"
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

// reservedKeys must never be supplied by project owners. They either replace
// the sandbox PATH/HOME, preload native code into build tools, or inject Node
// flags into every build script.
var reservedKeys = map[string]struct{}{
	"PATH": {}, "HOME": {}, "TMPDIR": {}, "TEMP": {}, "TMP": {}, "SHELL": {},
	"BASH_ENV": {}, "ENV": {}, "CDPATH": {}, "IFS": {},
	"LD_PRELOAD": {}, "LD_LIBRARY_PATH": {}, "DYLD_INSERT_LIBRARIES": {},
	"DYLD_LIBRARY_PATH": {}, "DYLD_FRAMEWORK_PATH": {},
	"NODE_OPTIONS": {}, "NODE_PATH": {}, "PORT": {},
	// SSR wajib tetap di loopback; user tidak boleh mengganti alamat bind.
	"HOST": {}, "BIND": {}, "LISTEN_ADDRESS": {},
	"NPM_CONFIG_CACHE": {}, "NPM_CONFIG_PREFIX": {}, "NPM_CONFIG_USERCONFIG": {},
}

// IsReservedKey reports whether a key may be overridden by user-provided env.
func IsReservedKey(key string) bool {
	upper := strings.ToUpper(strings.TrimSpace(key))
	_, reserved := reservedKeys[upper]
	return reserved
}

// WithUserEnv merges validated project env vars into the isolated base env.
// Internal values are appended last: os/exec keeps the last duplicate, so a
// legacy value stored in the database still cannot replace PATH/HOME/CI.
func WithUserEnv(base []string, user map[string]string) []string {
	env := make([]string, 0, len(user)+len(base))
	for key, value := range user {
		if key == "" || IsReservedKey(key) {
			continue
		}
		env = append(env, key+"="+value)
	}
	return append(env, base...)
}

// EnsureHome creates the HOME directory used by Base so npm can write its cache.
func EnsureHome(home string) error {
	if home == "" {
		home = filepath.Join(os.TempDir(), "only4.fun-home")
	}
	return os.MkdirAll(filepath.Join(home, ".npm"), 0o700)
}
