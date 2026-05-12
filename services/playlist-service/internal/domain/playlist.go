package domain

import "time"

type Playlist struct {
	ID              string
	OwnerID         string
	Name            string
	Description     string
	Public          bool
	Permalink       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CollaboratorIDs []string
	Tracks          []*PlaylistTrack
}

type PlaylistTrack struct {
	PlaylistID string
	TrackID    string
	Rank       string
	AddedBy    string
	AddedAt    time.Time
}

type Collaborator struct {
	PlaylistID string
	UserID     string
	AddedAt    time.Time
}
