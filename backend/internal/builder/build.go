package builder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"only4.fun/backend/internal/config"
)

// Job adalah satu permintaan deployment.
type Job struct {
	ProjectID string
	UserID    string
	Subdomain string
	RepoURL   string
	EnvVars   map[string]string
	// Token adalah GitHub access token milik user, dipakai untuk repo privat.
	Token string
}

// Builder mengeksekusi clone + build. Build dijalankan langsung di host (VPS).
type Builder struct {
	cfg config.BuildConfig
}

// New membuat Builder.
func New(cfg config.BuildConfig) *Builder {
	return &Builder{cfg: cfg}
}

func (b *Builder) Mode() string { return "local" }

// Build menjalankan seluruh pipeline dan mengembalikan folder hasil build di SitesDir.
func (b *Builder) Build(ctx context.Context, job Job, logf func(string)) (string, error) {
	if err := validateSubdomain(job.Subdomain); err != nil {
		return "", err
	}

	timeout := time.Duration(b.cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := os.MkdirAll(b.cfg.CacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to prepare cache dir: %w", err)
	}
	if err := os.MkdirAll(b.cfg.SitesDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to prepare sites dir: %w", err)
	}

	workDir, err := os.MkdirTemp(b.cfg.CacheDir, fmt.Sprintf("build-%s-", job.ProjectID))
	if err != nil {
		return "", fmt.Errorf("failed to create workdir: %w", err)
	}
	defer os.RemoveAll(workDir)

	logf(fmt.Sprintf("build mode: %s", b.Mode()))
	logf(fmt.Sprintf("clone %s", job.RepoURL))
	if err := b.clone(ctx, job, workDir, logf); err != nil {
		return "", err
	}

	outputDir, err := b.runBuild(ctx, job, workDir, logf)
	if err != nil {
		return "", err
	}

	target := filepath.Join(b.cfg.SitesDir, job.Subdomain)
	logf(fmt.Sprintf("publish %s -> %s", outputDir, target))
	if err := publish(outputDir, target); err != nil {
		return "", err
	}
	logf("build finished")
	return target, nil
}

func (b *Builder) clone(ctx context.Context, job Job, workDir string, logf func(string)) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not available on the server")
	}

	cloneURL := authenticatedURL(job.RepoURL, job.Token)

	// Coba shallow clone dulu (jauh lebih cepat), lalu fallback ke clone penuh
	// bila server git tidak mendukungnya.
	attempts := [][]string{
		{"clone", "--depth", "1", "--quiet", cloneURL, workDir},
		{"clone", "--quiet", cloneURL, workDir},
	}
	var out []byte
	var err error
	for i, args := range attempts {
		if i > 0 {
			_ = os.RemoveAll(workDir)
			_ = os.MkdirAll(workDir, 0o755)
		}
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		out, err = cmd.CombinedOutput()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to clone repository: %s", lastLines(redact(string(out), job.Token), 8))
}

func (b *Builder) runBuild(ctx context.Context, job Job, workDir string, logf func(string)) (string, error) {
	isNode := fileExists(filepath.Join(workDir, "package.json"))

	if !isNode {
		// Situs statis murni: tidak ada langkah build.
		if fileExists(filepath.Join(workDir, "index.html")) {
			logf("no package.json found, serving the static files as-is")
			return workDir, nil
		}
		return "", fmt.Errorf("no package.json or index.html found at the repository root")
	}

	// Deteksi framework
	framework := detectFramework(workDir)
	logf(fmt.Sprintf("detected framework: %s", framework))

	// Deteksi package manager
	pm := detectPackageManager(workDir)
	logf(fmt.Sprintf("package manager: %s", pm))

	installs := []string{}
	switch {
	case fileExists(filepath.Join(workDir, "pnpm-lock.yaml")):
		installs = []string{"corepack enable || true", "pnpm install --frozen-lockfile"}
	case fileExists(filepath.Join(workDir, "yarn.lock")):
		installs = []string{"corepack enable || true", "yarn install --frozen-lockfile"}
	case fileExists(filepath.Join(workDir, "package-lock.json")):
		installs = []string{"npm ci --no-audit --no-fund"}
	default:
		installs = []string{"npm install --no-audit --no-fund"}
	}

	// Build command berdasarkan framework
	buildCmd := "npm run build"
	if !hasBuildScript(filepath.Join(workDir, "package.json")) {
		logf("package.json has no build script, skipping the build step")
		buildCmd = ""
	}

	// Framework-specific build commands
	switch framework {
	case "nextjs":
		// Next.js: coba build dengan export, fallback ke build biasa
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx next build"
		}
	case "nuxt":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx nuxt build"
		}
	case "astro":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx astro build"
		}
	case "sveltekit":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx svelte-kit build"
		}
	case "remix":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx remix build"
		}
	case "gatsby":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx gatsby build"
		}
	case "angular":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx ng build"
		}
	case "docusaurus":
		if !hasBuildScript(filepath.Join(workDir, "package.json")) {
			buildCmd = "npx docusaurus build"
		}
	}

	for _, step := range installs {
		logf("$ " + step)
		if err := b.exec(ctx, job, workDir, step, logf); err != nil {
			return "", err
		}
	}
	if buildCmd != "" {
		logf("$ " + buildCmd)
		if err := b.exec(ctx, job, workDir, buildCmd, logf); err != nil {
			return "", err
		}
	}

	outputDir, err := detectOutput(workDir)
	if err != nil {
		return "", err
	}
	logf("detected output: " + filepath.Base(outputDir))
	return outputDir, nil
}

func (b *Builder) exec(ctx context.Context, job Job, workDir, command string, logf func(string)) error {
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.Dir = workDir
	env := append(os.Environ(), "CI=1")
	for k, v := range job.EnvVars {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	stdout, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		return fmt.Errorf("failed to read build output: %w", pipeErr)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("command failed (%s): %w", command, err)
	}

	// Baca output baris per baris supaya progres build bisa ditampilkan
	// secara langsung, bukan menunggu perintah selesai.
	var tail []string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := redact(scanner.Text(), job.Token)
		if strings.TrimSpace(line) == "" {
			continue
		}
		logf(line)
		tail = append(tail, line)
		if len(tail) > 200 {
			tail = tail[len(tail)-200:]
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		_ = cmd.Wait()
		return fmt.Errorf("failed to read build output: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("build timed out")
		}
		return fmt.Errorf("command failed (%s): %s", command, lastLines(strings.Join(tail, "\n"), 6))
	}
	return nil
}

// detectOutput mencari folder hasil build yang umum dipakai.
func detectOutput(workDir string) (string, error) {
	candidates := []string{
		"out",              // Next.js (export), Nuxt
		"dist",             // Vite, Vue, Astro, Svelte, Angular
		"build",            // CRA, SvelteKit, Remix, Gatsby
		"public",           // Hugo, Jekyll, Gatsby
		".next",            // Next.js (SSR - butuh server, fallback ke static export)
		".svelte-kit",      // SvelteKit
		"_site",            // Jekyll, Eleventy
		"docs",             // Docusaurus
		".vitepress/dist",  // VitePress
		"storybook-static", // Storybook
	}

	for _, name := range candidates {
		dir := filepath.Join(workDir, name)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			if fileExists(filepath.Join(dir, "index.html")) {
				return dir, nil
			}
			if name == ".next" {
				// Next.js SSR: standalone output
				if fileExists(filepath.Join(dir, "standalone", "server.js")) {
					return filepath.Join(dir, "standalone"), nil
				}
				// Next.js SSR without standalone: .next/ with server/ directory
				if info, err := os.Stat(filepath.Join(dir, "server")); err == nil && info.IsDir() {
					// Check if next binary is available to run 'next start'
					if fileExists(filepath.Join(workDir, "node_modules", ".bin", "next")) {
						// Create a wrapper server.js that runs 'next start'
						wrapper := filepath.Join(workDir, "server.js")
						content := `const { spawn } = require('child_process');
const path = require('path');
const port = process.env.PORT || 3000;
const nextBin = path.join(__dirname, 'node_modules', '.bin', 'next');
const child = spawn(nextBin, ['start', '-p', String(port)], { stdio: 'inherit', cwd: __dirname, env: process.env });
child.on('exit', (code) => process.exit(code || 0));
`
						_ = os.WriteFile(wrapper, []byte(content), 0644)
						// Publish the whole project so node_modules is available
						return workDir, nil
					}
				}
				// Next.js static export
				if fileExists(filepath.Join(dir, "out", "index.html")) {
					return filepath.Join(dir, "out"), nil
				}
			}
		}
	}

	if fileExists(filepath.Join(workDir, "index.html")) {
		return workDir, nil
	}

	entries, err := os.ReadDir(workDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				dir := filepath.Join(workDir, entry.Name())
				if fileExists(filepath.Join(dir, "index.html")) {
					return dir, nil
				}
			}
		}
	}

	return "", fmt.Errorf("build output not found (looked for dist/, out/, build/, public/, .next/, or index.html)")
}

// detectPackageManager mendeteksi package manager dari lockfile.
func detectPackageManager(workDir string) string {
	if fileExists(filepath.Join(workDir, "bun.lockb")) {
		return "bun"
	}
	if fileExists(filepath.Join(workDir, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if fileExists(filepath.Join(workDir, "yarn.lock")) {
		return "yarn"
	}
	if fileExists(filepath.Join(workDir, "package-lock.json")) {
		return "npm"
	}
	return "npm"
}

// detectFramework mendeteksi framework dari package.json.
func detectFramework(workDir string) string {
	raw, err := os.ReadFile(filepath.Join(workDir, "package.json"))
	if err != nil {
		return ""
	}
	text := string(raw)
	lower := strings.ToLower(text)

	switch {
	case strings.Contains(lower, "\"next\""):
		return "nextjs"
	case strings.Contains(lower, "\"nuxt\""):
		return "nuxt"
	case strings.Contains(lower, "\"@sveltejs/kit\""):
		return "sveltekit"
	case strings.Contains(lower, "\"astro\""):
		return "astro"
	case strings.Contains(lower, "\"@remix-run\""):
		return "remix"
	case strings.Contains(lower, "\"vite\""):
		return "vite"
	case strings.Contains(lower, "\"vue\""):
		return "vue"
	case strings.Contains(lower, "\"react\""):
		return "react"
	case strings.Contains(lower, "\"angular\"") || strings.Contains(lower, "\"@angular/core\""):
		return "angular"
	case strings.Contains(lower, "\"gatsby\""):
		return "gatsby"
	case strings.Contains(lower, "\"@docusaurus/core\""):
		return "docusaurus"
	case strings.Contains(lower, "\"eleventy\"") || strings.Contains(lower, "\"@11ty/eleventy\""):
		return "eleventy"
	default:
		return "node"
	}
}

// hasBuildScript memeriksa keberadaan "build" di bagian scripts package.json.
func hasBuildScript(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(raw)
	idx := strings.Index(text, "\"scripts\"")
	if idx < 0 {
		return false
	}
	return strings.Contains(text[idx:], "\"build\"")
}

func (b *Builder) publish(outputDir, target string) error {
	return publish(outputDir, target)
}

// publish menerbitkan hasil build secara atomik.
//
// Setiap build disalin ke folder rilis terpisah di <sitesDir>/.releases/<subdomain>/,
// lalu <sitesDir>/<subdomain> adalah symlink yang di-rename ke rilis baru.
// Rename symlink bersifat atomik di POSIX, sehingga tidak ada momen ketika
// subdomain kehilangan folder aktif (versi lama tetap dipakai sampai detik terakhir).
func publish(src, dst string) error {
	parent := filepath.Dir(dst)
	base := filepath.Base(dst)
	releasesDir := filepath.Join(parent, ".releases", base)
	if err := os.MkdirAll(releasesDir, 0o755); err != nil {
		return fmt.Errorf("failed to prepare release folder: %w", err)
	}

	release, err := os.MkdirTemp(releasesDir, "release-")
	if err != nil {
		return fmt.Errorf("failed to prepare release folder: %w", err)
	}
	if err := copyTree(src, release); err != nil {
		_ = os.RemoveAll(release)
		return fmt.Errorf("failed to copy build output: %w", err)
	}

	linkTmp := filepath.Join(parent, fmt.Sprintf(".link-%s-%d", base, time.Now().UnixNano()))
	_ = os.Remove(linkTmp)
	if err := os.Symlink(release, linkTmp); err != nil {
		_ = os.RemoveAll(release)
		return fmt.Errorf("failed to create site link: %w", err)
	}

	// Migrasi sekali jalan untuk site lama yang masih berupa direktori biasa.
	var legacy string
	if info, statErr := os.Lstat(dst); statErr == nil && info.Mode()&os.ModeSymlink == 0 {
		legacy = filepath.Join(releasesDir, fmt.Sprintf("legacy-%d", time.Now().UnixNano()))
		if err := os.Rename(dst, legacy); err != nil {
			_ = os.Remove(linkTmp)
			_ = os.RemoveAll(release)
			return fmt.Errorf("failed to swap the previous site: %w", err)
		}
	}

	if err := os.Rename(linkTmp, dst); err != nil {
		if legacy != "" {
			_ = os.Rename(legacy, dst)
		}
		_ = os.Remove(linkTmp)
		_ = os.RemoveAll(release)
		return fmt.Errorf("failed to activate the build output: %w", err)
	}

	pruneReleases(releasesDir, 3)
	return nil
}

// pruneReleases menyimpan hanya beberapa rilis terbaru sebagai versi cadangan.
func pruneReleases(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type release struct {
		path    string
		modTime time.Time
	}
	tracked := make([]release, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		tracked = append(tracked, release{path: filepath.Join(dir, entry.Name()), modTime: info.ModTime()})
	}
	if len(tracked) <= keep {
		return
	}
	sort.Slice(tracked, func(i, j int) bool { return tracked[i].modTime.After(tracked[j].modTime) })
	for _, old := range tracked[keep:] {
		_ = os.RemoveAll(old.path)
	}
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Jangan ikut menyalin metadata repo dan dependency bila output di root repo.
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	})
}

// authenticatedURL menyisipkan token ke URL clone agar repo privat bisa diakses.
func authenticatedURL(repoURL, token string) string {
	if token == "" {
		return repoURL
	}
	u, err := url.Parse(repoURL)
	if err != nil || u.Host == "" || u.Scheme != "https" {
		return repoURL
	}
	u.User = url.UserPassword("x-access-token", token)
	return u.String()
}

func redact(s, secret string) string {
	if secret == "" {
		return s
	}
	return strings.ReplaceAll(s, secret, "***")
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func lastLines(s string, n int) string {
	lines := splitLines(s)
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func validateSubdomain(subdomain string) error {
	if subdomain == "" {
		return fmt.Errorf("project name is required")
	}
	if len(subdomain) > 63 {
		return fmt.Errorf("project name must be at most 63 characters")
	}
	for _, r := range subdomain {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			return fmt.Errorf("invalid subdomain")
		}
	}
	return nil
}
