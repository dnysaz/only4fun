package database

import (
	"log"
	"time"

	"only4.fun/backend/internal/id"
	"only4.fun/backend/internal/secure"

	"github.com/jmoiron/sqlx"
)

type SessionStore struct {
	db *sqlx.DB
}

func NewSessionStore(db *Database) *SessionStore {
	return &SessionStore{db: db.DB()}
}

func (s *SessionStore) Create(userID, sessionID, token string, expiresAt time.Time) (*Session, error) {
	// Buang session yang sudah kedaluwarsa agar tabel tidak menumpuk.
	// datetime(expires_at) dipakai karena driver menyimpan time.Time lengkap
	// dengan offset lokal (mis. +08:00); membandingkan string mentah dengan
	// datetime('now') (UTC) membuat masa berlaku meleset sebesar offset zona waktu.
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE datetime(expires_at) <= datetime('now')`)

	// ID session dan token GitHub tidak pernah disimpan dalam bentuk asli.
	encToken, err := secure.Encrypt(token)
	if err != nil {
		return nil, err
	}

	newID := id.New()
	query := `INSERT INTO sessions (id, user_id, session_id, token, expires_at) VALUES (?, ?, ?, ?, ?)`
	if _, err := s.db.Exec(query, newID, userID, secure.HashID(sessionID), encToken, expiresAt); err != nil {
		return nil, err
	}
	session := Session{
		ID:        newID,
		UserID:    userID,
		SessionID: secure.HashID(sessionID),
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return &session, nil
}

// GetBySessionID mencari session memakai hash dari ID yang dikirim klien.
func (s *SessionStore) GetBySessionID(sessionID string) (*Session, error) {
	var session Session
	query := `SELECT * FROM sessions WHERE session_id = ? AND datetime(expires_at) > datetime('now')`
	err := s.db.Get(&session, query, secure.HashID(sessionID))
	if err != nil {
		return &session, err
	}
	return &session, s.decryptToken(&session)
}

// decryptToken mengisi Session.Token dengan nilai asli hasil dekripsi.
func (s *SessionStore) decryptToken(session *Session) error {
	token, err := secure.Decrypt(session.Token)
	if err != nil {
		log.Printf("failed to decrypt the session token %s: %v", session.ID, err)
		session.Token = ""
		return nil
	}
	session.Token = token
	return nil
}

// GetLatestByUserID mengambil session aktif terbaru milik user, dipakai untuk
// mendapatkan GitHub token saat clone repo privat.
func (s *SessionStore) GetLatestByUserID(userID string) (*Session, error) {
	var session Session
	query := `SELECT * FROM sessions WHERE user_id = ? AND datetime(expires_at) > datetime('now') ORDER BY id DESC LIMIT 1`
	err := s.db.Get(&session, query, userID)
	if err != nil {
		return &session, err
	}
	return &session, s.decryptToken(&session)
}

func (s *SessionStore) Delete(sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = ?`
	_, err := s.db.Exec(query, secure.HashID(sessionID))
	return err
}
