package dto

type AuthResponse struct {
	Token   string         `json:"token"`
	Wedding WeddingSummary `json:"wedding"`
	User    UserSummary    `json:"user"`
}

type WeddingSummary struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type UserSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
	HasGoogle bool   `json:"has_google"`
	CreatedAt string `json:"created_at"`
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
	ID     string `json:"id"`
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

// Gifts

type GiftResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url,omitempty"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
	Kind        string  `json:"kind,omitempty"` // catalog | cash_template — omitido no público quando vazio
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type GiftCategoriesResponse struct {
	Categories []GiftCategoryItem `json:"categories"`
}

type GiftCategoryItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Payments

type CheckoutResponse struct {
	PaymentID    string  `json:"payment_id"`
	ProviderID   string  `json:"provider_id"`
	Status       string  `json:"status"`
	CheckoutURL  string  `json:"checkout_url,omitempty"`
	QRCode       string  `json:"qr_code,omitempty"`
	QRCodeBase64 string  `json:"qr_code_base64,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
}

type PaymentLineResponse struct {
	GiftID            string  `json:"gift_id"`
	Kind              string  `json:"kind"`
	Amount            float64 `json:"amount"`
	CustomName        string  `json:"custom_name,omitempty"`
	CustomDescription string  `json:"custom_description,omitempty"`
	Label             string  `json:"label"`
}

type PaymentResponse struct {
	ID            string                `json:"id"`
	ProviderID    string                `json:"provider_id,omitempty"`
	Amount        float64               `json:"amount"`
	Status        string                `json:"status"`
	PaymentMethod string                `json:"payment_method"`
	PayerName     string                `json:"payer_name"`
	PayerEmail    string                `json:"payer_email,omitempty"`
	Message       string                `json:"message,omitempty"`
	Items         []PaymentLineResponse `json:"items"`
	PaidAt        *string               `json:"paid_at"`
	CreatedAt     string                `json:"created_at"`
}

type PaymentLineStatusDTO struct {
	GiftID string  `json:"gift_id"`
	Kind   string  `json:"kind"`
	Amount float64 `json:"amount"`
	Label  string  `json:"label"`
}

type PaymentStatusResponse struct {
	PaymentID string                 `json:"payment_id"`
	Status    string                 `json:"status"`
	Lines     []PaymentLineStatusDTO `json:"lines"`
}

// Dashboard

type DashboardResponse struct {
	RSVP  RSVPStats  `json:"rsvp"`
	Gifts *GiftStats `json:"gifts,omitempty"`
}

type RSVPStats struct {
	TotalInvitations int     `json:"total_invitations"`
	TotalGuests      int     `json:"total_guests"`
	Confirmed        int     `json:"confirmed"`
	Pending          int     `json:"pending"`
	Declined         int     `json:"declined"`
	ConfirmationRate float64 `json:"confirmation_rate"`
}

type GiftStats struct {
	TotalGifts    int     `json:"total_gifts"`
	Purchased     int     `json:"purchased"`
	Available     int     `json:"available"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalPayments int     `json:"total_payments"`
}
