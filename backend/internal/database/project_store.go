package database

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"only4.fun/backend/internal/id"
	"only4.fun/backend/internal/secure"

	"github.com/jmoiron/sqlx"
)

type ProjectStore struct {
	db *sqlx.DB
}

func NewProjectStore(db *Database) *ProjectStore {
	return &ProjectStore{db: db.DB()}
}

func (s *ProjectStore) Create(userID string, repoURL, subdomain string, envVars map[string]string) (*Project, error) {
	envJSON, _ := json.Marshal(envVars)
	encryptedEnv, err := secure.Encrypt(string(envJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt environment variables: %w", err)
	}
	newID := id.New()
	// Secret webhook dibuat sekali per project dan ditampilkan ke pemiliknya.
	webhookSecret, err := secure.NewSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to create the webhook secret: %w", err)
	}
	encryptedSecret, err := secure.Encrypt(webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt the webhook secret: %w", err)
	}
	query := `INSERT INTO projects (id, user_id, repo_url, subdomain, env_vars, webhook_secret) VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := s.db.Exec(query, newID, userID, repoURL, subdomain, encryptedEnv, encryptedSecret); err != nil {
		return nil, err
	}
	project := Project{
		ID:            newID,
		UserID:        userID,
		RepoURL:       repoURL,
		Subdomain:     subdomain,
		Status:        "pending",
		EnvVars:       encryptedEnv,
		WebhookSecret: encryptedSecret,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return &project, nil
}

// decryptProjectEnv membuka env vars yang terenkripsi sebelum dikeluarkan.
func decryptProjectEnv(project *Project) {
	if project == nil {
		return
	}
	if decrypted, err := secure.Decrypt(project.EnvVars); err == nil && decrypted != "" {
		project.EnvVars = decrypted
	}
}

func (s *ProjectStore) GetByID(projectID string) (*Project, error) {
	var project Project
	query := `SELECT * FROM projects WHERE id = ?`
	err := s.db.Get(&project, query, projectID)
	if err == nil {
		decryptProjectEnv(&project)
	}
	return &project, err
}

func (s *ProjectStore) GetByUserID(userID string) ([]*Project, error) {
	var projects []*Project
	query := `SELECT * FROM projects WHERE user_id = ? ORDER BY created_at DESC`
	err := s.db.Select(&projects, query, userID)
	if err == nil {
		for _, p := range projects {
			decryptProjectEnv(p)
		}
	}
	return projects, err
}

// ParseEnvKeys mengembalikan daftar nama key env (tanpa nilai) dari JSON env vars.
func ParseEnvKeys(envJSON string) []string {
	keys := []string{}
	if envJSON == "" {
		return keys
	}
	// Decrypt jika terenkripsi.
	if decrypted, err := secure.Decrypt(envJSON); err == nil && decrypted != "" {
		envJSON = decrypted
	}
	var envVars map[string]string
	if err := json.Unmarshal([]byte(envJSON), &envVars); err != nil {
		return keys
	}
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *ProjectStore) GetBySubdomain(subdomain string) (*Project, error) {
	var project Project
	query := `SELECT * FROM projects WHERE subdomain = ?`
	err := s.db.Get(&project, query, subdomain)
	if err == nil {
		decryptProjectEnv(&project)
	}
	return &project, err
}

func (s *ProjectStore) Update(project *Project) error {
	project.UpdatedAt = time.Now()
	encrypted, err := secure.Encrypt(project.EnvVars)
	if err != nil {
		return fmt.Errorf("failed to encrypt environment variables: %w", err)
	}
	query := `UPDATE projects SET repo_url = ?, subdomain = ?, status = ?, env_vars = ?, updated_at = ? WHERE id = ?`
	_, err = s.db.Exec(query, project.RepoURL, project.Subdomain, project.Status, encrypted, project.UpdatedAt, project.ID)
	return err
}

func (s *ProjectStore) UpdateStatus(projectID string, status string) error {
	query := `UPDATE projects SET status = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, time.Now(), projectID)
	return err
}

func (s *ProjectStore) Delete(projectID string) error {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := s.db.Exec(query, projectID)
	return err
}

func (s *ProjectStore) GetEnvVars(projectID string) (map[string]string, error) {
	var project Project
	if err := s.db.Get(&project, `SELECT env_vars FROM projects WHERE id = ?`, projectID); err != nil {
		return nil, err
	}
	decryptProjectEnv(&project)
	var envVars map[string]string
	if err := json.Unmarshal([]byte(project.EnvVars), &envVars); err != nil {
		return make(map[string]string), nil
	}
	return envVars, nil
}

// RecordVisit menambah hitungan kunjungan project (total dan per hari).
func (s *ProjectStore) RecordVisit(projectID string) error {
	if _, err := s.db.Exec(`UPDATE projects SET views = views + 1, last_visited_at = CURRENT_TIMESTAMP WHERE id = ?`, projectID); err != nil {
		return err
	}
	_, err := s.db.Exec(`
		INSERT INTO site_visits (project_id, day, count) VALUES (?, date('now'), 1)
		ON CONFLICT(project_id, day) DO UPDATE SET count = count + 1`, projectID)
	return err
}

// GetVisits mengembalikan kunjungan harian beberapa hari terakhir.
func (s *ProjectStore) GetVisits(projectID string, days int) ([]DailyVisit, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	visits := []DailyVisit{}
	query := `SELECT day, count FROM site_visits WHERE project_id = ? AND day >= date('now', ?) ORDER BY day ASC`
	err := s.db.Select(&visits, query, projectID, fmt.Sprintf("-%d day", days-1))
	return visits, err
}

// GetByCustomDomain mengambil project yang memakai domain kustom tertentu.
func (s *ProjectStore) GetByCustomDomain(domain string) (*Project, error) {
	var project Project
	query := `SELECT * FROM projects WHERE lower(custom_domain) = lower(?)`
	err := s.db.Get(&project, query, strings.TrimSpace(domain))
	if err == nil {
		decryptProjectEnv(&project)
	}
	return &project, err
}

// UpdateCustomDomain menyimpan (atau menghapus bila kosong) domain kustom.
func (s *ProjectStore) UpdateCustomDomain(projectID, domain string) error {
	_, err := s.db.Exec(`UPDATE projects SET custom_domain = ?, updated_at = ? WHERE id = ?`, domain, time.Now(), projectID)
	return err
}

// GetByRepoURL mengambil semua project yang memakai URL repository tertentu.
// Satu repo bisa di-deploy ke beberapa project (subdomain berbeda).
func (s *ProjectStore) GetByRepoURL(repoURL string) ([]*Project, error) {
	projects := []*Project{}
	query := `SELECT * FROM projects WHERE lower(repo_url) = lower(?)`
	err := s.db.Select(&projects, query, strings.TrimSpace(repoURL))
	return projects, err
}

// EnsureWebhookSecret mengembalikan secret webhook project, membuatnya bila
// project lama belum punya. Secret disimpan terenkripsi.
func (s *ProjectStore) EnsureWebhookSecret(projectID string) (string, error) {
	var stored string
	if err := s.db.Get(&stored, `SELECT webhook_secret FROM projects WHERE id = ?`, projectID); err != nil {
		return "", err
	}
	if stored != "" {
		if plain, err := secure.Decrypt(stored); err == nil && plain != "" {
			return plain, nil
		}
	}

	secret, err := secure.NewSecret()
	if err != nil {
		return "", err
	}
	encrypted, err := secure.Encrypt(secret)
	if err != nil {
		return "", err
	}
	if _, err := s.db.Exec(`UPDATE projects SET webhook_secret = ? WHERE id = ?`, encrypted, projectID); err != nil {
		return "", err
	}
	return secret, nil
}

func (s *ProjectStore) UpdateEnvVars(projectID string, envVars map[string]string) error {
	// Ambil env vars yang ada sekarang.
	existing, err := s.GetEnvVars(projectID)
	if err != nil {
		existing = make(map[string]string)
	}

	// Merge: key baru mengganti value lama; value kosong menghapus key.
	for k, v := range envVars {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if strings.TrimSpace(v) == "" {
			delete(existing, k)
		} else {
			existing[k] = v
		}
	}

	envJSON, _ := json.Marshal(existing)
	encrypted, err := secure.Encrypt(string(envJSON))
	if err != nil {
		return fmt.Errorf("failed to encrypt environment variables: %w", err)
	}
	query := `UPDATE projects SET env_vars = ?, updated_at = ? WHERE id = ?`
	_, err = s.db.Exec(query, encrypted, time.Now(), projectID)
	return err
}
