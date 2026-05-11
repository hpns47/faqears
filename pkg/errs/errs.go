package errs

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindNotFound
	KindAlreadyExists
	KindInvalidArgument
	KindUnauthenticated
	KindPermissionDenied
	KindConflict
	KindInternal
	KindUnavailable
)

type Error struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func New(kind Kind, msg string) *Error {
	return &Error{Kind: kind, Message: msg}
}

func Wrap(kind Kind, msg string, cause error) *Error {
	return &Error{Kind: kind, Message: msg, Cause: cause}
}

func NotFound(msg string) *Error          { return New(KindNotFound, msg) }
func AlreadyExists(msg string) *Error     { return New(KindAlreadyExists, msg) }
func InvalidArgument(msg string) *Error   { return New(KindInvalidArgument, msg) }
func Unauthenticated(msg string) *Error   { return New(KindUnauthenticated, msg) }
func PermissionDenied(msg string) *Error  { return New(KindPermissionDenied, msg) }
func Conflict(msg string) *Error          { return New(KindConflict, msg) }
func Internal(msg string, cause error) *Error {
	return Wrap(KindInternal, msg, cause)
}

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if !errors.As(err, &e) {
		return status.Error(codes.Internal, err.Error())
	}
	return status.Error(toCode(e.Kind), e.Message)
}

func toCode(k Kind) codes.Code {
	switch k {
	case KindNotFound:
		return codes.NotFound
	case KindAlreadyExists:
		return codes.AlreadyExists
	case KindInvalidArgument:
		return codes.InvalidArgument
	case KindUnauthenticated:
		return codes.Unauthenticated
	case KindPermissionDenied:
		return codes.PermissionDenied
	case KindConflict:
		return codes.FailedPrecondition
	case KindUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}
