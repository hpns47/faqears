package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrJobNotFound      = errs.NotFound("job not found")
	ErrNotOwner         = errs.PermissionDenied("caller does not own this job")
	ErrNoLyrics         = errs.InvalidArgument("job has no lyrics yet")
	ErrPremiumRequired  = errs.PermissionDenied("premium tier required")
	ErrJobNotCancellable = errs.InvalidArgument("job cannot be cancelled in current status")
)
