package builder

import (
	"bufio"
	"context"
	"encoding/base64"
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
	"only4.fun/backend/internal/execenv"
)

// nextSSRMarker ditulis bersama wrapper server.js agar server tahu bahwa
// root server.js memang dibuat untuk Next.js, bukan file acuan dari repo.
const nextSSRMarker = ".only4-next-ssr"

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

// envHome is the only HOME exposed to build scripts. It lives under CACHE_DIR
// so npm caches do not depend on the server process's real home directory.
func (b *Builder) envHome() string {
	return filepath.Join(b.cfg.CacheDir, "home")
}

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
	if err := execenv.EnsureHome(b.envHome()); err != nil {
		return "", fmt.Errorf("failed to prepare build home: %w", err)
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

	// Log commit details for debugging redeploy / source changes.
	for _, info := range []struct {
		label string
		args  []string
	}{
		{"git commit", []string{"-C", workDir, "rev-parse", "HEAD"}},
		{"git branch", []string{"-C", workDir, "branch", "--show-current"}},
		{"git last commit", []string{"-C", workDir, "log", "--oneline", "-1"}},
	} {
		cmd := exec.CommandContext(ctx, "git", info.args...)
		cmd.Env = execenv.Base(b.envHome())
		if out, err := cmd.Output(); err == nil {
			logf(fmt.Sprintf("%s: %s", info.label, strings.TrimSpace(string(out))))
		}
	}

	outputDir, err := b.runBuild(ctx, job, workDir, logf)
	if err != nil {
		return "", err
	}

	target := filepath.Join(b.cfg.SitesDir, job.Subdomain)
	logf(fmt.Sprintf("publish %s -> %s", outputDir, target))
	if err := publishWithNodeModules(outputDir, target, needsRuntimeNodeModules(outputDir)); err != nil {
		return "", err
	}
	logf("build finished")
	return target, nil
}

func (b *Builder) clone(ctx context.Context, job Job, workDir string, logf func(string)) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not available on the server")
	}

	cloneURL := job.RepoURL
	authEnv := gitAuthEnv(job.RepoURL, job.Token)

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
		cmd.Env = append(append(execenv.Base(b.envHome()), "GIT_TERMINAL_PROMPT=0"), authEnv...)
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
	// Next.js standalone tidak menyertakan aset statis/public di dalam folder
	// standalone-nya. Tanpa ini CSS/JS/image 404 walaupun build sukses.
	if copied, err := prepareStandaloneAssets(workDir, outputDir); err != nil {
		return "", err
	} else if copied > 0 {
		logf(fmt.Sprintf("standalone assets: copied %d files", copied))
	}
	logf("detected output: " + filepath.Base(outputDir))
	return outputDir, nil
}

// prepareStandaloneAssets mengikuti rekomendasi deploy Next.js: sajikan
// .next/static dan public dari dalam folder .next/standalone. Mengembalikan
// jumlah file yang disalin supaya bisa dicatat di log build.
func prepareStandaloneAssets(workDir, outputDir string) (int, error) {
	if filepath.Base(outputDir) != "standalone" ||
		filepath.Base(filepath.Dir(outputDir)) != ".next" {
		return 0, nil
	}

	copied := 0
	assets := []struct{ src, dst string }{
		{filepath.Join(workDir, ".next", "static"), filepath.Join(outputDir, ".next", "static")},
		{filepath.Join(workDir, "public"), filepath.Join(outputDir, "public")},
	}
	for _, asset := range assets {
		info, err := os.Stat(asset.src)
		if err != nil || !info.IsDir() {
			continue
		}
		if err := copyTree(asset.src, asset.dst, "", true); err != nil {
			return copied, err
		}
		_ = filepath.Walk(asset.src, func(_ string, info os.FileInfo, err error) error {
			if err == nil && info != nil && !info.IsDir() {
				copied++
			}
			return nil
		})
	}
	return copied, nil
}

// needsRuntimeNodeModules menandakan output butuh dependency saat runtime.
// Next.js mode biasa (next start) memakai wrapper server.js + .next/server;
// standalone membawa node_modules-nya sendiri di dalam folder output.
func needsRuntimeNodeModules(outputDir string) bool {
	if !fileExists(filepath.Join(outputDir, "server.js")) {
		return false
	}
	return dirExists(filepath.Join(outputDir, ".next", "server")) ||
		dirExists(filepath.Join(outputDir, "node_modules"))
}

func (b *Builder) exec(ctx context.Context, job Job, workDir, command string, logf func(string)) error {
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.Dir = workDir
	// User-supplied keys are validated by the handler; the merge is repeated
	// here so legacy rows cannot override sandbox PATH/HOME or inject flags.
	cmd.Env = execenv.WithUserEnv(execenv.Base(b.envHome()), job.EnvVars)

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
				// Static export didahulukan agar tidak salah masuk mode SSR.
				if fileExists(filepath.Join(dir, "out", "index.html")) {
					return filepath.Join(dir, "out"), nil
				}
				// Next.js standalone: server + node_modules sudah dibundel.
				if fileExists(filepath.Join(dir, "standalone", "server.js")) {
					return filepath.Join(dir, "standalone"), nil
				}
				// Next.js biasa: jalankan 'next start' lewat wrapper server.js.
				// Pakai CLI asli (bukan symlink node_modules/.bin) agar tetap
				// bisa dipanggil setelah publish.
				nextCLI := filepath.Join(workDir, "node_modules", "next", "dist", "bin", "next")
				if serverInfo, err := os.Stat(filepath.Join(dir, "server")); err == nil &&
					serverInfo.IsDir() && fileExists(nextCLI) {
					wrapper := filepath.Join(workDir, "server.js")
					content := `const { spawn } = require('child_process');
const path = require('path');
const port = Number(process.env.PORT) || 3000;
const nextCLI = path.join(__dirname, 'node_modules', 'next', 'dist', 'bin', 'next');
const child = spawn(process.execPath, [nextCLI, 'start', '-p', String(port), '-H', '127.0.0.1'], {
  stdio: 'inherit',
  cwd: __dirname,
  env: { ...process.env, NODE_ENV: 'production' },
});
child.on('error', (err) => { console.error(err); process.exit(1); });
child.on('exit', (code) => process.exit(code || 0));
`
					if err := os.WriteFile(wrapper, []byte(content), 0o644); err != nil {
						return "", err
					}
					// Penanda ini dipakai server saat memilih entry SSR: Next.js
					// tanpa script "start" tetap boleh memakai wrapper ini.
					if err := os.WriteFile(filepath.Join(workDir, nextSSRMarker), []byte("next-start\n"), 0o644); err != nil {
						return "", err
					}
					// node_modules ikut dipublikasikan (lihat needsRuntimeNodeModules).
					return workDir, nil
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
	return publishWithNodeModules(outputDir, target, needsRuntimeNodeModules(outputDir))
}

// publish menerbitkan hasil build secara atomik.
//
// Setiap build disalin ke folder rilis terpisah di <sitesDir>/.releases/<subdomain>/,
// lalu <sitesDir>/<subdomain> adalah symlink yang di-rename ke rilis baru.
// Rename symlink bersifat atomik di POSIX, sehingga tidak ada momen ketika
// subdomain kehilangan folder aktif (versi lama tetap dipakai sampai detik terakhir).
func publish(src, dst string) error {
	return publishWithNodeModules(src, dst, false)
}

func publishWithNodeModules(src, dst string, includeNodeModules bool) error {
	parent := filepath.Dir(dst)
	base := filepath.Base(dst)
	releasesDir := filepath.Join(parent, ".releases", base)
	if err := os.MkdirAll(releasesDir, 0o755); err != nil {
		return fmt.Errorf("failed to prepare release folder: %w", err)
	}

	// Cari rilis sebelumnya (yang masih aktif) sebagai referensi incremental.
	var previousRelease string
	entries, _ := os.ReadDir(releasesDir)
	if len(entries) > 0 {
		type relInfo struct {
			name string
			mod  time.Time
		}
		var rels []relInfo
		for _, e := range entries {
			// Info() mengikuti symlink; entry bisa hilang atau rusak saat
			// housekeeping berjalan. Jangan pernah memakai info nil.
			info, err := e.Info()
			if err != nil || info == nil {
				continue
			}
			rels = append(rels, relInfo{name: e.Name(), mod: info.ModTime()})
		}
		sort.Slice(rels, func(i, j int) bool { return rels[i].mod.After(rels[j].mod) })
		// Rilis terbaru (rels[0]) adalah yang sedang aktif, jadi kita bandingkan
		// dengan rilis itu untuk incremental copy.
		if len(rels) >= 1 {
			previousRelease = filepath.Join(releasesDir, rels[0].name)
		}
	}

	release, err := os.MkdirTemp(releasesDir, "release-")
	if err != nil {
		return fmt.Errorf("failed to prepare release folder: %w", err)
	}
	if err := copyTree(src, release, previousRelease, includeNodeModules); err != nil {
		_ = os.RemoveAll(release)
		return fmt.Errorf("failed to copy build output: %w", err)
	}

	linkTmp := filepath.Join(parent, fmt.Sprintf(".link-%s-%d", base, time.Now().UnixNano()))
	_ = os.Remove(linkTmp)
	// Target mutlak: SITES_DIR bisa berupa path relatif, sehingga symlink
	// relatif akan menunjuk ke folder yang salah saat situs diakses.
	releaseAbs, err := filepath.Abs(release)
	if err != nil {
		_ = os.RemoveAll(release)
		return fmt.Errorf("failed to resolve the release folder: %w", err)
	}
	if err := os.Symlink(releaseAbs, linkTmp); err != nil {
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

func copyTree(src, dst string, previousRelease string, includeNodeModules bool) error {
	srcRoot, err := filepath.EvalSymlinks(src)
	if err != nil {
		srcRoot = src
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		// .git selalu dibuang. node_modules hanya ikut untuk output SSR yang
		// memang butuh runtime dependency (Next.js next start / standalone).
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if info.IsDir() && info.Name() == "node_modules" && !includeNodeModules {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			// pnpm/npm workspaces memakai symlink di node_modules. Symlink yang
			// keluar dari folder build sengaja dibuang agar tidak bocor atau
			// menunjuk ke file server.
			if !includeNodeModules {
				return nil
			}
			link, err := os.Readlink(path)
			if err != nil {
				return nil
			}
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil
			}
			relResolved, err := filepath.Rel(srcRoot, resolved)
			if err != nil || relResolved == ".." || strings.HasPrefix(relResolved, ".."+string(os.PathSeparator)) {
				return nil
			}
			if filepath.IsAbs(link) {
				link, err = filepath.Rel(filepath.Dir(target), resolved)
				if err != nil {
					return nil
				}
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			return os.Symlink(link, target)
		}

		// Jika ada previousRelease dan file sudah ada di previousRelease dengan ukuran dan waktu yang sama,
		// gunakan hardlink agar tidak perlu menyalin ulang.
		if previousRelease != "" {
			prevPath := filepath.Join(previousRelease, rel)
			if info.IsDir() {
				prevInfo, _ := os.Stat(prevPath)
				if prevInfo != nil && prevInfo.IsDir() && prevInfo.ModTime().Equal(info.ModTime()) && prevInfo.Size() == info.Size() {
					return nil
				}
			} else {
				prevInfo, _ := os.Stat(prevPath)
				if prevInfo != nil && prevInfo.Size() == info.Size() && prevInfo.ModTime().Equal(info.ModTime()) {
					return os.Link(prevPath, target)
				}
			}
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

// gitAuthEnv mengirim kredensial lewat konfigurasi Git di environment, bukan
// di URL clone, agar token tidak muncul di `ps` maupun tersimpan di
// .git/config pada direktori kerja.
func gitAuthEnv(repoURL, token string) []string {
	if token == "" {
		return nil
	}
	u, err := url.Parse(repoURL)
	if err != nil || u.Host == "" || u.Scheme != "https" {
		return nil
	}
	base := u.Scheme + "://" + u.Host + "/"
	cred := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))
	return []string{
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=http." + base + ".extraheader",
		"GIT_CONFIG_VALUE_0=Authorization: Basic " + cred,
	}
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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
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
