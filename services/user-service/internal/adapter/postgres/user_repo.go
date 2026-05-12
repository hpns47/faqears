package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/user-service/internal/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
INSERT INTO users (id, email, display_name, avatar_url, country, language, tier, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	tier := u.Tier
	if tier == "" {
		tier = domain.TierFree
	}
	_, err := r.db.Exec(ctx, q,
		u.ID, u.Email, u.DisplayName, u.AvatarURL, u.Country, u.Language, tier, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
SELECT id, email, display_name, avatar_url, country, language, tier, created_at, updated_at
FROM users WHERE id = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Country, &u.Language, &u.Tier, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id string, p domain.Profile) (*domain.User, error) {
	const q = `
UPDATE users
SET display_name = $2,
    avatar_url   = $3,
    country      = $4,
    language     = $5,
    updated_at   = $6
WHERE id = $1
RETURNING id, email, display_name, avatar_url, country, language, tier, created_at, updated_at`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id, p.DisplayName, p.AvatarURL, p.Country, p.Language, time.Now().UTC()).Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Country, &u.Language, &u.Tier, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) UpdateTier(ctx context.Context, id, tier string) (*domain.User, error) {
	const q = `
UPDATE users
SET tier       = $2,
    updated_at = $3
WHERE id = $1
RETURNING id, email, display_name, avatar_url, country, language, tier, created_at, updated_at`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id, tier, time.Now().UTC()).Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Country, &u.Language, &u.Tier, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}
