package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/user-service/internal/domain"
)

type FollowRepo struct {
	db *pgxpool.Pool
}

func NewFollowRepo(db *pgxpool.Pool) *FollowRepo {
	return &FollowRepo{db: db}
}

func (r *FollowRepo) Add(ctx context.Context, f *domain.Follow) error {
	const q = `
INSERT INTO user_follows (follower_id, followee_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, q, f.FollowerID, f.FolloweeID, f.CreatedAt)
	return err
}

func (r *FollowRepo) Remove(ctx context.Context, followerID, followeeID string) error {
	const q = `DELETE FROM user_follows WHERE follower_id = $1 AND followee_id = $2`
	_, err := r.db.Exec(ctx, q, followerID, followeeID)
	return err
}

func (r *FollowRepo) ListFollowers(ctx context.Context, userID string, limit, offset int) ([]*domain.User, error) {
	const q = `
SELECT u.id, u.email, u.display_name, u.avatar_url, u.country, u.language, u.tier, u.created_at, u.updated_at
FROM user_follows f
JOIN users u ON u.id = f.follower_id
WHERE f.followee_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3`
	return r.scanUsers(ctx, q, userID, limit, offset)
}

func (r *FollowRepo) ListFollowing(ctx context.Context, userID string, limit, offset int) ([]*domain.User, error) {
	const q = `
SELECT u.id, u.email, u.display_name, u.avatar_url, u.country, u.language, u.tier, u.created_at, u.updated_at
FROM user_follows f
JOIN users u ON u.id = f.followee_id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3`
	return r.scanUsers(ctx, q, userID, limit, offset)
}

func (r *FollowRepo) scanUsers(ctx context.Context, q, userID string, limit, offset int) ([]*domain.User, error) {
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.User, 0)
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(
			&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Country, &u.Language, &u.Tier, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
