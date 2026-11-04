package database

const schema = `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		github_id TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		email TEXT,
		avatar_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		session_id TEXT UNIQUE NOT NULL,
		token TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		repo_url TEXT NOT NULL,
		subdomain TEXT UNIQUE NOT NULL,
		status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'building', 'deployed', 'failed')),
		env_vars TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		views INTEGER DEFAULT 0,
		last_visited_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS deployments (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		status TEXT NOT NULL CHECK(status IN ('pending', 'building', 'success', 'failed')),
		log TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS site_visits (
		project_id TEXT NOT NULL,
		day TEXT NOT NULL,
		count INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (project_id, day),
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);
`

// Migrate menyiapkan skema. Database lama yang masih memakai id INTEGER
// dibangun ulang dengan id UUID v7 sebelum skema baru diterapkan.
func Migrate(db *Database) error {
	legacy, err := hasLegacyIntegerIDs(db)
	if err != nil {
		return err
	}
	if legacy {
		if err := rebuildWithUUIDs(db); err != nil {
			return err
		}
	}

	if _, err := db.DB().Exec(schema); err != nil {
		return err
	}

	// Kolom tambahan untuk database lama yang belum punya statistik kunjungan.
	for _, stmt := range []string{
		`ALTER TABLE projects ADD COLUMN views INTEGER DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN last_visited_at DATETIME`,
		// release menyimpan nama folder rilis (.releases/<subdomain>/<release>)
		// agar deployment lama bisa dipromosikan kembali (rollback).
		`ALTER TABLE deployments ADD COLUMN release TEXT DEFAULT ''`,
		// webhook_secret menyimpan secret HMAC webhook GitHub (terenkripsi).
		`ALTER TABLE projects ADD COLUMN webhook_secret TEXT DEFAULT ''`,
		// custom_domain memungkinkan project disajikan di domain milik user.
		`ALTER TABLE projects ADD COLUMN custom_domain TEXT DEFAULT ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_custom_domain ON projects(custom_domain) WHERE custom_domain <> ''`,
	} {
		_, _ = db.DB().Exec(stmt)
	}

	// Index tambahan untuk query yang sering dipakai.
	_, _ = db.DB().Exec(`CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id)`)
	_, _ = db.DB().Exec(`CREATE INDEX IF NOT EXISTS idx_projects_subdomain ON projects(subdomain)`)
	_, _ = db.DB().Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`)
	_, _ = db.DB().Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`)
	return nil
}
