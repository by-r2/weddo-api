package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	gw "github.com/by-r2/weddo-api/internal/domain/gateway"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"github.com/google/uuid"
)

var ErrGiftUnavailable = errors.New("presente não está disponível")

// ErrCashAmountOutOfRange valor fora dos limites de contribuição em dinheiro.
var ErrCashAmountOutOfRange = errors.New("valor deve estar entre R$ 1,00 e R$ 100.000,00")

const minCashGiftBRL = 1.0
const maxCashGiftBRL = 100_000.0

const maxCustomNameLen = 200
const maxCustomDescriptionLen = 2000

type UseCase struct {
	paymentRepo repository.PaymentRepository
	giftRepo    repository.GiftRepository
	gateway     gw.PaymentGateway
}

func NewUseCase(pr repository.PaymentRepository, gr repository.GiftRepository, gateway gw.PaymentGateway) *UseCase {
	return &UseCase{paymentRepo: pr, giftRepo: gr, gateway: gateway}
}

// ProviderName retorna o nome do provedor de pagamento ativo.
func (uc *UseCase) ProviderName() string {
	return uc.gateway.Name()
}

type CheckoutLineInput struct {
	GiftID            string
	Amount            *float64 // obrigatório para cash_template quando enviado; catálogo: omitido
	CustomName        string
	CustomDescription string
}

type CheckoutInput struct {
	WeddingID       string
	Lines           []CheckoutLineInput
	PayerName       string
	PayerEmail      string
	Message         string
	PaymentMethod   string
	CardToken       string
	PaymentMethodID string
	Installments    int
	RedirectURL     string
}

// PurchaseResult resultado do fluxo checkout (PIX/cartão).
type PurchaseResult struct {
	Payment      *entity.Payment
	QRCode       string
	QRCodeBase64 string
	CheckoutURL  string
}

// PaymentStatusLine linha para exibição de status do checkout.
type PaymentStatusLine struct {
	GiftID string          `json:"gift_id"`
	Kind   entity.GiftKind `json:"kind"`
	Amount float64         `json:"amount"`
	Label  string          `json:"label"`
}

// Checkout cria cobrança com N linhas (catálogo e/ou contribuição em dinheiro).
func (uc *UseCase) Checkout(ctx context.Context, input CheckoutInput) (*PurchaseResult, error) {
	if len(input.Lines) == 0 {
		return nil, ErrCheckoutEmptyItems
	}

	seen := make(map[string]struct{}, len(input.Lines))
	var total float64
	var items []entity.PaymentItem

	for _, line := range input.Lines {
		if _, dup := seen[line.GiftID]; dup {
			return nil, ErrCheckoutDuplicateGiftLine
		}
		seen[line.GiftID] = struct{}{}

		gift, err := uc.giftRepo.FindByID(ctx, input.WeddingID, line.GiftID)
		if err != nil {
			return nil, err
		}
		if gift.Kind == "" {
			gift.Kind = entity.GiftKindCatalog
		}

		it := entity.PaymentItem{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
		}

		switch gift.Kind {
		case entity.GiftKindCatalog:
			if line.Amount != nil {
				return nil, ErrCheckoutInvalidCatalogExtras
			}
			if strings.TrimSpace(line.CustomName) != "" || strings.TrimSpace(line.CustomDescription) != "" {
				return nil, ErrCheckoutInvalidCatalogExtras
			}
			if gift.Status != entity.GiftStatusAvailable {
				return nil, ErrGiftUnavailable
			}
			it.GiftID = gift.ID
			it.Amount = math.Round(gift.Price*100) / 100

		case entity.GiftKindCashTemplate:
			if line.Amount == nil {
				return nil, ErrCheckoutCashAmountMissing
			}
			am := math.Round((*line.Amount)*100) / 100
			if am < minCashGiftBRL || am > maxCashGiftBRL {
				return nil, ErrCashAmountOutOfRange
			}
			if len([]rune(line.CustomName)) > maxCustomNameLen || len([]rune(line.CustomDescription)) > maxCustomDescriptionLen {
				return nil, fmt.Errorf("texto da contribuição excede o limite")
			}
			it.GiftID = gift.ID
			it.Amount = am
			it.CustomName = strings.TrimSpace(line.CustomName)
			it.CustomDescription = strings.TrimSpace(line.CustomDescription)

		default:
			return nil, fmt.Errorf("tipo de presente não suportado")
		}

		total += it.Amount
		items = append(items, it)
	}

	total = math.Round(total*100) / 100
	now := time.Now()
	p := &entity.Payment{
		ID:            uuid.New().String(),
		WeddingID:     input.WeddingID,
		Amount:        total,
		Status:        entity.PaymentStatusPending,
		PaymentMethod: entity.PaymentMethod(input.PaymentMethod),
		PayerName:     input.PayerName,
		PayerEmail:    input.PayerEmail,
		Message:       input.Message,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for i := range items {
		items[i].PaymentID = p.ID
	}

	desc := buildGatewayDescription(ctx, uc, input.WeddingID, items)
	gwInput := gw.CreatePaymentInput{
		Amount:            total,
		Description:       desc,
		PayerName:         input.PayerName,
		PayerEmail:        input.PayerEmail,
		PaymentMethod:     input.PaymentMethod,
		ExternalReference: p.ID,
		CardToken:         input.CardToken,
		PaymentMethodID:   input.PaymentMethodID,
		Installments:      input.Installments,
		RedirectURL:       input.RedirectURL,
	}

	gwResult, err := uc.gateway.CreatePayment(ctx, gwInput)
	if err != nil {
		return nil, fmt.Errorf("payment.Checkout: gateway: %w", err)
	}

	p.ProviderID = gwResult.ProviderID
	p.Status = mapProviderStatus(gwResult.Status)
	p.PixQRCode = gwResult.QRCode
	p.PixExpiration = gwResult.ExpiresAt

	if p.Status == entity.PaymentStatusApproved {
		t := now
		p.PaidAt = &t
		if err := uc.markCatalogPurchased(ctx, input.WeddingID, items); err != nil {
			slog.Error("payment.Checkout: mark catalog gifts", "payment_id", p.ID, "error", err)
		}
	}

	if err := uc.paymentRepo.CreateWithItems(ctx, p, items); err != nil {
		return nil, fmt.Errorf("payment.Checkout: save: %w", err)
	}

	return &PurchaseResult{
		Payment:      p,
		QRCode:       gwResult.QRCode,
		QRCodeBase64: gwResult.QRCodeBase64,
		CheckoutURL:  gwResult.CheckoutURL,
	}, nil
}

func buildGatewayDescription(ctx context.Context, uc *UseCase, weddingID string, items []entity.PaymentItem) string {
	if len(items) == 0 {
		return "Lista de presentes"
	}
	if len(items) == 1 {
		g, err := uc.giftRepo.FindByID(ctx, weddingID, items[0].GiftID)
		if err == nil && g.Kind == entity.GiftKindCatalog {
			return fmt.Sprintf("Presente: %s", g.Name)
		}
		if err == nil && g.Kind == entity.GiftKindCashTemplate {
			return entity.PaymentCashGiftLabel
		}
		return "Pagamento lista de presentes"
	}
	return fmt.Sprintf("%d presentes", len(items))
}

func (uc *UseCase) markCatalogPurchased(ctx context.Context, weddingID string, items []entity.PaymentItem) error {
	for _, it := range items {
		g, err := uc.giftRepo.FindByID(ctx, weddingID, it.GiftID)
		if err != nil || g.Kind != entity.GiftKindCatalog {
			continue
		}
		if g.Status != entity.GiftStatusAvailable {
			continue
		}
		now := time.Now()
		g.Status = entity.GiftStatusPurchased
		g.UpdatedAt = now
		if err := uc.giftRepo.Update(ctx, g); err != nil {
			slog.Error("payment.markCatalogPurchased: update", "gift_id", g.ID, "error", err)
		}
	}
	return nil
}

func (uc *UseCase) revertCatalogPurchases(ctx context.Context, weddingID string, items []entity.PaymentItem) {
	for _, it := range items {
		g, err := uc.giftRepo.FindByID(ctx, weddingID, it.GiftID)
		if err != nil || g.Kind != entity.GiftKindCatalog {
			continue
		}
		if g.Status != entity.GiftStatusPurchased {
			continue
		}
		now := time.Now()
		g.Status = entity.GiftStatusAvailable
		g.UpdatedAt = now
		if err := uc.giftRepo.Update(ctx, g); err != nil {
			slog.Error("payment.revertCatalogPurchases: update", "gift_id", g.ID, "error", err)
		}
	}
}

// PaymentDetailLine linha para admin + status público.
type PaymentDetailLine struct {
	LineID            string
	GiftID            string
	GiftKind          entity.GiftKind
	Amount            float64
	Label             string
	CustomName        string
	CustomDescription string
}

func (uc *UseCase) paymentLinesDetail(ctx context.Context, weddingID, paymentID string) ([]PaymentDetailLine, error) {
	items, err := uc.paymentRepo.FindItemsByPaymentID(ctx, weddingID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment.paymentLinesDetail: %w", err)
	}

	out := make([]PaymentDetailLine, 0, len(items))
	for _, it := range items {
		g, err := uc.giftRepo.FindByID(ctx, weddingID, it.GiftID)
		row := PaymentDetailLine{
			LineID:            it.ID,
			GiftID:            it.GiftID,
			Amount:            it.Amount,
			CustomName:        it.CustomName,
			CustomDescription: it.CustomDescription,
		}
		if err != nil {
			row.Label = "(presente removido)"
			row.GiftKind = entity.GiftKindCatalog
			out = append(out, row)
			continue
		}
		if g.Kind == "" {
			g.Kind = entity.GiftKindCatalog
		}
		row.GiftKind = g.Kind
		switch g.Kind {
		case entity.GiftKindCatalog:
			row.Label = g.Name
		case entity.GiftKindCashTemplate:
			if cn := strings.TrimSpace(it.CustomName); cn != "" {
				row.Label = cn
			} else {
				row.Label = entity.PaymentCashGiftLabel
			}
		default:
			row.Label = g.Name
		}
		out = append(out, row)
	}
	return out, nil
}

func (uc *UseCase) GetStatus(ctx context.Context, weddingID, paymentID string) (*entity.Payment, []PaymentStatusLine, error) {
	p, err := uc.paymentRepo.FindByID(ctx, weddingID, paymentID)
	if err != nil {
		return nil, nil, err
	}
	details, err := uc.paymentLinesDetail(ctx, weddingID, p.ID)
	if err != nil {
		return nil, nil, err
	}

	lines := make([]PaymentStatusLine, len(details))
	for i, d := range details {
		lines[i] = PaymentStatusLine{
			GiftID: d.GiftID,
			Kind:   d.GiftKind,
			Amount: d.Amount,
			Label:  d.Label,
		}
	}
	return p, lines, nil
}

// HandleWebhook processa notificação do Mercado Pago via provider id.
func (uc *UseCase) HandleWebhook(ctx context.Context, providerID string) error {
	gwResult, err := uc.gateway.GetPaymentStatus(ctx, providerID)
	if err != nil {
		return fmt.Errorf("payment.HandleWebhook: get from provider: %w", err)
	}

	p, err := uc.paymentRepo.FindByProviderID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("payment.HandleWebhook: find payment: %w", err)
	}

	return uc.propagateWebhookStatus(ctx, p, gwResult.Status)
}

func (uc *UseCase) HandleInfinitePayWebhook(ctx context.Context, orderNSU, invoiceSlug string, paid bool) error {
	var p *entity.Payment
	var err error
	if orderNSU != "" {
		p, err = uc.paymentRepo.FindByIDAny(ctx, orderNSU)
	}
	if err != nil && invoiceSlug != "" {
		p, err = uc.paymentRepo.FindByProviderID(ctx, invoiceSlug)
	}
	if err != nil {
		return fmt.Errorf("payment.HandleInfinitePayWebhook: find payment: %w", err)
	}

	var gwStatus string
	if paid {
		gwStatus = "approved"
	} else {
		gwStatus = "pending"
	}
	return uc.propagateWebhookStatus(ctx, p, gwStatus)
}

func (uc *UseCase) propagateWebhookStatus(ctx context.Context, p *entity.Payment, gwStatusRaw string) error {
	items, err := uc.paymentRepo.FindItemsByPaymentID(ctx, p.WeddingID, p.ID)
	if err != nil {
		return fmt.Errorf("payment.propagateWebhookStatus: items: %w", err)
	}

	newStatus := mapProviderStatus(gwStatusRaw)
	if p.Status == newStatus {
		return nil
	}

	old := p.Status
	now := time.Now()
	p.Status = newStatus
	p.UpdatedAt = now

	switch newStatus {
	case entity.PaymentStatusApproved:
		if p.PaidAt == nil {
			t := now
			p.PaidAt = &t
		}
		if old != entity.PaymentStatusApproved {
			_ = uc.markCatalogPurchased(ctx, p.WeddingID, items)
		}
	case entity.PaymentStatusExpired, entity.PaymentStatusRejected:
		if old == entity.PaymentStatusApproved {
			uc.revertCatalogPurchases(ctx, p.WeddingID, items)
		}
	}

	if err := uc.paymentRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("payment.propagateWebhookStatus: update: %w", err)
	}

	slog.Info("payment.webhook updated", "payment_id", p.ID, "status", newStatus)
	return nil
}

func (uc *UseCase) List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error) {
	return uc.paymentRepo.List(ctx, weddingID, page, perPage, status, giftID)
}

func (uc *UseCase) FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error) {
	return uc.paymentRepo.FindByID(ctx, weddingID, id)
}

func (uc *UseCase) FindPaymentLines(ctx context.Context, weddingID, paymentID string) ([]entity.PaymentItem, error) {
	return uc.paymentRepo.FindItemsByPaymentID(ctx, weddingID, paymentID)
}

// PaymentLinesForAdmin monta linhas para exibir no painel / integrações.
func (uc *UseCase) PaymentLinesForAdmin(ctx context.Context, weddingID, paymentID string) ([]PaymentDetailLine, error) {
	return uc.paymentLinesDetail(ctx, weddingID, paymentID)
}

func mapProviderStatus(s string) entity.PaymentStatus {
	switch s {
	case "approved":
		return entity.PaymentStatusApproved
	case "rejected", "cancelled", "refunded", "charged_back":
		return entity.PaymentStatusRejected
	case "expired":
		return entity.PaymentStatusExpired
	default:
		return entity.PaymentStatusPending
	}
}
