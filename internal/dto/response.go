package dto

type LoginResponse struct {
	Token   string          `json:"token"`
	Wedding WeddingSummary  `json:"wedding"`
}

type WeddingSummary struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaginatedResponse struct {
	Data any            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// RSVP

type RSVPResponse struct {
	Guest      GuestSummary      `json:"guest"`
	Invitation InvitationSummary `json:"invitation"`
	Message    string            `json:"message"`
}

type RSVPInvitationResponse struct {
	Invitation InvitationSummary `json:"invitation"`
	Guests     []GuestPublic     `json:"guests"`
}

type GuestPublic struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Invitations

type InvitationResponse struct {
	ID        string          `json:"id"`
	Code      string          `json:"code"`
	Label     string          `json:"label"`
	MaxGuests int             `json:"max_guests"`
	Notes     string          `json:"notes,omitempty"`
	Guests    []GuestResponse `json:"guests,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type InvitationSummary struct {
	Label     string `json:"label"`
	MaxGuests int    `json:"max_guests,omitempty"`
}

// Guests

type GuestResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Phone       string  `json:"phone,omitempty"`
	Email       string  `json:"email,omitempty"`
	Status      string  `json:"status"`
	ConfirmedAt *string `json:"confirmed_at"`
}

type GuestSummary struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	ConfirmedAt *string `json:"confirmed_at"`
}

// Dashboard

type DashboardResponse struct {
	RSVP RSVPStats `json:"rsvp"`
}

type RSVPStats struct {
	TotalInvitations int     `json:"total_invitations"`
	TotalGuests      int     `json:"total_guests"`
	Confirmed        int     `json:"confirmed"`
	Pending          int     `json:"pending"`
	Declined         int     `json:"declined"`
	ConfirmationRate float64 `json:"confirmation_rate"`
}
