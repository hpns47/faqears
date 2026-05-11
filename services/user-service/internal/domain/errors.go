package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrUserNotFound    = errs.NotFound("user not found")
	ErrAlreadyFollowed = errs.AlreadyExists("already followed")
	ErrSelfFollow      = errs.InvalidArgument("cannot follow yourself")
	ErrInvalidUserID   = errs.InvalidArgument("invalid user id")
)
