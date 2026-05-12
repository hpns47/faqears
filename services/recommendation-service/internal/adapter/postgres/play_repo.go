package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
)

type PlayRepo struct {
	db *pgxpool.Pool
}

func NewPlayRepo(db *pgxpool.Pool) *PlayRepo {
	return &PlayRepo{db: db}
}

func (r *PlayRepo) Save(ctx context.Context, e *domain.PlayEvent) error {
	const q = `
INSERT INTO play_events (user_id, track_id, session_id, event_type, occurred_at)
VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5)`
	_, err := r.db.Exec(ctx, q, e.UserID, e.TrackID, e.SessionID, e.EventType, e.OccurredAt)
	return err
}

func (r *PlayRepo) TopPlayedByUser(ctx context.Context, userID string, limit int, days int) ([]string, error) {
	const q = `
SELECT track_id
FROM play_events
WHERE user_id = $1
  AND occurred_at > NOW() - ($2 || ' days')::interval
GROUP BY track_id
ORDER BY COUNT(*) DESC
LIMIT $3`
	rows, err := r.db.Query(ctx, q, userID, days, limit)
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

func (r *PlayRepo) ActiveTrackIDs(ctx context.Context, days int) ([]string, error) {
	const q = `
SELECT DISTINCT track_id
FROM play_events
WHERE occurred_at > NOW() - ($1 || ' days')::interval`
	rows, err := r.db.Query(ctx, q, days)
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

func (r *PlayRepo) TopGlobal(ctx context.Context, limit int) ([]*domain.TrackRef, error) {
	const q = `
SELECT track_id, COUNT(*) AS play_count
FROM play_events
GROUP BY track_id
ORDER BY play_count DESC
LIMIT $1`
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []*domain.TrackRef
	for rows.Next() {
		var ref domain.TrackRef
		var count int64
		if err := rows.Scan(&ref.TrackID, &count); err != nil {
			return nil, err
		}
		ref.Score = float64(count)
		refs = append(refs, &ref)
	}
	return refs, rows.Err()
}

func (r *PlayRepo) PlayedByUser(ctx context.Context, userID string, days int) (map[string]struct{}, error) {
	const q = `
SELECT DISTINCT track_id
FROM play_events
WHERE user_id = $1
  AND occurred_at > NOW() - ($2 || ' days')::interval`
	rows, err := r.db.Query(ctx, q, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = struct{}{}
	}
	return result, rows.Err()
}
