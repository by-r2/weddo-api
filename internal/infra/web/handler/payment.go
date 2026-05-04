package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	paymentuc "github.com/by-r2/weddo-api/internal/usecase/payment"
	"github.com/go-chi/chi/v5"
)

type PaymentHandler struct {
	paymentUC *paymentuc.UseCase
}

func NewPaymentHandler(uc *paymentuc.UseCase) *PaymentHandler {
	return &PaymentHandler{paymentUC: uc}
}

func (h *PaymentHandler) checkAvailable(w http.ResponseWriter) bool {
	if h.paymentUC == nil {
		respondError(w, http.StatusServiceUnavailable, "Pagamentos não configurados. Configure PAYMENT_PROVIDER no .env.")
		return false
	}
	return true
}

func (h *PaymentHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.CheckoutRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Informe lista de presentes e dados do pagador.")
		return
	}

	lines := make([]paymentuc.CheckoutLineInput, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, paymentuc.CheckoutLineInput{
			GiftID:            strings.TrimSpace(it.GiftID),
			Amount:            it.Amount,
			CustomName:        it.CustomName,
			CustomDescription: it.CustomDescription,
		})
	}

	result, err := h.paymentUC.Checkout(r.Context(), paymentuc.CheckoutInput{
		WeddingID:       weddingID,
		Lines:           lines,
		PayerName:       req.PayerName,
		PayerEmail:      req.PayerEmail,
		Message:         req.Message,
		PaymentMethod:   req.PaymentMethod,
		CardToken:       req.CardToken,
		PaymentMethodID: req.PaymentMethodID,
		Installments:    req.Installments,
		RedirectURL:     req.RedirectURL,
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
		case errors.Is(err, paymentuc.ErrGiftUnavailable):
			respondError(w, http.StatusConflict, "Um dos presentes não está mais disponível.")
		case errors.Is(err, paymentuc.ErrCheckoutEmptyItems):
			respondError(w, http.StatusBadRequest, "Informe pelo menos um item.")
		case errors.Is(err, paymentuc.ErrCheckoutDuplicateGiftLine):
			respondError(w, http.StatusBadRequest, "Linha repetida para o mesmo presente.")
		case errors.Is(err, paymentuc.ErrCheckoutInvalidCatalogExtras):
			respondError(w, http.StatusBadRequest, "Itens do catálogo não aceitam valor ou texto próprio.")
		case errors.Is(err, paymentuc.ErrCheckoutCashAmountMissing):
			respondError(w, http.StatusBadRequest, "Informe o valor da contribuição em dinheiro.")
		case errors.Is(err, paymentuc.ErrCashAmountOutOfRange):
			respondError(w, http.StatusBadRequest, "Valor inválido. Use entre R$ 1,00 e R$ 100.000,00.")
		default:
			if strings.Contains(err.Error(), "excede o limite") {
				respondError(w, http.StatusBadRequest, "Texto da contribuição excede o limite.")
				return
			}
			respondInternalError(w, r, "payment.handler.Checkout", err, "Erro ao processar pagamento.")
		}
		return
	}

	resp := dto.CheckoutResponse{
		PaymentID:  result.Payment.ID,
		ProviderID: result.Payment.ProviderID,
		Status:     string(result.Payment.Status),
	}

	if result.CheckoutURL != "" {
		resp.CheckoutURL = result.CheckoutURL
	}

	if result.QRCode != "" {
		resp.QRCode = result.QRCode
		resp.QRCodeBase64 = result.QRCodeBase64
		if result.Payment.PixExpiration != nil {
			s := result.Payment.PixExpiration.Format("2006-01-02T15:04:05Z")
			resp.ExpiresAt = &s
		}
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *PaymentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	p, lines, err := h.paymentUC.GetStatus(r.Context(), weddingID, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Pagamento não encontrado.")
			return
		}
		respondInternalError(w, r, "payment.handler.GetStatus", err, "Erro interno do servidor.")
		return
	}

	out := make([]dto.PaymentLineStatusDTO, 0, len(lines))
	for _, ln := range lines {
		out = append(out, dto.PaymentLineStatusDTO{
			GiftID: ln.GiftID,
			Kind:   string(ln.Kind),
			Amount: ln.Amount,
			Label:  ln.Label,
		})
	}

	respondJSON(w, http.StatusOK, dto.PaymentStatusResponse{
		PaymentID: p.ID,
		Status:    string(p.Status),
		Lines:     out,
	})
}

// --- Webhooks ---

// WebhookMercadoPago processa notificações do Mercado Pago.
func (h *PaymentHandler) WebhookMercadoPago(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}

	var payload struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Action != "payment.updated" && payload.Action != "payment.created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		slog.Error("webhook.mp: failed to parse data", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if data.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.paymentUC.HandleWebhook(r.Context(), data.ID); err != nil {
		slog.Error("webhook.mp: failed to handle", "provider_id", data.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// WebhookInfinitePay processa notificações da InfinitePay.
func (h *PaymentHandler) WebhookInfinitePay(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}

	var payload struct {
		InvoiceSlug    string `json:"invoice_slug"`
		Amount         int    `json:"amount"`
		PaidAmount     int    `json:"paid_amount"`
		Installments   int    `json:"installments"`
		CaptureMethod  string `json:"capture_method"`
		TransactionNSU string `json:"transaction_nsu"`
		OrderNSU       string `json:"order_nsu"`
		ReceiptURL     string `json:"receipt_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.OrderNSU == "" && payload.InvoiceSlug == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	paid := payload.PaidAmount > 0

	if err := h.paymentUC.HandleInfinitePayWebhook(r.Context(), payload.OrderNSU, payload.InvoiceSlug, paid); err != nil {
		slog.Error("webhook.ip: failed to handle", "order_nsu", payload.OrderNSU, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Webhook roteia para o provedor correto conforme PAYMENT_PROVIDER.
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}

	if h.paymentUC.ProviderName() == "infinitepay" {
		h.WebhookInfinitePay(w, r)
		return
	}
	h.WebhookMercadoPago(w, r)
}

func (h *PaymentHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	status := r.URL.Query().Get("status")
	giftID := r.URL.Query().Get("gift_id")

	payments, total, err := h.paymentUC.List(r.Context(), weddingID, page, perPage, status, giftID)
	if err != nil {
		respondInternalError(w, r, "payment.handler.ListAdmin", err, "Erro ao listar pagamentos.")
		return
	}

	items := make([]dto.PaymentResponse, len(payments))
	for i := range payments {
		dtoResp, err := h.paymentToDTO(r.Context(), weddingID, &payments[i])
		if err != nil {
			respondInternalError(w, r, "payment.handler.ListAdmin.lines", err, "Erro ao montar pagamento.")
			return
		}
		items[i] = dtoResp
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *PaymentHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	p, err := h.paymentUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Pagamento não encontrado.")
			return
		}
		respondInternalError(w, r, "payment.handler.GetAdmin", err, "Erro interno do servidor.")
		return
	}

	resp, err := h.paymentToDTO(r.Context(), weddingID, p)
	if err != nil {
		respondInternalError(w, r, "payment.handler.GetAdmin.dto", err, "Erro interno do servidor.")
		return
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) paymentToDTO(ctx context.Context, weddingID string, p *entity.Payment) (dto.PaymentResponse, error) {
	details, err := h.paymentUC.PaymentLinesForAdmin(ctx, weddingID, p.ID)
	if err != nil {
		return dto.PaymentResponse{}, err
	}
	items := make([]dto.PaymentLineResponse, 0, len(details))
	for _, d := range details {
		items = append(items, dto.PaymentLineResponse{
			GiftID:            d.GiftID,
			Kind:              string(d.GiftKind),
			Amount:            d.Amount,
			CustomName:        d.CustomName,
			CustomDescription: d.CustomDescription,
			Label:             d.Label,
		})
	}

	resp := dto.PaymentResponse{
		ID:            p.ID,
		ProviderID:    p.ProviderID,
		Amount:        p.Amount,
		Status:        string(p.Status),
		PaymentMethod: string(p.PaymentMethod),
		PayerName:     p.PayerName,
		PayerEmail:    p.PayerEmail,
		Message:       p.Message,
		Items:         items,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.PaidAt != nil {
		s := p.PaidAt.Format("2006-01-02T15:04:05Z")
		resp.PaidAt = &s
	}
	return resp, nil
}
