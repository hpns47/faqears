package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/catalog-service/internal/domain"
)

type AlbumRepo struct {
	db *pgxpool.Pool
}

func NewAlbumRepo(db *pgxpool.Pool) *AlbumRepo {
	return &AlbumRepo{db: db}
}

func (r *AlbumRepo) Create(ctx context.Context, a *domain.Album) error {
	const q = `
INSERT INTO albums (id, artist_id, title, year, cover_url, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, q, a.ID, a.ArtistID, a.Title, a.Year, a.CoverURL, a.CreatedAt)
	return err
}

func (r *AlbumRepo) GetByID(ctx context.Context, id string) (*domain.Album, error) {
	const q = `
SELECT id, artist_id, title, year, cover_url, created_at
FROM albums WHERE id = $1`
	a := &domain.Album{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.ArtistID, &a.Title, &a.Year, &a.CoverURL, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AlbumRepo) ListByArtist(ctx context.Context, artistID string, limit, offset int32) ([]*domain.Album, error) {
	const q = `
SELECT id, artist_id, title, year, cover_url, created_at
FROM albums WHERE artist_id = $1
ORDER BY year DESC, title
LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, artistID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var albums []*domain.Album
	for rows.Next() {
		a := &domain.Album{}
		if err := rows.Scan(&a.ID, &a.ArtistID, &a.Title, &a.Year, &a.CoverURL, &a.CreatedAt); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}
	return albums, rows.Err()
}

func (r *AlbumRepo) Search(ctx context.Context, query string, limit int32) ([]*domain.Album, error) {
	const q = `
SELECT id, artist_id, title, year, cover_url, created_at
FROM albums WHERE title ILIKE '%' || $1 || '%'
ORDER BY title
LIMIT $2`
	rows, err := r.db.Query(ctx, q, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var albums []*domain.Album
	for rows.Next() {
		a := &domain.Album{}
		if err := rows.Scan(&a.ID, &a.ArtistID, &a.Title, &a.Year, &a.CoverURL, &a.CreatedAt); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}
	return albums, rows.Err()
}
