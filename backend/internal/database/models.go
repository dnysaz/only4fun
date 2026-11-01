package database

import (
	"time"
)

type User struct {
	ID        string    `db:"id" json:"id"`
	GitHubID  string    `db:"github_id" json:"github_id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	AvatarURL string    `db:"avatar_url" json:"avatar_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Session struct {
	ID        string    `db:"id" json:"-"`
	UserID    string    `db:"user_id" json:"-"`
	SessionID string    `db:"session_id" json:"session_id"`
	Token     string    `db:"token" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Project struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"user_id"`
	RepoURL   string `db:"repo_url" json:"repo_url"`
	Subdomain string `db:"subdomain" json:"subdomain"`
	Status    string `db:"status" json:"status"`
	// EnvVars berisi nilai rahasia, jadi tidak pernah di-serialize ke response.
	// Klien hanya menerima daftar nama key lewat EnvKeys.
	EnvVars string `db:"env_vars" json:"-"`
	// WebhookSecret adalah secret HMAC untuk webhook GitHub project ini.
	// Hanya pemilik yang bisa melihatnya lewat endpoint khusus.
	WebhookSecret string `db:"webhook_secret" json:"-"`
	// CustomDomain adalah domain milik user yang diarahkan ke project ini.
	CustomDomain  string     `db:"custom_domain" json:"custom_domain,omitempty"`
	EnvKeys       []string   `db:"-" json:"env_keys"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	DeployURL     string     `db:"deploy_url" json:"deploy_url,omitempty"`
	Views         int64      `db:"views" json:"views"`
	LastVisitedAt *time.Time `db:"last_visited_at" json:"last_visited_at,omitempty"`
	// Visits diisi hanya pada endpoint detail.
	Visits []DailyVisit `db:"-" json:"visits,omitempty"`
}

// DailyVisit adalah jumlah kunjungan per hari untuk satu project.
type DailyVisit struct {
	Day   string `db:"day" json:"day"`
	Count int64  `db:"count" json:"count"`
} // WebhookInfo adalah data yang ditampilkan ke pemilik project agar bisa
// mendaftarkan webhook di GitHub.
type WebhookInfo struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

type Deployment struct {
	ID        string `db:"id" json:"id"`
	ProjectID string `db:"project_id" json:"project_id"`
	Status    string `db:"status" json:"status"`
	Log       string `db:"log" json:"log,omitempty"`
	// Release adalah nama folder rilis yang dipublikasikan deployment ini.
	// Kosong untuk deployment lama yang dibuat sebelum fitur rollback ada.
	Release   string    `db:"release" json:"release,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
