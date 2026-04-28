package repository

import (
	"context"

	"github.com/by-r2/weddo-api/internal/domain/entity"
)

type PaymentRepository interface {
	// CreateWithItems persiste o cabeçalho e as linhas do checkout em uma transação única (pagamento já aprovado no gateway antes).
	CreateWithItems(ctx context.Context, payment *entity.Payment, items []entity.PaymentItem) error
	FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error)
	// FindByIDAny busca pelo id do pagamento (ex.: webhooks que enviam order_nsu = payment.id).
	FindByIDAny(ctx context.Context, id string) (*entity.Payment, error)
	FindByProviderID(ctx context.Context, providerID string) (*entity.Payment, error)
	FindItemsByPaymentID(ctx context.Context, weddingID, paymentID string) ([]entity.PaymentItem, error)
	List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error)
	Update(ctx context.Context, payment *entity.Payment) error
	SumByWedding(ctx context.Context, weddingID string) (totalRevenue float64, totalPayments int, err error)
}
