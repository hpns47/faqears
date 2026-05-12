package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/playlist-service/internal/domain"
)

type TrackRepo struct {
	db *pgxpool.Pool
}

func NewTrackRepo(db *pgxpool.Pool) *TrackRepo {
	return &TrackRepo{db: db}
}

func (r *TrackRepo) AddTrack(ctx context.Context, t *domain.PlaylistTrack) error {
	const q = `
INSERT INTO playlist_tracks (playlist_id, track_id, rank, added_by, added_at)
VALUES ($1, $2, $3::numeric, $4, $5)`
	_, err := r.db.Exec(ctx, q, t.PlaylistID, t.TrackID, t.Rank, t.AddedBy, t.AddedAt)
	return err
}

func (r *TrackRepo) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	const q = `DELETE FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`
	_, err := r.db.Exec(ctx, q, playlistID, trackID)
	return err
}

func (r *TrackRepo) GetTrack(ctx context.Context, playlistID, trackID string) (*domain.PlaylistTrack, error) {
	const q = `
SELECT playlist_id, track_id, rank::text, added_by, added_at
FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`
	t := &domain.PlaylistTrack{}
	err := r.db.QueryRow(ctx, q, playlistID, trackID).Scan(
		&t.PlaylistID, &t.TrackID, &t.Rank, &t.AddedBy, &t.AddedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TrackRepo) ListTracks(ctx context.Context, playlistID string) ([]*domain.PlaylistTrack, error) {
	const q = `
SELECT playlist_id, track_id, rank::text, added_by, added_at
FROM playlist_tracks WHERE playlist_id = $1
ORDER BY rank ASC`
	rows, err := r.db.Query(ctx, q, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.PlaylistTrack
	for rows.Next() {
		t := &domain.PlaylistTrack{}
		if err := rows.Scan(&t.PlaylistID, &t.TrackID, &t.Rank, &t.AddedBy, &t.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TrackRepo) GetMaxRank(ctx context.Context, playlistID string) (string, error) {
	const q = `SELECT COALESCE(MAX(rank)::text, '') FROM playlist_tracks WHERE playlist_id = $1`
	var rank string
	err := r.db.QueryRow(ctx, q, playlistID).Scan(&rank)
	return rank, err
}

func (r *TrackRepo) GetMinRank(ctx context.Context, playlistID string) (string, error) {
	const q = `SELECT COALESCE(MIN(rank)::text, '') FROM playlist_tracks WHERE playlist_id = $1`
	var rank string
	err := r.db.QueryRow(ctx, q, playlistID).Scan(&rank)
	return rank, err
}

func (r *TrackRepo) GetRankAfterTrack(ctx context.Context, playlistID, trackID string) (string, bool, error) {
	const q = `
SELECT rank::text
FROM playlist_tracks
WHERE playlist_id = $1
  AND rank > (SELECT rank FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2)
ORDER BY rank ASC
LIMIT 1`
	var rank string
	err := r.db.QueryRow(ctx, q, playlistID, trackID).Scan(&rank)
	if err != nil {
		return "", false, nil
	}
	return rank, true, nil
}

func (r *TrackRepo) UpdateRank(ctx context.Context, playlistID, trackID, rank string) error {
	const q = `UPDATE playlist_tracks SET rank = $1::numeric WHERE playlist_id = $2 AND track_id = $3`
	_, err := r.db.Exec(ctx, q, rank, playlistID, trackID)
	return err
}

func (r *TrackRepo) ListPlaylistsForTrack(ctx context.Context, trackID string) ([]string, error) {
	const q = `SELECT playlist_id FROM playlist_tracks WHERE track_id = $1`
	rows, err := r.db.Query(ctx, q, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
