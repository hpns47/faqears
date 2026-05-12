package domain

import "time"

const (
	StatusCreated       = "created"
	StatusWaiting       = "waiting"
	StatusConfirming    = "confirming"
	StatusConfirmed     = "confirmed"
	StatusSending       = "sending"
	StatusFinished      = "finished"
	StatusPartiallyPaid = "partially_paid"
	StatusFailed        = "failed"
	StatusExpired       = "expired"
	StatusRefunded      = "refunded"

	ProviderNowPayments = "nowpayments"
)

type Payment struct {
	ID                string
	UserID            string
	Provider          string
	ProviderPaymentID string
	Status            string
	PriceCurrency     string
	PriceAmount       string
	PayCurrency       string
	PayAddress        string
	PayAmount         string
	Purpose           string
	InvoiceURL        string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Event struct {
	ID         int64
	PaymentID  string
	Status     string
	OccurredAt time.Time
	RawPayload []byte
}
