package database

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Init menginisialisasi koneksi database SQLite.
func Init(dbPath string) (*Database, error) {
	// Set lewat DSN supaya setiap koneksi (termasuk yang dibuat ulang pool)
	// tetap patuh FK dan memakai WAL. WAL + synchronous=NORMAL mengurangi
	// fsync per commit, sehingga deploy history/build log lebih ringan.
	dsn := dbPath
	params := strings.Join([]string{
		"_foreign_keys=on",
		"_journal_mode=WAL",
		"_busy_timeout=5000",
		"_synchronous=NORMAL",
	}, "&")
	if strings.Contains(dsn, "?") {
		dsn += "&" + params
	} else {
		dsn += "?" + params
	}

	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	// Pastikan juga untuk driver/versi yang mengabaikan flag DSN.
	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
	} {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to apply %s: %w", pragma, err)
		}
	}

	// Atur pooling dan keamanan
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return New(db), nil
}
