package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/catalog-service/internal/domain"
)

type TrackRepo struct {
	db *pgxpool.Pool
}

func NewTrackRepo(db *pgxpool.Pool) *TrackRepo {
	return &TrackRepo{db: db}
}

func (r *TrackRepo) Create(ctx context.Context, t *domain.Track) error {
	const q = `
INSERT INTO tracks (id, album_id, artist_id, title, duration_sec, isrc, genres, user_generated, owner_user_id, created_at)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, NULLIF($9, '')::uuid, $10)`
	_, err := r.db.Exec(ctx, q,
		t.ID, t.AlbumID, t.ArtistID, t.Title, t.DurationSec,
		t.ISRC, t.Genres, t.UserGenerated, t.OwnerUserID, t.CreatedAt,
	)
	return err
}

func (r *TrackRepo) GetByID(ctx context.Context, id string) (*domain.Track, error) {
	const q = `
SELECT id, album_id, artist_id, title, duration_sec,
       COALESCE(isrc, ''), genres, user_generated, COALESCE(owner_user_id::text, ''), created_at
FROM tracks WHERE id = $1`
	t := &domain.Track{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.AlbumID, &t.ArtistID, &t.Title, &t.DurationSec,
		&t.ISRC, &t.Genres, &t.UserGenerated, &t.OwnerUserID, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TrackRepo) ListByAlbum(ctx context.Context, albumID string, limit, offset int32) ([]*domain.Track, error) {
	const q = `
SELECT id, album_id, artist_id, title, duration_sec,
       COALESCE(isrc, ''), genres, user_generated, COALESCE(owner_user_id::text, ''), created_at
FROM tracks WHERE album_id = $1
ORDER BY title
LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, albumID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tracks []*domain.Track
	for rows.Next() {
		t := &domain.Track{}
		if err := rows.Scan(
			&t.ID, &t.AlbumID, &t.ArtistID, &t.Title, &t.DurationSec,
			&t.ISRC, &t.Genres, &t.UserGenerated, &t.OwnerUserID, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (r *TrackRepo) Search(ctx context.Context, query string, limit int32) ([]*domain.Track, error) {
	const q = `
SELECT id, album_id, artist_id, title, duration_sec,
       COALESCE(isrc, ''), genres, user_generated, COALESCE(owner_user_id::text, ''), created_at
FROM tracks WHERE title ILIKE '%' || $1 || '%'
ORDER BY title
LIMIT $2`
	rows, err := r.db.Query(ctx, q, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tracks []*domain.Track
	for rows.Next() {
		t := &domain.Track{}
		if err := rows.Scan(
			&t.ID, &t.AlbumID, &t.ArtistID, &t.Title, &t.DurationSec,
			&t.ISRC, &t.Genres, &t.UserGenerated, &t.OwnerUserID, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}
