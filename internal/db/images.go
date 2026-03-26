package db

import (
	"database/sql"
	"time"

	"xz.ar/internal/model"
)

func (s *Store) ListCarouselImages() ([]model.CarouselImage, error) {
	rows, err := s.db.Query(`SELECT id, filename, original, mime_type, sort_order, created_at FROM carousel_images ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.CarouselImage
	for rows.Next() {
		var img model.CarouselImage
		var createdAt string
		if err := rows.Scan(&img.ID, &img.Filename, &img.Original, &img.MimeType, &img.SortOrder, &createdAt); err != nil {
			return nil, err
		}
		img.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, img)
	}
	return out, rows.Err()
}

func (s *Store) CreateCarouselImage(filename, original, mimeType string) error {
	var maxOrder int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) FROM carousel_images`).Scan(&maxOrder)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO carousel_images (filename, original, mime_type, sort_order) VALUES (?, ?, ?, ?)`,
		filename, original, mimeType, maxOrder+1)
	return err
}

func (s *Store) DeleteCarouselImage(id int64) (*model.CarouselImage, error) {
	var img model.CarouselImage
	var createdAt string
	err := s.db.QueryRow(`SELECT id, filename, original, mime_type, sort_order, created_at FROM carousel_images WHERE id = ?`, id).
		Scan(&img.ID, &img.Filename, &img.Original, &img.MimeType, &img.SortOrder, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	img.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	_, err = s.db.Exec(`DELETE FROM carousel_images WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (s *Store) ReorderCarouselImages(ids []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if _, err := tx.Exec(`UPDATE carousel_images SET sort_order = ? WHERE id = ?`, i, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
