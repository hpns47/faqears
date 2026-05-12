package grpc

import (
	"context"

	paymentv1 "github.com/faqears/faqears/gen/go/payment/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/payment-service/internal/domain"
	"github.com/faqears/faqears/services/payment-service/internal/usecase"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc *usecase.Payment
}

func NewServer(uc *usecase.Payment) *Server {
	return &Server{uc: uc}
}

func (s *Server) CreateInvoice(ctx context.Context, req *paymentv1.CreateInvoiceRequest) (*paymentv1.CreateInvoiceResponse, error) {
	p, err := s.uc.CreateInvoice(ctx, req.GetPurpose(), req.GetPriceCurrency(), req.GetPriceAmount(), req.GetPayCurrency())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &paymentv1.CreateInvoiceResponse{
		Payment: toProto(p),
		PayUrl:  p.InvoiceURL,
	}, nil
}

func (s *Server) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	p, err := s.uc.GetPayment(ctx, req.GetPaymentId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &paymentv1.GetPaymentResponse{Payment: toProto(p)}, nil
}

func (s *Server) ListPaymentsByUser(ctx context.Context, req *paymentv1.ListPaymentsByUserRequest) (*paymentv1.ListPaymentsByUserResponse, error) {
	payments, err := s.uc.ListPaymentsByUser(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	out := make([]*paymentv1.Payment, 0, len(payments))
	for _, p := range payments {
		out = append(out, toProto(p))
	}
	return &paymentv1.ListPaymentsByUserResponse{Payments: out}, nil
}

func (s *Server) HandleWebhook(ctx context.Context, req *paymentv1.HandleWebhookRequest) (*paymentv1.HandleWebhookResponse, error) {
	if err := s.uc.HandleWebhook(ctx, req.GetRawBody(), req.GetSignature()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &paymentv1.HandleWebhookResponse{}, nil
}

func toProto(p *domain.Payment) *paymentv1.Payment {
	return &paymentv1.Payment{
		Id:                p.ID,
		UserId:            p.UserID,
		Provider:          p.Provider,
		ProviderPaymentId: p.ProviderPaymentID,
		Status:            p.Status,
		PriceCurrency:     p.PriceCurrency,
		PriceAmount:       p.PriceAmount,
		PayCurrency:       p.PayCurrency,
		PayAddress:        p.PayAddress,
		PayAmount:         p.PayAmount,
		Purpose:           p.Purpose,
		CreatedAt:         p.CreatedAt.Unix(),
		UpdatedAt:         p.UpdatedAt.Unix(),
	}
}
