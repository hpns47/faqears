package nowpayments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/payment-service/internal/domain"
)

type Client struct {
	baseURL     string
	apiKey      string
	ipnSecret   string
	callbackURL string
	http        *http.Client
}

func NewClient(baseURL, apiKey, ipnSecret, callbackURL string) *Client {
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiKey:      apiKey,
		ipnSecret:   ipnSecret,
		callbackURL: callbackURL,
		http:        &http.Client{},
	}
}

type invoiceRequest struct {
	PriceAmount     string `json:"price_amount"`
	PriceCurrency   string `json:"price_currency"`
	PayCurrency     string `json:"pay_currency"`
	OrderID         string `json:"order_id"`
	OrderDescription string `json:"order_description"`
	IPNCallbackURL  string `json:"ipn_callback_url"`
}

type invoiceResponse struct {
	ID         string `json:"id"`
	InvoiceURL string `json:"invoice_url"`
	PayAddress string `json:"pay_address"`
	PayAmount  string `json:"pay_amount"`
}

func (c *Client) CreateInvoice(ctx context.Context, p *domain.Payment) (providerPaymentID, invoiceURL, payAddress, payAmount string, err error) {
	if c.apiKey == "" {
		return "", "", "", "", errs.New(errs.KindUnavailable, "nowpayments api key not configured")
	}

	reqBody := invoiceRequest{
		PriceAmount:      p.PriceAmount,
		PriceCurrency:    p.PriceCurrency,
		PayCurrency:      p.PayCurrency,
		OrderID:          p.ID,
		OrderDescription: p.Purpose,
		IPNCallbackURL:   c.callbackURL,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", "", errs.Internal("marshal invoice request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/invoice", bytes.NewReader(b))
	if err != nil {
		return "", "", "", "", errs.Internal("build invoice request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", "", "", errs.Internal("call nowpayments", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", "", errs.Internal("read nowpayments response", err)
	}

	if resp.StatusCode >= 400 {
		return "", "", "", "", errs.Internal(fmt.Sprintf("nowpayments error %d: %s", resp.StatusCode, string(raw)), nil)
	}

	var out invoiceResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", "", "", "", errs.Internal("decode nowpayments response", err)
	}

	return out.ID, out.InvoiceURL, out.PayAddress, out.PayAmount, nil
}

func (c *Client) VerifySignature(body []byte, sig string) error {
	if c.ipnSecret == "" {
		return domain.ErrInvalidSignature
	}

	sorted, err := sortedJSONBody(body)
	if err != nil {
		return domain.ErrInvalidSignature
	}

	mac := hmac.New(sha512.New, []byte(c.ipnSecret))
	mac.Write([]byte(sorted))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return domain.ErrInvalidSignature
	}
	return nil
}

func sortedJSONBody(body []byte) (string, error) {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return "", err
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteByte('{')
	for i, k := range keys {
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return "", err
		}
		valJSON, err := json.Marshal(m[k])
		if err != nil {
			return "", err
		}
		sb.Write(keyJSON)
		sb.WriteByte(':')
		sb.Write(valJSON)
		if i < len(keys)-1 {
			sb.WriteByte(',')
		}
	}
	sb.WriteByte('}')
	return sb.String(), nil
}
