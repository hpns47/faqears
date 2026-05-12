package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrAudioNotFound        = errs.NotFound("audio not found")
	ErrAlreadyExists        = errs.AlreadyExists("audio already exists for track")
	ErrInvalidContentType   = errs.InvalidArgument("invalid content type: must start with audio/")
	ErrSessionNotFound      = errs.NotFound("session not found")
)
