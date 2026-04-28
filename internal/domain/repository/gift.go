package repository

import (
	"context"

	"github.com/by-r2/weddo-api/internal/domain/entity"
)

type GiftRepository interface {
	Create(ctx context.Context, gift *entity.Gift) error
	FindByID(ctx context.Context, weddingID, id string) (*entity.Gift, error)
	// FindCashTemplateByWeddingID retorna o gift modelo de contribuição em dinheiro (único por casamento).
	FindCashTemplateByWeddingID(ctx context.Context, weddingID string) (*entity.Gift, error)
	// List com catalogOnly=true exclui kind=cash_template (lista pública/admin de presentes de catálogo).
	List(ctx context.Context, weddingID string, page, perPage int, category, status, search string, catalogOnly bool) ([]entity.Gift, int, error)
	Update(ctx context.Context, gift *entity.Gift) error
	Delete(ctx context.Context, weddingID, id string) error
	CountByWedding(ctx context.Context, weddingID string) (total, available, purchased int, err error)
	// ListCategories retorna categorias distintas dos gifts de catálogo (ordenadas), sem repetir valores após trim; vazio se não houver.
	ListCategories(ctx context.Context, weddingID string) ([]string, error)
}
