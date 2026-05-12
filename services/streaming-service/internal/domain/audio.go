package domain

import "time"

type TrackAudio struct {
	TrackID     string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
	UploadedBy  string
	UploadedAt  time.Time
}

type Session struct {
	SessionID string
	UserID    string
	TrackID   string
	StartedAt int64
}
