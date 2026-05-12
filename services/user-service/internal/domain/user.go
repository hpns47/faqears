package domain

import "time"

type User struct {
	ID          string
	Email       string
	DisplayName string
	AvatarURL   string
	Country     string
	Language    string
	Tier        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	TierFree    = "free"
	TierPremium = "premium"
)

type Follow struct {
	FollowerID string
	FolloweeID string
	CreatedAt  time.Time
}

type Profile struct {
	DisplayName string
	AvatarURL   string
	Country     string
	Language    string
}
