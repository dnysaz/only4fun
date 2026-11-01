package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActivateRelease(t *testing.T) {
	sitesDir := t.TempDir()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "index.html"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sitesDir, "demo")
	if err := Publish(src, target); err != nil {
		t.Fatalf("publish pertama: %v", err)
	}
	first, err := CurrentReleaseName(target)
	if err != nil {
		t.Fatalf("CurrentReleaseName: %v", err)
	}

	if err := os.WriteFile(filepath.Join(src, "index.html"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Publish(src, target); err != nil {
		t.Fatalf("publish kedua: %v", err)
	}
	second, err := CurrentReleaseName(target)
	if err != nil {
		t.Fatalf("CurrentReleaseName: %v", err)
	}
	if first == second {
		t.Fatalf("rilis kedua seharusnya berbeda dari %q", first)
	}

	if err := ActivateRelease(target, first); err != nil {
		t.Fatalf("ActivateRelease: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(target, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "v1" {
		t.Fatalf("isi situs = %q, ingin v1", content)
	}

	if err := ActivateRelease(target, "../../etc"); err == nil {
		t.Fatal("ActivateRelease menerima nama rilis tidak valid")
	}
	if err := ActivateRelease(target, "release-tidak-ada"); err == nil {
		t.Fatal("ActivateRelease menerima rilis yang tidak ada")
	}
}
