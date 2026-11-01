package handler

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "site.zip")
	writeZip(t, src, map[string]string{
		"index.html":       "<h1>hi</h1>",
		"assets/style.css": "body{}",
	})

	dst := t.TempDir()
	if err := extractZip(src, dst); err != nil {
		t.Fatalf("extractZip error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "index.html")); err != nil {
		t.Fatalf("index.html tidak diekstrak: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "assets", "style.css")); err != nil {
		t.Fatalf("aset tidak diekstrak: %v", err)
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "evil.zip")
	writeZip(t, src, map[string]string{
		"../../evil.txt": "pwned",
	})

	parent := t.TempDir()
	dst := filepath.Join(parent, "site")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	// Hasil boleh error atau dinormalkan, yang penting file tidak keluar
	// dari folder ekstraksi.
	_ = extractZip(src, dst)

	for _, escaped := range []string{
		filepath.Join(parent, "evil.txt"),
		filepath.Join(filepath.Dir(parent), "evil.txt"),
	} {
		if _, err := os.Stat(escaped); err == nil {
			t.Fatalf("file keluar dari folder ekstraksi: %s", escaped)
		}
	}
}

func TestStripSingleRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "site")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "index.html"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := stripSingleRoot(root); got != nested {
		t.Fatalf("stripSingleRoot = %q, ingin %q", got, nested)
	}

	// Dua entri di root berarti folder root dipakai apa adanya.
	if err := os.WriteFile(filepath.Join(root, "extra.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := stripSingleRoot(root); got != root {
		t.Fatalf("stripSingleRoot = %q, ingin %q", got, root)
	}
}
