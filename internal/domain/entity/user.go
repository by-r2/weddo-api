package entity

import "time"

type User struct {
	ID           string
	WeddingID    string
	Name         string
	Email        string
	AvatarURL    string
	GoogleID     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
