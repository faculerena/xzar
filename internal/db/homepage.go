package db

import (
	"time"

	"xz.ar/internal/model"
)

func (s *Store) GetHomepageConfig() (*model.HomepageConfig, error) {
	var cfg model.HomepageConfig
	var updatedAt string
	err := s.db.QueryRow(`SELECT mode, redirect_url, updated_at FROM homepage_config WHERE id = 1`).
		Scan(&cfg.Mode, &cfg.RedirectURL, &updatedAt)
	if err != nil {
		return nil, err
	}
	cfg.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &cfg, nil
}

func (s *Store) UpdateHomepageConfig(mode model.HomepageMode, redirectURL string) error {
	_, err := s.db.Exec(`UPDATE homepage_config SET mode = ?, redirect_url = ?, updated_at = ? WHERE id = 1`,
		mode, redirectURL, time.Now().UTC().Format(time.RFC3339))
	return err
}
