package database

import (
	"fmt"
	"log"
	"time"

	"only4.fun/backend/internal/id"
)

// hasLegacyIntegerIDs melaporkan apakah tabel users masih memakai id INTEGER.
func hasLegacyIntegerIDs(db *Database) (bool, error) {
	var total int
	err := db.DB().Get(&total, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'`)
	if err != nil {
		return false, err
	}
	if total == 0 {
		return false, nil
	}

	var columnType string
	err = db.DB().Get(&columnType, `SELECT type FROM pragma_table_info('users') WHERE name = 'id'`)
	if err != nil {
		return false, nil
	}
	for i := 0; i+2 < len(columnType); i++ {
		if (columnType[i] == 'I' || columnType[i] == 'i') &&
			(columnType[i+1] == 'N' || columnType[i+1] == 'n') &&
			(columnType[i+2] == 'T' || columnType[i+2] == 't') {
			return true, nil
		}
	}
	return false, nil
}

type legacyUser struct {
	ID        int64     `db:"id"`
	GitHubID  string    `db:"github_id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	AvatarURL string    `db:"avatar_url"`
	CreatedAt time.Time `db:"created_at"`
}

type legacySession struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	SessionID string    `db:"session_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type legacyProject struct {
	ID            int64      `db:"id"`
	UserID        int64      `db:"user_id"`
	RepoURL       string     `db:"repo_url"`
	Subdomain     string     `db:"subdomain"`
	Status        string     `db:"status"`
	EnvVars       string     `db:"env_vars"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	Views         int64      `db:"views"`
	LastVisitedAt *time.Time `db:"last_visited_at"`
}

type legacyDeployment struct {
	ID        int64     `db:"id"`
	ProjectID int64     `db:"project_id"`
	Status    string    `db:"status"`
	Log       string    `db:"log"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// rebuildWithUUIDs mengubah semua id integer menjadi UUID v7 sambil menjaga relasi.
func rebuildWithUUIDs(db *Database) error {
	conn := db.DB()
	log.Printf("migration: converting integer ids to UUID v7")

	var users []legacyUser
	if err := conn.Select(&users, `SELECT id, github_id, username, COALESCE(email,'') AS email, COALESCE(avatar_url,'') AS avatar_url, created_at FROM users`); err != nil {
		return fmt.Errorf("read legacy users: %w", err)
	}

	var sessions []legacySession
	if err := conn.Select(&sessions, `SELECT id, user_id, session_id, token, expires_at, created_at FROM sessions`); err != nil {
		return fmt.Errorf("read legacy sessions: %w", err)
	}

	projectHasViews := columnExists(conn, "projects", "views")

	var projects []legacyProject
	projectQuery := `SELECT id, user_id, repo_url, subdomain, status, COALESCE(env_vars,'{}') AS env_vars, created_at, updated_at`
	if projectHasViews {
		projectQuery += `, COALESCE(views,0) AS views, last_visited_at`
	}
	projectQuery += ` FROM projects`
	if err := conn.Select(&projects, projectQuery); err != nil {
		return fmt.Errorf("read legacy projects: %w", err)
	}

	var deployments []legacyDeployment
	if err := conn.Select(&deployments, `SELECT id, project_id, status, COALESCE(log,'') AS log, created_at, updated_at FROM deployments`); err != nil {
		return fmt.Errorf("read legacy deployments: %w", err)
	}

	type legacyVisit struct {
		ProjectID int64  `db:"project_id"`
		Day       string `db:"day"`
		Count     int64  `db:"count"`
	}
	var visits []legacyVisit
	if columnExists(conn, "site_visits", "project_id") {
		_ = conn.Select(&visits, `SELECT project_id, day, count FROM site_visits`)
	}

	userMap := make(map[int64]string, len(users))
	projectMap := make(map[int64]string, len(projects))

	if _, err := conn.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		return err
	}
	for _, stmt := range []string{
		`DROP TABLE IF EXISTS site_visits`,
		`DROP TABLE IF EXISTS deployments`,
		`DROP TABLE IF EXISTS projects`,
		`DROP TABLE IF EXISTS sessions`,
		`DROP TABLE IF EXISTS users`,
	} {
		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("drop legacy tables: %w", err)
		}
	}

	if _, err := conn.Exec(schema); err != nil {
		return fmt.Errorf("create new schema: %w", err)
	}

	for _, u := range users {
		newID := id.New()
		userMap[u.ID] = newID
		if _, err := conn.Exec(
			`INSERT INTO users (id, github_id, username, email, avatar_url, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			newID, u.GitHubID, u.Username, u.Email, u.AvatarURL, u.CreatedAt,
		); err != nil {
			return fmt.Errorf("copy user: %w", err)
		}
	}

	for _, s := range sessions {
		newUserID, ok := userMap[s.UserID]
		if !ok {
			continue
		}
		if _, err := conn.Exec(
			`INSERT INTO sessions (id, user_id, session_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			id.New(), newUserID, s.SessionID, s.Token, s.ExpiresAt, s.CreatedAt,
		); err != nil {
			return fmt.Errorf("copy session: %w", err)
		}
	}

	for _, p := range projects {
		newUserID, ok := userMap[p.UserID]
		if !ok {
			continue
		}
		newID := id.New()
		projectMap[p.ID] = newID
		if _, err := conn.Exec(
			`INSERT INTO projects (id, user_id, repo_url, subdomain, status, env_vars, created_at, updated_at, views, last_visited_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newID, newUserID, p.RepoURL, p.Subdomain, p.Status, p.EnvVars, p.CreatedAt, p.UpdatedAt, p.Views, p.LastVisitedAt,
		); err != nil {
			return fmt.Errorf("copy project: %w", err)
		}
	}

	for _, d := range deployments {
		newProjectID, ok := projectMap[d.ProjectID]
		if !ok {
			continue
		}
		if _, err := conn.Exec(
			`INSERT INTO deployments (id, project_id, status, log, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
			id.New(), newProjectID, d.Status, d.Log, d.CreatedAt, d.UpdatedAt,
		); err != nil {
			return fmt.Errorf("copy deployment: %w", err)
		}
	}

	for _, v := range visits {
		newProjectID, ok := projectMap[v.ProjectID]
		if !ok {
			continue
		}
		if _, err := conn.Exec(
			`INSERT OR REPLACE INTO site_visits (project_id, day, count) VALUES (?, ?, ?)`,
			newProjectID, v.Day, v.Count,
		); err != nil {
			return fmt.Errorf("copy visits: %w", err)
		}
	}

	if _, err := conn.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return err
	}

	log.Printf("migration finished: %d users, %d sessions, %d projects, %d deployments", len(users), len(sessions), len(projects), len(deployments))
	return nil
}

func columnExists(conn interface {
	Get(dest interface{}, query string, args ...interface{}) error
}, table, column string) bool {
	var name string
	err := conn.Get(&name, fmt.Sprintf(`SELECT name FROM pragma_table_info('%s') WHERE name = ?`, table), column)
	return err == nil && name != ""
}
