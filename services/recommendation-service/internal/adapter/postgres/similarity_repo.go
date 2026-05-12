package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
)

type SimilarityRepo struct {
	db *pgxpool.Pool
}

func NewSimilarityRepo(db *pgxpool.Pool) *SimilarityRepo {
	return &SimilarityRepo{db: db}
}

func (r *SimilarityRepo) CoListeners(ctx context.Context, trackID string, limit int) ([]*domain.Similarity, error) {
	const q = `
SELECT b.track_id, COUNT(DISTINCT a.user_id) AS co_count
FROM play_events a
JOIN play_events b ON a.user_id = b.user_id AND a.track_id != b.track_id
WHERE a.track_id = $1
  AND a.occurred_at > NOW() - INTERVAL '30 days'
  AND b.occurred_at > NOW() - INTERVAL '30 days'
GROUP BY b.track_id
ORDER BY co_count DESC
LIMIT $2`
	rows, err := r.db.Query(ctx, q, trackID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sims []*domain.Similarity
	for rows.Next() {
		s := &domain.Similarity{}
		if err := rows.Scan(&s.TrackID, &s.CoCount); err != nil {
			return nil, err
		}
		sims = append(sims, s)
	}
	return sims, rows.Err()
}
