package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/services/payment-service/internal/domain"
	"github.com/faqears/faqears/services/payment-service/internal/port"
)

var validTransitions = map[string][]string{
	domain.StatusCreated:    {domain.StatusWaiting},
	domain.StatusWaiting:    {domain.StatusConfirming, domain.StatusRefunded, domain.StatusExpired, domain.StatusFailed},
	domain.StatusConfirming: {domain.StatusConfirmed, domain.StatusPartiallyPaid, domain.StatusFailed, domain.StatusExpired},
	domain.StatusConfirmed:  {domain.StatusSending},
	domain.StatusSending:    {domain.StatusFinished, domain.StatusFailed},
}

func transition(current, next string) error {
	if current == next {
		return nil
	}
	allowed, ok := validTransitions[current]
	if !ok {
		return domain.ErrStateMachine
	}
	for _, s := range allowed {
		if s == next {
			return nil
		}
	}
	return domain.ErrStateMachine
}

type Payment struct {
	repo      port.PaymentRepository
	events    port.EventRepository
	provider  port.Provider
	publisher port.EventPublisher
	log       *slog.Logger
}

func NewPayment(
	repo port.PaymentRepository,
	events port.EventRepository,
	provider port.Provider,
	publisher port.EventPublisher,
	log *slog.Logger,
) *Payment {
	return &Payment{
		repo:      repo,
		events:    events,
		provider:  provider,
		publisher: publisher,
		log:       log,
	}
}

func (uc *Payment) CreateInvoice(ctx context.Context, purpose, priceCurrency, priceAmount, payCurrency string) (*domain.Payment, error) {
	userID := grpcx.UserIDFromCtx(ctx)
	if userID == "" {
		return nil, errs.Unauthenticated("missing user identity")
	}

	p := &domain.Payment{
		ID:            uuid.NewString(),
		UserID:        userID,
		Provider:      domain.ProviderNowPayments,
		Status:        domain.StatusCreated,
		PriceCurrency: priceCurrency,
		PriceAmount:   priceAmount,
		PayCurrency:   payCurrency,
		Purpose:       purpose,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, errs.Internal("create payment", err)
	}

	providerID, invoiceURL, payAddress, payAmount, err := uc.provider.CreateInvoice(ctx, p)
	if err != nil {
		return nil, err
	}

	p.ProviderPaymentID = providerID
	p.InvoiceURL = invoiceURL
	p.PayAddress = payAddress
	p.PayAmount = payAmount
	p.Status = domain.StatusWaiting
	p.UpdatedAt = time.Now().UTC()

	if err := uc.repo.UpdateFromWebhook(ctx, p.ID, p); err != nil {
		return nil, errs.Internal("update payment after invoice", err)
	}

	if err := uc.publisher.PaymentInvoiceCreated(ctx, p); err != nil {
		uc.log.WarnContext(ctx, "publish invoice_created failed", slog.String("payment_id", p.ID), slog.String("error", err.Error()))
	}

	return p, nil
}

func (uc *Payment) GetPayment(ctx context.Context, paymentID string) (*domain.Payment, error) {
	userID := grpcx.UserIDFromCtx(ctx)
	if userID == "" {
		return nil, errs.Unauthenticated("missing user identity")
	}

	p, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, errs.NotFound("payment not found")
	}

	if p.UserID != userID {
		return nil, errs.PermissionDenied("payment belongs to another user")
	}

	return p, nil
}

func (uc *Payment) ListPaymentsByUser(ctx context.Context, limit, offset int) ([]*domain.Payment, error) {
	userID := grpcx.UserIDFromCtx(ctx)
	if userID == "" {
		return nil, errs.Unauthenticated("missing user identity")
	}

	return uc.repo.ListByUserID(ctx, userID, limit, offset)
}

type webhookPayload struct {
	PaymentID    string `json:"payment_id"`
	OrderID      string `json:"order_id"`
	PaymentStatus string `json:"payment_status"`
	ActuallyPaid string `json:"actually_paid"`
	PayAddress   string `json:"pay_address"`
	PayAmount    string `json:"pay_amount"`
}

func (uc *Payment) HandleWebhook(ctx context.Context, body []byte, sig string) error {
	if err := uc.provider.VerifySignature(body, sig); err != nil {
		return errs.Unauthenticated("invalid signature")
	}

	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return errs.InvalidArgument("invalid webhook payload")
	}

	paymentID := payload.OrderID
	if paymentID == "" {
		return errs.InvalidArgument("missing order_id")
	}

	p, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return errs.NotFound("payment not found")
	}

	evt := &domain.Event{
		PaymentID:  p.ID,
		Status:     payload.PaymentStatus,
		OccurredAt: time.Now().UTC(),
		RawPayload: body,
	}

	if err := uc.events.Insert(ctx, evt); err != nil {
		uc.log.WarnContext(ctx, "duplicate or failed event insert", slog.String("payment_id", p.ID), slog.String("error", err.Error()))
	}

	if terr := transition(p.Status, payload.PaymentStatus); terr != nil {
		uc.log.WarnContext(ctx, "invalid state transition", slog.String("from", p.Status), slog.String("to", payload.PaymentStatus))
	} else {
		p.Status = payload.PaymentStatus
		p.UpdatedAt = time.Now().UTC()
		if err := uc.repo.UpdateStatus(ctx, p.ID, p.Status); err != nil {
			return errs.Internal("update payment status", err)
		}
	}

	if err := uc.publisher.PaymentStatusChanged(ctx, p); err != nil {
		uc.log.WarnContext(ctx, "publish status_changed failed", slog.String("payment_id", p.ID), slog.String("error", err.Error()))
	}

	if p.Status == domain.StatusFinished {
		if err := uc.publisher.PaymentFinished(ctx, p); err != nil {
			uc.log.WarnContext(ctx, "publish finished failed", slog.String("payment_id", p.ID), slog.String("error", err.Error()))
		}
	}

	if p.Status == domain.StatusFailed {
		if err := uc.publisher.PaymentFailed(ctx, p); err != nil {
			uc.log.WarnContext(ctx, "publish failed event failed", slog.String("payment_id", p.ID), slog.String("error", err.Error()))
		}
	}

	return nil
}
