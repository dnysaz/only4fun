package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	GitHub     GitHubConfig
	Cloudflare CloudflareConfig
	Build      BuildConfig
}

type ServerConfig struct {
	Addr string `env:"SERVER_ADDR" default:":8080"`
	// FrontendURL dipakai sebagai tujuan redirect setelah login berhasil.
	FrontendURL string `env:"FRONTEND_URL" default:"http://localhost:3000"`
	// CORSOrigins membatasi origin yang boleh mengakses API.
	CORSOrigins []string
}

type DatabaseConfig struct {
	Path string `env:"DATABASE_PATH" default:"data/only4.fun.db"`
}

type GitHubConfig struct {
	ClientID     string `env:"GITHUB_CLIENT_ID"`
	ClientSecret string `env:"GITHUB_CLIENT_SECRET"`
	RedirectURL  string `env:"GITHUB_REDIRECT_URL" default:"http://localhost:8080/auth/github/callback"`
}

type CloudflareConfig struct {
	APIKey string `env:"CLOUDFLARE_API_KEY"`
	ZoneID string `env:"CLOUDFLARE_ZONE_ID"`
	Domain string `env:"CLOUDFLARE_DOMAIN" default:"only4.fun"`
	Proxy  bool   `env:"CLOUDFLARE_PROXY" default:"true"`
	// Target adalah tujuan record subdomain: IP VPS (A) atau hostname (CNAME).
	Target string `env:"CLOUDFLARE_TARGET"`
}

type BuildConfig struct {
	DockerSocket string `env:"DOCKER_SOCKET" default:"/var/run/docker.sock"`
	BuilderImage string `env:"BUILDER_IMAGE" default:"node:20-alpine"`
	CacheDir     string `env:"CACHE_DIR" default:"/tmp/only4.fun/cache"`
	// SitesDir adalah folder tempat hasil build dipublikasikan.
	SitesDir string `env:"SITES_DIR" default:"./data/sites"`
	// Mode menentukan cara build: "docker", "local", atau "auto" (docker bila tersedia).
	Mode string `env:"BUILDER_MODE" default:"auto"`
	// AllowLocal harus true secara eksplisit agar build dijalankan langsung di
	// host. Default false karena build lokal mengeksekusi script repo pengguna.
	AllowLocal bool `env:"BUILDER_ALLOW_LOCAL" default:"false"`
	// TimeoutSeconds membatasi durasi satu build.
	TimeoutSeconds int `env:"BUILD_TIMEOUT_SECONDS" default:"600"`
}

func Load() (*Config, error) {
	godotenv.Load()
	cfg := &Config{
		Server: ServerConfig{
			Addr:        getEnv("SERVER_ADDR", ":8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
			CORSOrigins: getEnvList("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DATABASE_PATH", "data/only4.fun.db"),
		},
		GitHub: GitHubConfig{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),
		},
		Cloudflare: CloudflareConfig{
			APIKey: os.Getenv("CLOUDFLARE_API_KEY"),
			ZoneID: os.Getenv("CLOUDFLARE_ZONE_ID"),
			Domain: getEnv("CLOUDFLARE_DOMAIN", "only4.fun"),
			Proxy:  getEnvBool("CLOUDFLARE_PROXY", true),
			Target: os.Getenv("CLOUDFLARE_TARGET"),
		},
		Build: BuildConfig{
			DockerSocket:   getEnv("DOCKER_SOCKET", "/var/run/docker.sock"),
			BuilderImage:   getEnv("BUILDER_IMAGE", "node:20-alpine"),
			CacheDir:       getEnv("CACHE_DIR", "/tmp/only4.fun/cache"),
			SitesDir:       getEnv("SITES_DIR", "./data/sites"),
			Mode:           getEnv("BUILDER_MODE", "auto"),
			AllowLocal:     getEnvBool("BUILDER_ALLOW_LOCAL", false),
			TimeoutSeconds: getEnvInt("BUILD_TIMEOUT_SECONDS", 600),
		},
	}
	return cfg, nil
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func getEnvList(key, defaultVal string) []string {
	raw := getEnv(key, defaultVal)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return []string{"http://localhost:3000"}
	}
	return out
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}
