package port

import (
	"context"

	"github.com/faqears/faqears/services/payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Payment, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateFromWebhook(ctx context.Context, id string, p *domain.Payment) error
}

type EventRepository interface {
	Insert(ctx context.Context, e *domain.Event) error
}

type Provider interface {
	CreateInvoice(ctx context.Context, p *domain.Payment) (providerPaymentID, invoiceURL, payAddress, payAmount string, err error)
	VerifySignature(body []byte, sig string) error
}

type EventPublisher interface {
	PaymentInvoiceCreated(ctx context.Context, p *domain.Payment) error
	PaymentStatusChanged(ctx context.Context, p *domain.Payment) error
	PaymentFinished(ctx context.Context, p *domain.Payment) error
	PaymentFailed(ctx context.Context, p *domain.Payment) error
}
