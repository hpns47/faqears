package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/payment-service/internal/domain"
)

type EventRepo struct {
	db *pgxpool.Pool
}

func NewEventRepo(db *pgxpool.Pool) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Insert(ctx context.Context, e *domain.Event) error {
	const q = `
INSERT INTO payments_events (payment_id, status, occurred_at, raw_payload)
VALUES ($1, $2, $3, $4)
ON CONFLICT (payment_id, status, occurred_at) DO NOTHING`
	_, err := r.db.Exec(ctx, q, e.PaymentID, e.Status, e.OccurredAt, e.RawPayload)
	return err
}
