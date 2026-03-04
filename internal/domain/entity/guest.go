package entity

import "time"

type GuestStatus string

const (
	GuestStatusPending   GuestStatus = "pending"
	GuestStatusConfirmed GuestStatus = "confirmed"
	GuestStatusDeclined  GuestStatus = "declined"
)

type Guest struct {
	ID           string
	InvitationID string
	WeddingID    string
	Name         string
	Phone        string
	Email        string
	Status       GuestStatus
	ConfirmedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
