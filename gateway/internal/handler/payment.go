package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	paymentv1 "github.com/faqears/faqears/gen/go/payment/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/go-chi/chi/v5"
)

type PaymentHandler struct {
	c paymentv1.PaymentServiceClient
}

func NewPaymentHandler(c paymentv1.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{c: c}
}

func (h *PaymentHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Purpose       string `json:"purpose"`
		PriceCurrency string `json:"price_currency"`
		PriceAmount   string `json:"price_amount"`
		PayCurrency   string `json:"pay_currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.CreateInvoice(ctx, &paymentv1.CreateInvoiceRequest{
		UserId:        identity.UserID,
		Purpose:       body.Purpose,
		PriceCurrency: body.PriceCurrency,
		PriceAmount:   body.PriceAmount,
		PayCurrency:   body.PayCurrency,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"payment": resp.Payment,
		"pay_url": resp.PayUrl,
	})
}

func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetPayment(ctx, &paymentv1.GetPaymentRequest{PaymentId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Payment)
}

func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	identity := callerIdentity(r.Context())
	q := r.URL.Query()
	limit := int32(20)
	offset := int32(0)
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.ListPaymentsByUser(ctx, &paymentv1.ListPaymentsByUserRequest{
		UserId: identity.UserID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"payments": resp.Payments})
}

func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}
	sig := r.Header.Get("x-nowpayments-sig")
	resp, err := h.c.HandleWebhook(r.Context(), &paymentv1.HandleWebhookRequest{
		RawBody:   body,
		Signature: sig,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
