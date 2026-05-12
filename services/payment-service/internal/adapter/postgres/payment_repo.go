package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/payment-service/internal/domain"
)

type PaymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepo(db *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	const q = `
INSERT INTO payments (id, user_id, provider, provider_payment_id, status, price_currency, price_amount, pay_currency, pay_address, pay_amount, purpose, invoice_url, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err := r.db.Exec(ctx, q,
		p.ID, p.UserID, p.Provider, p.ProviderPaymentID, p.Status,
		p.PriceCurrency, p.PriceAmount, p.PayCurrency, p.PayAddress, p.PayAmount,
		p.Purpose, p.InvoiceURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	const q = `
SELECT id, user_id, provider, provider_payment_id, status, price_currency, price_amount, pay_currency, pay_address, pay_amount, purpose, invoice_url, created_at, updated_at
FROM payments WHERE id = $1`
	p := &domain.Payment{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.UserID, &p.Provider, &p.ProviderPaymentID, &p.Status,
		&p.PriceCurrency, &p.PriceAmount, &p.PayCurrency, &p.PayAddress, &p.PayAmount,
		&p.Purpose, &p.InvoiceURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("payment not found")
		}
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Payment, error) {
	const q = `
SELECT id, user_id, provider, provider_payment_id, status, price_currency, price_amount, pay_currency, pay_address, pay_amount, purpose, invoice_url, created_at, updated_at
FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Provider, &p.ProviderPaymentID, &p.Status,
			&p.PriceCurrency, &p.PriceAmount, &p.PayCurrency, &p.PayAddress, &p.PayAmount,
			&p.Purpose, &p.InvoiceURL, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, id, status string) error {
	const q = `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, q, status, time.Now().UTC(), id)
	return err
}

func (r *PaymentRepo) UpdateFromWebhook(ctx context.Context, id string, p *domain.Payment) error {
	const q = `
UPDATE payments SET provider_payment_id=$1, invoice_url=$2, pay_address=$3, pay_amount=$4, status=$5, updated_at=$6
WHERE id=$7`
	_, err := r.db.Exec(ctx, q, p.ProviderPaymentID, p.InvoiceURL, p.PayAddress, p.PayAmount, p.Status, p.UpdatedAt, id)
	return err
}
