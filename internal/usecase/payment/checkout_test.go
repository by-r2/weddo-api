package payment_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	gw "github.com/by-r2/weddo-api/internal/domain/gateway"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"github.com/by-r2/weddo-api/internal/usecase/payment"
)

func TestCheckout_emptyItems(t *testing.T) {
	t.Parallel()
	uc := payment.NewUseCase(nil, nil, &gwFake{})
	_, err := uc.Checkout(context.Background(), payment.CheckoutInput{WeddingID: "w1", Lines: nil})
	if err != payment.ErrCheckoutEmptyItems {
		t.Fatalf("got %v", err)
	}
}

func TestCheckout_duplicateGiftID(t *testing.T) {
	t.Parallel()
	id := uuid.New().String()
	gCatalog := entity.Gift{
		ID:        id,
		WeddingID: "w1",
		Name:      "X",
		Price:     10,
		Status:    entity.GiftStatusAvailable,
		Kind:      entity.GiftKindCatalog,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	uc := payment.NewUseCase(&payFake{}, &giftFake{gift: &gCatalog}, &gwFake{})
	_, err := uc.Checkout(context.Background(), payment.CheckoutInput{
		WeddingID:     "w1",
		Lines:         []payment.CheckoutLineInput{{GiftID: id}, {GiftID: id}},
		PaymentMethod: "pix",
		PayerName:     "A",
		PayerEmail:    "a@test.com",
	})
	if err != payment.ErrCheckoutDuplicateGiftLine {
		t.Fatalf("got %v", err)
	}
}

func TestCheckout_catalogExtrafieldsRejected(t *testing.T) {
	t.Parallel()
	id := uuid.New().String()
	gCatalog := entity.Gift{
		ID:        id,
		WeddingID: "w1",
		Name:      "X",
		Price:     10,
		Status:    entity.GiftStatusAvailable,
		Kind:      entity.GiftKindCatalog,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	am := 10.0
	uc := payment.NewUseCase(&payFake{}, &giftFake{gift: &gCatalog}, &gwFake{})
	_, err := uc.Checkout(context.Background(), payment.CheckoutInput{
		WeddingID:     "w1",
		Lines:         []payment.CheckoutLineInput{{GiftID: id, Amount: &am}},
		PaymentMethod: "pix",
		PayerName:     "A",
		PayerEmail:    "a@test.com",
	})
	if err != payment.ErrCheckoutInvalidCatalogExtras {
		t.Fatalf("got %v", err)
	}
}

func TestCheckout_successWritesItems(t *testing.T) {
	t.Parallel()
	id := uuid.New().String()
	gCatalog := entity.Gift{
		ID:        id,
		WeddingID: "w1",
		Name:      "Panela",
		Price:     99,
		Status:    entity.GiftStatusAvailable,
		Kind:      entity.GiftKindCatalog,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo := &payFake{}
	uc := payment.NewUseCase(repo, &giftFake{gift: &gCatalog}, &gwFake{status: "pending"})
	_, err := uc.Checkout(context.Background(), payment.CheckoutInput{
		WeddingID:     "w1",
		Lines:         []payment.CheckoutLineInput{{GiftID: id}},
		PaymentMethod: "pix",
		PayerName:     "A",
		PayerEmail:    "a@test.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.lastItemsLen != 1 || repo.savedTotal != 99 {
		t.Fatalf("persist: items=%d total=%v", repo.lastItemsLen, repo.savedTotal)
	}
}

// --- fakes ---

type gwFake struct {
	status string
}

func (g *gwFake) Name() string { return "test" }

func (g *gwFake) CreatePayment(context.Context, gw.CreatePaymentInput) (*gw.PaymentResult, error) {
	s := g.status
	if s == "" {
		s = "pending"
	}
	return &gw.PaymentResult{ProviderID: "prov-1", Status: s}, nil
}

func (g *gwFake) GetPaymentStatus(context.Context, string) (*gw.WebhookResult, error) {
	return &gw.WebhookResult{ProviderID: "prov-1", Status: "pending"}, nil
}

var _ gw.PaymentGateway = (*gwFake)(nil)

type giftFake struct {
	gift *entity.Gift
}

func (g *giftFake) Create(context.Context, *entity.Gift) error { panic("unexpected") }

func (g *giftFake) FindByID(ctx context.Context, weddingID string, id string) (*entity.Gift, error) {
	if g.gift != nil && g.gift.ID == id && weddingID == g.gift.WeddingID {
		cp := *g.gift
		return &cp, nil
	}
	return nil, entity.ErrNotFound
}

func (g *giftFake) FindCashTemplateByWeddingID(context.Context, string) (*entity.Gift, error) {
	return nil, entity.ErrNotFound
}

func (g *giftFake) List(context.Context, string, repository.GiftListParams) ([]entity.Gift, int, error) {
	return nil, 0, nil
}

func (g *giftFake) Update(context.Context, *entity.Gift) error { return nil }

func (g *giftFake) Delete(context.Context, string, string) error { return nil }

func (g *giftFake) CountByWedding(context.Context, string) (int, int, int, error) {
	return 0, 0, 0, nil
}

func (g *giftFake) ListCategories(context.Context, string) ([]repository.GiftCategoryCount, error) {
	return []repository.GiftCategoryCount{}, nil
}

var _ repository.GiftRepository = (*giftFake)(nil)

type payFake struct {
	lastItemsLen int
	savedTotal   float64
}

func (p *payFake) CreateWithItems(_ context.Context, pay *entity.Payment, items []entity.PaymentItem) error {
	p.lastItemsLen = len(items)
	p.savedTotal = pay.Amount
	return nil
}

func (p *payFake) FindByID(context.Context, string, string) (*entity.Payment, error) {
	return nil, entity.ErrNotFound
}

func (p *payFake) FindByIDAny(context.Context, string) (*entity.Payment, error) {
	return nil, entity.ErrNotFound
}

func (p *payFake) FindByProviderID(context.Context, string) (*entity.Payment, error) {
	return nil, entity.ErrNotFound
}

func (p *payFake) FindItemsByPaymentID(context.Context, string, string) ([]entity.PaymentItem, error) {
	return nil, nil
}

func (p *payFake) List(context.Context, string, int, int, string, string) ([]entity.Payment, int, error) {
	return nil, 0, nil
}

func (p *payFake) Update(context.Context, *entity.Payment) error { return nil }

func (p *payFake) SumByWedding(context.Context, string) (float64, int, error) { return 0, 0, nil }

var _ repository.PaymentRepository = (*payFake)(nil)
