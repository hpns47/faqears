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
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, q, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.CreatedAt, t.Revoked, t.UserAgent)
	return err
}

func (r *TokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const q = `
SELECT id, user_id, token_hash, expires_at, created_at, revoked, user_agent
FROM refresh_tokens WHERE token_hash = $1`
	t := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.Revoked, &t.UserAgent,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TokenRepo) GetByID(ctx context.Context, id string) (*domain.RefreshToken, error) {
	const q = `
SELECT id, user_id, token_hash, expires_at, created_at, revoked, user_agent
FROM refresh_tokens WHERE id = $1`
	t := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.Revoked, &t.UserAgent,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TokenRepo) ListByUser(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	const q = `
SELECT id, user_id, token_hash, expires_at, created_at, revoked, user_agent
FROM refresh_tokens
WHERE user_id = $1
ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []*domain.RefreshToken
	for rows.Next() {
		t := &domain.RefreshToken{}
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.Revoked, &t.UserAgent,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func (r *TokenRepo) Revoke(ctx context.Context, id string) error {
	const q = `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}
