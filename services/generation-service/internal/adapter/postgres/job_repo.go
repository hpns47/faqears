package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/generation-service/internal/domain"
)

type JobRepo struct {
	db *pgxpool.Pool
}

func NewJobRepo(db *pgxpool.Pool) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(ctx context.Context, j *domain.Job) error {
	const q = `
INSERT INTO generation_jobs
  (id, user_id, status, prompt, title, genre, mood, language, lyrics, track_id, failure_reason, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	_, err := r.db.Exec(ctx, q,
		j.ID, j.UserID, j.Status, j.Prompt, j.Title, j.Genre, j.Mood, j.Language,
		j.Lyrics, j.TrackID, j.FailureReason, j.CreatedAt, j.UpdatedAt,
	)
	return err
}

func (r *JobRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	const q = `
SELECT id, user_id, status, prompt, title, genre, mood, language, lyrics, track_id, failure_reason, created_at, updated_at
FROM generation_jobs WHERE id = $1`
	j := &domain.Job{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&j.ID, &j.UserID, &j.Status, &j.Prompt, &j.Title, &j.Genre, &j.Mood, &j.Language,
		&j.Lyrics, &j.TrackID, &j.FailureReason, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (r *JobRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Job, error) {
	const q = `
SELECT id, user_id, status, prompt, title, genre, mood, language, lyrics, track_id, failure_reason, created_at, updated_at
FROM generation_jobs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []*domain.Job
	for rows.Next() {
		j := &domain.Job{}
		if err := rows.Scan(
			&j.ID, &j.UserID, &j.Status, &j.Prompt, &j.Title, &j.Genre, &j.Mood, &j.Language,
			&j.Lyrics, &j.TrackID, &j.FailureReason, &j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *JobRepo) UpdateStatus(ctx context.Context, id, status string) error {
	const q = `UPDATE generation_jobs SET status=$1, updated_at=$2 WHERE id=$3`
	_, err := r.db.Exec(ctx, q, status, time.Now().UTC(), id)
	return err
}

func (r *JobRepo) UpdateStatusField(ctx context.Context, id, status string) error {
	return r.UpdateStatus(ctx, id, status)
}

func (r *JobRepo) UpdateStatusAndLyrics(ctx context.Context, id, status, lyrics string) error {
	const q = `UPDATE generation_jobs SET status=$1, lyrics=$2, updated_at=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, q, status, lyrics, time.Now().UTC(), id)
	return err
}

func (r *JobRepo) UpdateStatusAndFailure(ctx context.Context, id, status, reason string) error {
	const q = `UPDATE generation_jobs SET status=$1, failure_reason=$2, updated_at=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, q, status, reason, time.Now().UTC(), id)
	return err
}

func (r *JobRepo) UpdatePublished(ctx context.Context, id, trackID string) error {
	const q = `UPDATE generation_jobs SET status=$1, track_id=$2, updated_at=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, q, domain.StatusPublished, trackID, time.Now().UTC(), id)
	return err
}
