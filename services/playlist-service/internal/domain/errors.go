package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrNotFound              = errs.NotFound("playlist not found")
	ErrPermissionDenied      = errs.PermissionDenied("not authorized to modify this playlist")
	ErrTrackAlreadyInPlaylist = errs.Conflict("track already in playlist")
	ErrTrackNotInPlaylist    = errs.NotFound("track not in playlist")
)
