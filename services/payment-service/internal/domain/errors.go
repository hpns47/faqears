package domain

import "errors"

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrUnsupportedProvider  = errors.New("unsupported provider")
	ErrStateMachine         = errors.New("invalid state transition")
)
