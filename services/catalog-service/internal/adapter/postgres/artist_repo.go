package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/catalog-service/internal/domain"
)

type ArtistRepo struct {
	db *pgxpool.Pool
}

func NewArtistRepo(db *pgxpool.Pool) *ArtistRepo {
	return &ArtistRepo{db: db}
}

func (r *ArtistRepo) Create(ctx context.Context, a *domain.Artist) error {
	const q = `
INSERT INTO artists (id, name, country, biography, created_at)
VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, q, a.ID, a.Name, a.Country, a.Biography, a.CreatedAt)
	return err
}

func (r *ArtistRepo) GetByID(ctx context.Context, id string) (*domain.Artist, error) {
	const q = `
SELECT id, name, country, biography, created_at
FROM artists WHERE id = $1`
	a := &domain.Artist{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.Name, &a.Country, &a.Biography, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *ArtistRepo) Search(ctx context.Context, query string, limit int32) ([]*domain.Artist, error) {
	const q = `
SELECT id, name, country, biography, created_at
FROM artists WHERE name ILIKE '%' || $1 || '%'
ORDER BY name
LIMIT $2`
	rows, err := r.db.Query(ctx, q, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var artists []*domain.Artist
	for rows.Next() {
		a := &domain.Artist{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Country, &a.Biography, &a.CreatedAt); err != nil {
			return nil, err
		}
		artists = append(artists, a)
	}
	return artists, rows.Err()
}
