package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/auth-service/internal/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
INSERT INTO auth_users (id, email, password_hash, roles, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, q, u.ID, u.Email, u.PasswordHash, u.Roles, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, roles, created_at, updated_at
FROM auth_users WHERE email = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Roles, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, roles, created_at, updated_at
FROM auth_users WHERE id = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Roles, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	const q = `UPDATE auth_users SET password_hash = $2, updated_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id, passwordHash)
	return err
}
