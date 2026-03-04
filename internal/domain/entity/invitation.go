package entity

import "time"

type Invitation struct {
	ID        string
	WeddingID string
	Code      string
	Label     string
	MaxGuests int
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Guests []Guest
}
