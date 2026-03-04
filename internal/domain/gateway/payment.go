package gateway

import (
	"context"
	"time"
)

// CreatePaymentInput contém os dados necessários para criar um pagamento
// independente do provedor (Mercado Pago, InfinitePay, etc.).
type CreatePaymentInput struct {
	Amount            float64
	Description       string
	PayerName         string
	PayerEmail        string
	PayerPhone        string
	PaymentMethod     string // "pix" ou "credit_card"
	ExternalReference string

	// Campos específicos para cartão (Mercado Pago)
	CardToken       string
	PaymentMethodID string
	Installments    int

	// URL de retorno após pagamento (InfinitePay)
	RedirectURL string
}

// PaymentResult contém o resultado da criação de um pagamento.
// Campos são preenchidos conforme o provedor.
type PaymentResult struct {
	ProviderID   string
	Status       string
	QRCode       string
	QRCodeBase64 string
	ExpiresAt    *time.Time

	// URL do checkout externo (InfinitePay) — frontend redireciona para cá
	CheckoutURL string
}

// WebhookResult contém o status de um pagamento obtido via consulta ao provedor.
type WebhookResult struct {
	ProviderID string
	Status     string
}

// PaymentGateway define o contrato que qualquer provedor de pagamento deve implementar.
type PaymentGateway interface {
	// Name retorna o identificador do provedor ("mercadopago" ou "infinitepay").
	Name() string

	// CreatePayment cria um pagamento no provedor externo.
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*PaymentResult, error)

	// GetPaymentStatus consulta o status de um pagamento existente.
	GetPaymentStatus(ctx context.Context, providerID string) (*WebhookResult, error)
}
