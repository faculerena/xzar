package db

import (
	"database/sql"
	"fmt"
	"time"

	"xz.ar/internal/model"
)

func (s *Store) ListShortcuts() ([]model.Shortcut, error) {
	rows, err := s.db.Query(`SELECT id, slug, target_url, type, click_count, created_at, updated_at FROM shortcuts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Shortcut
	for rows.Next() {
		sc, err := scanShortcut(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Store) GetShortcutBySlug(typ model.ShortcutType, slug string) (*model.Shortcut, error) {
	row := s.db.QueryRow(`SELECT id, slug, target_url, type, click_count, created_at, updated_at FROM shortcuts WHERE type = ? AND slug = ?`, typ, slug)
	sc, err := scanShortcutRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (s *Store) GetShortcutByID(id int64) (*model.Shortcut, error) {
	row := s.db.QueryRow(`SELECT id, slug, target_url, type, click_count, created_at, updated_at FROM shortcuts WHERE id = ?`, id)
	sc, err := scanShortcutRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (s *Store) CreateShortcut(slug, targetURL string, typ model.ShortcutType) error {
	_, err := s.db.Exec(`INSERT INTO shortcuts (slug, target_url, type) VALUES (?, ?, ?)`, slug, targetURL, typ)
	if err != nil {
		return fmt.Errorf("create shortcut: %w", err)
	}
	return nil
}

func (s *Store) UpdateShortcut(id int64, slug, targetURL string, typ model.ShortcutType) error {
	_, err := s.db.Exec(`UPDATE shortcuts SET slug = ?, target_url = ?, type = ?, updated_at = ? WHERE id = ?`,
		slug, targetURL, typ, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func (s *Store) DeleteShortcut(id int64) error {
	_, err := s.db.Exec(`DELETE FROM shortcuts WHERE id = ?`, id)
	return err
}

func (s *Store) IncrementClickCount(id int64) error {
	_, err := s.db.Exec(`UPDATE shortcuts SET click_count = click_count + 1 WHERE id = ?`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanShortcutFromScanner(s scanner) (model.Shortcut, error) {
	var sc model.Shortcut
	var createdAt, updatedAt string
	err := s.Scan(&sc.ID, &sc.Slug, &sc.TargetURL, &sc.Type, &sc.ClickCount, &createdAt, &updatedAt)
	if err != nil {
		return sc, err
	}
	sc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	sc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return sc, nil
}

func scanShortcut(rows *sql.Rows) (model.Shortcut, error) {
	return scanShortcutFromScanner(rows)
}

func scanShortcutRow(row *sql.Row) (model.Shortcut, error) {
	return scanShortcutFromScanner(row)
}
