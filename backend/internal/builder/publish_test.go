package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSite(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("gagal membuat direktori sumber: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatalf("gagal menulis index.html: %v", err)
	}
}

func readIndex(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatalf("gagal membaca index.html di %s: %v", dir, err)
	}
	return string(raw)
}

// publish harus mengganti isi situs dan menyimpan rilis lama sebagai cadangan.
func TestPublishSwapsReleaseAndKeepsBackup(t *testing.T) {
	root := t.TempDir()
	sitesDir := filepath.Join(root, "sites")
	src := filepath.Join(root, "src")
	dst := filepath.Join(sitesDir, "demo")

	writeSite(t, src, "v1")
	if err := publish(src, dst); err != nil {
		t.Fatalf("publish pertama gagal: %v", err)
	}
	if got := readIndex(t, dst); got != "v1" {
		t.Fatalf("isi situs = %q, ingin %q", got, "v1")
	}
	info, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("gagal lstat %s: %v", dst, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s bukan symlink (mode %v)", dst, info.Mode())
	}

	writeSite(t, src, "v2")
	if err := publish(src, dst); err != nil {
		t.Fatalf("publish kedua gagal: %v", err)
	}
	if got := readIndex(t, dst); got != "v2" {
		t.Fatalf("isi situs setelah publish kedua = %q, ingin %q", got, "v2")
	}

	releases, err := os.ReadDir(filepath.Join(sitesDir, ".releases", "demo"))
	if err != nil {
		t.Fatalf("gagal membaca folder rilis: %v", err)
	}
	if len(releases) < 2 {
		t.Fatalf("jumlah rilis tersimpan = %d, ingin minimal 2", len(releases))
	}
}

// Instalasi lama menyimpan situs sebagai direktori biasa; publish harus
// memindahkannya ke folder rilis tanpa kehilangan file.
func TestPublishMigratesLegacyDirectory(t *testing.T) {
	root := t.TempDir()
	sitesDir := filepath.Join(root, "sites")
	src := filepath.Join(root, "src")
	dst := filepath.Join(sitesDir, "legacy")

	writeSite(t, dst, "old")
	writeSite(t, src, "new")

	if err := publish(src, dst); err != nil {
		t.Fatalf("publish gagal: %v", err)
	}
	if got := readIndex(t, dst); got != "new" {
		t.Fatalf("isi situs = %q, ingin %q", got, "new")
	}
	info, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("gagal lstat %s: %v", dst, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("direktori lama tidak dikonversi menjadi symlink")
	}
}

// node_modules dan .git tidak boleh ikut dipublikasikan.
func TestPublishSkipsHeavyFolders(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "sites", "demo")

	writeSite(t, src, "ok")
	if err := os.MkdirAll(filepath.Join(src, "node_modules", "pkg"), 0o755); err != nil {
		t.Fatalf("gagal membuat node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "node_modules", "pkg", "index.js"), []byte("x"), 0o644); err != nil {
		t.Fatalf("gagal menulis file dependency: %v", err)
	}

	if err := publish(src, dst); err != nil {
		t.Fatalf("publish gagal: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("node_modules ikut dipublikasikan (err=%v)", err)
	}
}
