package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SetMigrations must be called before New to set the SQL migration script.
func SetMigrations(sql string) {
	migrationsSQL = sql
}

var migrationsSQL string

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	if migrationsSQL == "" {
		return fmt.Errorf("no migrations set; call db.SetMigrations first")
	}
	_, err := s.db.Exec(migrationsSQL)
	return err
}
