package models

import "time"

type RefreshToken struct {
	ID        int       `json:"-"`
	TokenHash string    `json:"-"`
	UserID    int       `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"creates_at"`
}
