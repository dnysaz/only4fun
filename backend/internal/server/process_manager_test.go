package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestSSREnvForcesLoopback(t *testing.T) {
	env := ssrEnv(4321)
	for _, want := range []string{
		"PORT=4321",
		"NODE_ENV=production",
		"HOST=127.0.0.1",
		"HOSTNAME=127.0.0.1",
		"BIND=127.0.0.1",
		"LISTEN_ADDRESS=127.0.0.1",
	} {
		if !slices.Contains(env, want) {
			t.Fatalf("ssrEnv missing %q: %v", want, env)
		}
	}
	for _, leaked := range []string{"0.0.0.0", "::"} {
		if slices.Contains(env, "HOST="+leaked) || slices.Contains(env, "BIND="+leaked) {
			t.Fatalf("ssrEnv exposes public bind address: %v", env)
		}
	}
}

func TestStartPassesProjectEnvToServer(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir := t.TempDir()
	out := filepath.Join(dir, "env.txt")
	script := filepath.Join(dir, "server.js")
	body := fmt.Sprintf(`
const fs = require('fs');
fs.writeFileSync(%q, (process.env.MY_VAR || '') + '|' + (process.env.PORT || ''));
const http = require('http');
http.createServer((req, res) => res.end('ok')).listen(Number(process.env.PORT), '127.0.0.1');
`, out)
	if err := os.WriteFile(script, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	pm := NewProcessManager()
	defer pm.Stop("env-site")
	port, err := pm.Start("env-site", script, map[string]string{
		"MY_VAR": "hello",
		"PORT":   "1",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	want := fmt.Sprintf("hello|%d", port)
	if string(raw) != want {
		t.Fatalf("runtime env = %q, want %q", raw, want)
	}
}

// TestStartDoesNotBlockOtherSubdomains memastikan masa readiness (maksimal
// 10 detik) tidak memegang kunci global ProcessManager.
func TestStartDoesNotBlockOtherSubdomains(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	script := filepath.Join(t.TempDir(), "server.js")
	body := `
const http = require('http');
setTimeout(() => {
  http.createServer((req, res) => res.end('ok')).listen(Number(process.env.PORT), '127.0.0.1');
}, 1200);
`
	if err := os.WriteFile(script, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	pm := NewProcessManager()
	started := make(chan struct{})
	go func() {
		pm.Start("slow-site", script)
		close(started)
	}()

	// Beri waktu Start masuk ke loop readiness sebelum mengukur GetPort.
	time.Sleep(250 * time.Millisecond)
	begin := time.Now()
	pm.GetPort("other-site")
	if elapsed := time.Since(begin); elapsed > 500*time.Millisecond {
		t.Fatalf("GetPort was blocked for %v while another SSR server was starting", elapsed)
	}

	select {
	case <-started:
	case <-time.After(15 * time.Second):
		t.Fatal("slow SSR server did not finish starting")
	}
	pm.Stop("slow-site")
}
