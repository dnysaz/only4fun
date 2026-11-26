package database

import (
	"time"

	"only4.fun/backend/internal/id"

	"github.com/jmoiron/sqlx"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *Database) *UserStore {
	return &UserStore{db: db.DB()}
}

func (s *UserStore) Create(githubID, username, email, avatarURL string) (*User, error) {
	var user User
	newID := id.New()
	query := `INSERT INTO users (id, github_id, username, email, avatar_url) VALUES (?, ?, ?, ?, ?)`
	if _, err := s.db.Exec(query, newID, githubID, username, email, avatarURL); err != nil {
		return nil, err
	}
	user = User{
		ID:        newID,
		GitHubID:  githubID,
		Username:  username,
		Email:     email,
		AvatarURL: avatarURL,
		CreatedAt: time.Now(),
	}
	return &user, nil
}

func (s *UserStore) GetByGitHubID(githubID string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE github_id = ?`
	err := s.db.Get(&user, query, githubID)
	return &user, err
}

func (s *UserStore) GetByID(userID string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE id = ?`
	err := s.db.Get(&user, query, userID)
	return &user, err
}
