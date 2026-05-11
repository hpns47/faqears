package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/auth-service/internal/domain"
)

type TokenRepo struct {
	db *pgxpool.Pool
}

func NewTokenRepo(db *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Save(ctx context.Context, t *domain.RefreshToken) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, q, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.CreatedAt, t.Revoked)
	return err
}

func (r *TokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const q = `
SELECT id, user_id, token_hash, expires_at, created_at, revoked
FROM refresh_tokens WHERE token_hash = $1`
	t := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.Revoked,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TokenRepo) Revoke(ctx context.Context, id string) error {
	const q = `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}
