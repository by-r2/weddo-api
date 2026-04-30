package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	giftuc "github.com/by-r2/weddo-api/internal/usecase/gift"
	"github.com/go-chi/chi/v5"
)

var (
	errGiftListInvalidSortBy   = errors.New("Parâmetro sort_by inválido. Use recommended, price ou name.")
	errGiftListInvalidSortDir  = errors.New("Parâmetro sort_dir inválido. Use asc ou desc.")
	errGiftListInvalidMinPrice = errors.New("Parâmetro min_price inválido.")
	errGiftListInvalidMaxPrice = errors.New("Parâmetro max_price inválido.")
	errGiftListPriceRange      = errors.New("O valor mínimo de preço não pode ser maior que o máximo.")
)

type GiftHandler struct {
	giftUC *giftuc.UseCase
}

func NewGiftHandler(uc *giftuc.UseCase) *GiftHandler {
	return &GiftHandler{giftUC: uc}
}

func parseGiftListQuery(r *http.Request, page, perPage int) (repository.GiftListParams, error) {
	q := r.URL.Query()
	p := repository.GiftListParams{
		Page:    page,
		PerPage: perPage,
		Search:  strings.TrimSpace(q.Get("search")),
	}

	seenCat := make(map[string]struct{})
	for _, c := range q["category"] {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		key := strings.ToLower(c)
		if _, ok := seenCat[key]; ok {
			continue
		}
		seenCat[key] = struct{}{}
		p.Categories = append(p.Categories, c)
	}

	sortBy := strings.TrimSpace(strings.ToLower(q.Get("sort_by")))
	if sortBy == "" {
		sortBy = "recommended"
	}
	switch sortBy {
	case "recommended", "price", "name":
		p.SortBy = sortBy
	default:
		return p, errGiftListInvalidSortBy
	}

	sortDir := strings.TrimSpace(strings.ToLower(q.Get("sort_dir")))
	if sortDir == "" {
		sortDir = "asc"
	}
	if sortDir != "asc" && sortDir != "desc" {
		return p, errGiftListInvalidSortDir
	}
	p.SortDir = sortDir

	if s := q.Get("min_price"); s != "" {
		v, scanErr := strconv.ParseFloat(s, 64)
		if scanErr != nil || v < 0 {
			return p, errGiftListInvalidMinPrice
		}
		p.MinPrice = &v
	}
	if s := q.Get("max_price"); s != "" {
		v, scanErr := strconv.ParseFloat(s, 64)
		if scanErr != nil || v < 0 {
			return p, errGiftListInvalidMaxPrice
		}
		p.MaxPrice = &v
	}
	if p.MinPrice != nil && p.MaxPrice != nil && *p.MinPrice > *p.MaxPrice {
		return p, errGiftListPriceRange
	}

	return p, nil
}

// ListPublic lista só presentes de catálogo; o modelo de contribuição em dinheiro (cash_template)
// não aparece aqui — doações ficam ligadas aos pagamentos/linhas, não ao catálogo.
func (h *GiftHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	params, err := parseGiftListQuery(r, page, perPage)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	params.Status = string(entity.GiftStatusAvailable)
	params.CatalogOnly = true

	gifts, total, err := h.giftUC.List(r.Context(), weddingID, params)
	if err != nil {
		respondInternalError(w, r, "gift.handler.ListPublic", err, "Erro ao listar presentes.")
		return
	}

	items := make([]dto.GiftResponse, len(gifts))
	for i, g := range gifts {
		items[i] = toGiftResponse(&g)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

// ListGiftCategories retorna categorias distintas já usadas em presentes de catálogo (público — tenant na URL).
func (h *GiftHandler) ListGiftCategories(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	cats, err := h.giftUC.ListCategories(r.Context(), weddingID)
	if err != nil {
		respondInternalError(w, r, "gift.handler.ListGiftCategories", err, "Erro ao listar categorias.")
		return
	}
	items := make([]dto.GiftCategoryItem, len(cats))
	for i, c := range cats {
		items[i] = dto.GiftCategoryItem{
			Name:  c.Category,
			Count: c.Count,
		}
	}
	respondJSON(w, http.StatusOK, dto.GiftCategoriesResponse{Categories: items})
}

func (h *GiftHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	g, err := h.giftUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondInternalError(w, r, "gift.handler.GetPublic", err, "Erro interno do servidor.")
		return
	}
	if g.Kind == entity.GiftKindCashTemplate {
		respondError(w, http.StatusNotFound, "Presente não encontrado.")
		return
	}

	respondJSON(w, http.StatusOK, toGiftResponse(g))
}

func (h *GiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.CreateGiftRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique os campos obrigatórios.")
		return
	}

	g, err := h.giftUC.Create(r.Context(), giftuc.CreateInput{
		WeddingID:   weddingID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	})
	if err != nil {
		respondInternalError(w, r, "gift.handler.Create", err, "Erro ao criar presente.")
		return
	}

	respondJSON(w, http.StatusCreated, toGiftResponse(g))
}

// List (admin): mesmo conjunto da lista “geral” de presentes — só catálogo, sem cash_template.
func (h *GiftHandler) List(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	params, err := parseGiftListQuery(r, page, perPage)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	params.Status = strings.TrimSpace(r.URL.Query().Get("status"))
	params.CatalogOnly = true

	gifts, total, err := h.giftUC.List(r.Context(), weddingID, params)
	if err != nil {
		respondInternalError(w, r, "gift.handler.List", err, "Erro ao listar presentes.")
		return
	}

	items := make([]dto.GiftResponse, len(gifts))
	for i, g := range gifts {
		items[i] = toGiftResponse(&g)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *GiftHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	g, err := h.giftUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondInternalError(w, r, "gift.handler.GetByID", err, "Erro interno do servidor.")
		return
	}
	if g.Kind == entity.GiftKindCashTemplate {
		respondError(w, http.StatusNotFound, "Presente não encontrado.")
		return
	}

	respondJSON(w, http.StatusOK, toGiftResponse(g))
}

func (h *GiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	var req dto.UpdateGiftRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	prev, err := h.giftUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondInternalError(w, r, "gift.handler.Update", err, "Erro ao buscar presente.")
		return
	}
	if prev.Kind == entity.GiftKindCashTemplate {
		respondError(w, http.StatusForbidden, "Este item não pode ser alterado pelo painel.")
		return
	}

	g, err := h.giftUC.Update(r.Context(), giftuc.UpdateInput{
		WeddingID:   weddingID,
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondInternalError(w, r, "gift.handler.Update", err, "Erro ao atualizar presente.")
		return
	}

	respondJSON(w, http.StatusOK, toGiftResponse(g))
}

func (h *GiftHandler) Delete(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	prev, err := h.giftUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondInternalError(w, r, "gift.handler.Delete", err, "Erro ao buscar presente.")
		return
	}
	if prev.Kind == entity.GiftKindCashTemplate {
		respondError(w, http.StatusForbidden, "Este item não pode ser removido pelo painel.")
		return
	}

	if err := h.giftUC.Delete(r.Context(), weddingID, id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondInternalError(w, r, "gift.handler.Delete", err, "Erro ao remover presente.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toGiftResponse(g *entity.Gift) dto.GiftResponse {
	kind := string(g.Kind)
	if g.Kind == "" {
		kind = string(entity.GiftKindCatalog)
	}
	resp := dto.GiftResponse{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Price:       g.Price,
		ImageURL:    g.ImageURL,
		Category:    g.Category,
		Status:      string(g.Status),
		CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   g.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if kind != string(entity.GiftKindCatalog) {
		resp.Kind = kind
	}
	return resp
}
