package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrArtistNotFound = errs.NotFound("artist not found")
	ErrAlbumNotFound  = errs.NotFound("album not found")
	ErrTrackNotFound  = errs.NotFound("track not found")
	ErrISRCTaken      = errs.AlreadyExists("isrc already exists")
)
