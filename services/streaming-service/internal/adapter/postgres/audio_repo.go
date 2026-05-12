package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/streaming-service/internal/domain"
)

type AudioRepo struct {
	db *pgxpool.Pool
}

func NewAudioRepo(db *pgxpool.Pool) *AudioRepo {
	return &AudioRepo{db: db}
}

func (r *AudioRepo) Upsert(ctx context.Context, a *domain.TrackAudio) error {
	const q = `
INSERT INTO track_audio (track_id, object_key, content_type, size_bytes, uploaded_by, uploaded_at)
VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, $6)
ON CONFLICT (track_id) DO UPDATE
  SET object_key   = EXCLUDED.object_key,
      content_type = EXCLUDED.content_type,
      size_bytes   = EXCLUDED.size_bytes,
      uploaded_by  = EXCLUDED.uploaded_by,
      uploaded_at  = EXCLUDED.uploaded_at`
	_, err := r.db.Exec(ctx, q,
		a.TrackID, a.ObjectKey, a.ContentType, a.SizeBytes, a.UploadedBy, a.UploadedAt,
	)
	return err
}

func (r *AudioRepo) GetByTrackID(ctx context.Context, trackID string) (*domain.TrackAudio, error) {
	const q = `
SELECT track_id, object_key, content_type, size_bytes,
       COALESCE(uploaded_by::text, ''), uploaded_at
FROM track_audio WHERE track_id = $1`
	a := &domain.TrackAudio{}
	err := r.db.QueryRow(ctx, q, trackID).Scan(
		&a.TrackID, &a.ObjectKey, &a.ContentType, &a.SizeBytes,
		&a.UploadedBy, &a.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (r *AudioRepo) Exists(ctx context.Context, trackID string) (bool, int64, error) {
	const q = `SELECT size_bytes FROM track_audio WHERE track_id = $1`
	var size int64
	err := r.db.QueryRow(ctx, q, trackID).Scan(&size)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, 0, nil
		}
		return false, 0, err
	}
	return true, size, nil
}
