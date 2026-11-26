package database

import (
	"github.com/jmoiron/sqlx"
)

type Database struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Database {
	return &Database{db: db}
}

func (db *Database) DB() *sqlx.DB {
	return db.db
}

func (db *Database) Close() error {
	return db.db.Close()
}
