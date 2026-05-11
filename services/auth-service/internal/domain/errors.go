package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrEmailTaken         = errs.AlreadyExists("email already registered")
	ErrInvalidCredentials = errs.Unauthenticated("invalid email or password")
	ErrUserNotFound       = errs.NotFound("user not found")
	ErrInvalidToken       = errs.Unauthenticated("invalid or expired token")
	ErrTokenRevoked       = errs.Unauthenticated("token has been revoked")
	ErrWeakPassword       = errs.InvalidArgument("password must be at least 8 characters")
	ErrInvalidEmail       = errs.InvalidArgument("invalid email format")
)
