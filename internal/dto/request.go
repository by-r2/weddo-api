package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RSVP

type RSVPRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// Invitations

type CreateInvitationRequest struct {
	Code      string               `json:"code" validate:"required,max=50"`
	Label     string               `json:"label" validate:"required,max=100"`
	MaxGuests int                  `json:"max_guests" validate:"required,min=1"`
	Notes     string               `json:"notes"`
	Guests    []CreateGuestInline  `json:"guests"`
}

type CreateGuestInline struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateInvitationRequest struct {
	Code      string `json:"code" validate:"required,max=50"`
	Label     string `json:"label" validate:"required,max=100"`
	MaxGuests int    `json:"max_guests" validate:"required,min=1"`
	Notes     string `json:"notes"`
}

// Guests

type AddGuestRequest struct {
	Name  string `json:"name" validate:"required,max=100"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type UpdateGuestRequest struct {
	Name   string `json:"name" validate:"required,max=100"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Status string `json:"status" validate:"omitempty,oneof=pending confirmed declined"`
}
