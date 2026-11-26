package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Init menginisialisasi koneksi database SQLite.
func Init(dbPath string) (*Database, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	// Atur pooling dan keamanan
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return New(db), nil
}
