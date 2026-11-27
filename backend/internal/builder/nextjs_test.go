package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestNeedsRuntimeNodeModules(t *testing.T) {
	root := t.TempDir()
	if needsRuntimeNodeModules(root) {
		t.Fatal("empty output should not keep node_modules")
	}

	writeFile(t, filepath.Join(root, "server.js"), "")
	if needsRuntimeNodeModules(root) {
		t.Fatal("server.js alone should not keep node_modules")
	}

	if err := os.MkdirAll(filepath.Join(root, ".next", "server"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !needsRuntimeNodeModules(root) {
		t.Fatal("Next.js next start output should keep node_modules")
	}

	standalone := t.TempDir()
	writeFile(t, filepath.Join(standalone, "server.js"), "")
	writeFile(t, filepath.Join(standalone, "node_modules", "next", "package.json"), "{}")
	if !needsRuntimeNodeModules(standalone) {
		t.Fatal("standalone output should keep its bundled node_modules")
	}
}

func TestPrepareStandaloneAssets(t *testing.T) {
	workDir := t.TempDir()
	standalone := filepath.Join(workDir, ".next", "standalone")
	writeFile(t, filepath.Join(standalone, "server.js"), "")
	writeFile(t, filepath.Join(workDir, ".next", "static", "css", "app.css"), "body{}")
	writeFile(t, filepath.Join(workDir, "public", "logo.svg"), "<svg/>")

	copied, err := prepareStandaloneAssets(workDir, standalone)
	if err != nil {
		t.Fatalf("prepareStandaloneAssets: %v", err)
	}
	if copied != 2 {
		t.Fatalf("copied = %d, want 2", copied)
	}
	for _, rel := range []string{
		filepath.Join(".next", "static", "css", "app.css"),
		filepath.Join("public", "logo.svg"),
	} {
		if _, err := os.Stat(filepath.Join(standalone, rel)); err != nil {
			t.Fatalf("missing standalone asset %s: %v", rel, err)
		}
	}
}

func TestPublishKeepsNodeModulesForSSR(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "sites", "demo")
	writeFile(t, filepath.Join(src, "server.js"), "")
	writeFile(t, filepath.Join(src, ".next", "server", "app.js"), "")
	writeFile(t, filepath.Join(src, "node_modules", "next", "dist", "bin", "next"), "cli")

	if err := publishWithNodeModules(src, dst, true); err != nil {
		t.Fatalf("publishWithNodeModules: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "node_modules", "next", "dist", "bin", "next")); err != nil {
		t.Fatalf("runtime dependency was not published: %v", err)
	}
}

func TestPublishKeepsInternalSymlinksForSSR(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "sites", "demo")
	writeFile(t, filepath.Join(src, ".pnpm", "pkg", "index.js"), "module.exports = 1")
	if err := os.MkdirAll(filepath.Join(src, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../.pnpm/pkg", filepath.Join(src, "node_modules", "pkg")); err != nil {
		t.Fatal(err)
	}

	if err := publishWithNodeModules(src, dst, true); err != nil {
		t.Fatalf("publishWithNodeModules: %v", err)
	}
	target := filepath.Join(dst, "node_modules", "pkg")
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("symlink was not published: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("published node_modules/pkg is not a symlink")
	}
	if _, err := os.Stat(filepath.Join(target, "index.js")); err != nil {
		t.Fatalf("symlink target is broken: %v", err)
	}
}

func TestDetectOutputNextStartWrapper(t *testing.T) {
	workDir := t.TempDir()
	writeFile(t, filepath.Join(workDir, "package.json"), `{"scripts":{"build":"next build","start":"next start"}}`)
	if err := os.MkdirAll(filepath.Join(workDir, ".next", "server"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(workDir, "node_modules", "next", "dist", "bin", "next"), "cli")

	outputDir, err := detectOutput(workDir)
	if err != nil {
		t.Fatalf("detectOutput: %v", err)
	}
	if outputDir != workDir {
		t.Fatalf("outputDir = %q, want %q", outputDir, workDir)
	}
	wrapper, err := os.ReadFile(filepath.Join(workDir, "server.js"))
	if err != nil {
		t.Fatalf("read wrapper: %v", err)
	}
	for _, want := range []string{"process.execPath", "'next', 'dist', 'bin', 'next'", "NODE_ENV: 'production'"} {
		if !strings.Contains(string(wrapper), want) {
			t.Fatalf("wrapper missing %q:\n%s", want, wrapper)
		}
	}
	if strings.Contains(string(wrapper), "node_modules/.bin/next") {
		t.Fatalf("wrapper still uses the .bin symlink:\n%s", wrapper)
	}
	if _, err := os.Stat(filepath.Join(workDir, nextSSRMarker)); err != nil {
		t.Fatalf("Next.js SSR marker was not written: %v", err)
	}
}

func TestDetectOutputPrefersStaticExport(t *testing.T) {
	workDir := t.TempDir()
	writeFile(t, filepath.Join(workDir, ".next", "server", "app.js"), "")
	writeFile(t, filepath.Join(workDir, ".next", "out", "index.html"), "static")

	outputDir, err := detectOutput(workDir)
	if err != nil {
		t.Fatalf("detectOutput: %v", err)
	}
	want := filepath.Join(workDir, ".next", "out")
	if outputDir != want {
		t.Fatalf("outputDir = %q, want %q", outputDir, want)
	}
	if _, err := os.Stat(filepath.Join(workDir, "server.js")); !os.IsNotExist(err) {
		t.Fatalf("static export should not create an SSR wrapper: %v", err)
	}
}
