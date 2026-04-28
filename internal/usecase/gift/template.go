package gift

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

// EnsureCashTemplate garante um gift modelo de contribuição em dinheiro por wedding (id determinístico).
func EnsureCashTemplate(ctx context.Context, r repository.GiftRepository, weddingID string) error {
	_, err := r.FindCashTemplateByWeddingID(ctx, weddingID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, entity.ErrNotFound) {
		return fmt.Errorf("gift.EnsureCashTemplate: find: %w", err)
	}

	now := time.Now()
	id := entity.GiftCashTemplateID(weddingID)
	g := &entity.Gift{
		ID:          id,
		WeddingID:   weddingID,
		Name:        entity.PaymentCashGiftLabel,
		Description: "",
		Price:       0,
		ImageURL:    "",
		Category:    "cash",
		Status:      entity.GiftStatusAvailable,
		Kind:        entity.GiftKindCashTemplate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := r.Create(ctx, g); err != nil {
		return fmt.Errorf("gift.EnsureCashTemplate: create: %w", err)
	}
	return nil
}
