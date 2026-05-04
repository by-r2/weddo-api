package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Partner1Name string `json:"partner1_name" validate:"required,max=100"`
	Partner2Name string `json:"partner2_name" validate:"required,max=100"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8,max=72"`
	Date         string `json:"date"`
	Slug         string `json:"slug" validate:"omitempty,max=100"`
}

type GoogleAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

type RegisterGoogleRequest struct {
	Partner1Name string `json:"partner1_name" validate:"required,max=100"`
	Partner2Name string `json:"partner2_name" validate:"required,max=100"`
	IDToken      string `json:"id_token" validate:"required"`
	Date         string `json:"date"`
	Slug         string `json:"slug" validate:"omitempty,max=100"`
}

type InviteUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"omitempty,max=100"`
}

// RSVP

type RSVPRequest struct {
	Code       string `json:"code" validate:"required,max=50"`
	Name       string `json:"name" validate:"required,max=100"`
	WillAttend *bool  `json:"will_attend,omitempty"` // omitido = confirma presença (retrocompatível)
}

// Invitations

type CreateInvitationRequest struct {
	Label     string              `json:"label" validate:"required,max=100"`
	MaxGuests int                 `json:"max_guests" validate:"required,min=1"`
	Notes     string              `json:"notes"`
	Guests    []CreateGuestInline `json:"guests"`
}

type CreateGuestInline struct {
	Name   string `json:"name" validate:"required,max=100"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Status string `json:"status" validate:"omitempty,oneof=pending confirmed declined"`
}

type UpdateInvitationRequest struct {
	Code      string `json:"code" validate:"required,max=50"`
	Label     string `json:"label" validate:"required,max=100"`
	MaxGuests int    `json:"max_guests" validate:"required,min=1"`
	Notes     string `json:"notes"`
}

// Guests

type AddGuestRequest struct {
	Name   string `json:"name" validate:"required,max=100"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Status string `json:"status" validate:"omitempty,oneof=pending confirmed declined"`
}

type UpdateGuestRequest struct {
	Name   string `json:"name" validate:"required,max=100"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Status string `json:"status" validate:"omitempty,oneof=pending confirmed declined"`
}

// Gifts

type CreateGiftRequest struct {
	Name        string  `json:"name" validate:"required,max=200"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category" validate:"required,max=50"`
}

type UpdateGiftRequest struct {
	Name        string  `json:"name" validate:"required,max=200"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category" validate:"required,max=50"`
	Status      string  `json:"status" validate:"omitempty,oneof=available purchased"`
}

// Payments — checkout público (carrinho no cliente).

type CheckoutItemRequest struct {
	GiftID            string   `json:"gift_id" validate:"required,max=128"`
	Amount            *float64 `json:"amount"` // obrigatório apenas para contribuição em dinheiro (gift cash_template)
	CustomName        string   `json:"custom_name" validate:"max=200"`
	CustomDescription string   `json:"custom_description"`
}

type CheckoutRequest struct {
	Items           []CheckoutItemRequest `json:"items" validate:"required,min=1,dive"`
	PayerName       string                `json:"payer_name" validate:"required,max=100"`
	PayerEmail      string                `json:"payer_email" validate:"required,email"`
	Message         string                `json:"message"`
	PaymentMethod   string                `json:"payment_method" validate:"required,oneof=pix credit_card"`
	CardToken       string                `json:"card_token"`
	PaymentMethodID string                `json:"payment_method_id"`
	Installments    int                   `json:"installments"`
	RedirectURL     string                `json:"redirect_url"`
}
