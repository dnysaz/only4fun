package database

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Init menginisialisasi koneksi database SQLite.
func Init(dbPath string) (*Database, error) {
	// SQLite menonaktifkan foreign key secara default. Set lewat DSN supaya
	// setiap koneksi (termasuk yang dibuat ulang pool) tetap patuh FK.
	dsn := dbPath
	if strings.Contains(dsn, "?") {
		dsn += "&_foreign_keys=on"
	} else {
		dsn += "?_foreign_keys=on"
	}

	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	// Pastikan juga untuk driver/versi yang mengabaikan flag DSN.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Atur pooling dan keamanan
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return New(db), nil
}
