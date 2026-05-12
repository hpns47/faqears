package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/generation-service/internal/domain"
)

type LyricsRepo struct {
	db *pgxpool.Pool
}

func NewLyricsRepo(db *pgxpool.Pool) *LyricsRepo {
	return &LyricsRepo{db: db}
}

func (r *LyricsRepo) Save(ctx context.Context, lv *domain.LyricsVersion) error {
	const q = `
INSERT INTO lyrics_versions (id, job_id, revision, body, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (job_id, revision) DO UPDATE SET body = EXCLUDED.body`
	_, err := r.db.Exec(ctx, q, lv.ID, lv.JobID, lv.Revision, lv.Body, lv.CreatedAt)
	return err
}

func (r *LyricsRepo) ListByJob(ctx context.Context, jobID string) ([]*domain.LyricsVersion, error) {
	const q = `
SELECT id, job_id, revision, body, created_at
FROM lyrics_versions WHERE job_id = $1 ORDER BY revision ASC`
	rows, err := r.db.Query(ctx, q, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.LyricsVersion
	for rows.Next() {
		lv := &domain.LyricsVersion{}
		if err := rows.Scan(&lv.ID, &lv.JobID, &lv.Revision, &lv.Body, &lv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, lv)
	}
	return out, rows.Err()
}

func (r *LyricsRepo) LatestRevision(ctx context.Context, jobID string) (int32, error) {
	const q = `SELECT COALESCE(MAX(revision), 0) FROM lyrics_versions WHERE job_id = $1`
	var rev int32
	err := r.db.QueryRow(ctx, q, jobID).Scan(&rev)
	return rev, err
}
