package domain

import "time"

type TrackRef struct {
	TrackID string
	Score   float64
}

type PlayEvent struct {
	UserID    string
	TrackID   string
	SessionID string
	EventType string
	OccurredAt time.Time
}

type Similarity struct {
	TrackID string
	CoCount int64
}
