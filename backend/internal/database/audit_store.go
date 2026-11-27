package database

import (
	"strings"
	"time"

	"only4.fun/backend/internal/id"

	"github.com/jmoiron/sqlx"
)

// AuditLog mencatat aksi sensitif tanpa menyimpan nilai rahasia (env var,
// token, atau isi body). target hanya berisi id project/deployment/session.
type AuditLog struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Action    string    `db:"action" json:"action"`
	Target    string    `db:"target" json:"target"`
	IP        string    `db:"ip" json:"ip"`
	Status    int       `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type AuditStore struct {
	db *sqlx.DB
}

func NewAuditStore(db *Database) *AuditStore {
	return &AuditStore{db: db.DB()}
}

func clip(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

// Record menyimpan satu entri audit. userID kosong dipakai untuk aksi publik.
// Kegagalan menulis audit tidak boleh menggagalkan request utama, tetapi tetap
// dikembalikan sebagai error agar pemanggil bisa mencatatnya di log server.
func (s *AuditStore) Record(userID, action, target, ip string, status int) error {
	if s == nil || action == "" {
		return nil
	}

	var actor any
	if userID = clip(userID, 128); userID != "" {
		actor = userID
	}

	_, err := s.db.Exec(
		`INSERT INTO audit_logs (id, user_id, action, target, ip, status) VALUES (?, ?, ?, ?, ?, ?)`,
		id.New(),
		actor,
		clip(action, 80),
		clip(target, 256),
		clip(ip, 64),
		status,
	)
	return err
}

// ListByUser mengembalikan riwayat aksi termutakhir milik satu user.
func (s *AuditStore) ListByUser(userID string, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var logs []AuditLog
	err := s.db.Select(&logs,
		`SELECT id, coalesce(user_id, '') AS user_id, action, target, ip, status, created_at
		 FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`,
		userID,
		limit,
	)
	return logs, err
}
