package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	"github.com/by-r2/weddo-api/internal/usecase/user"
)

type UserHandler struct {
	userUC *user.UseCase
}

func NewUserHandler(uc *user.UseCase) *UserHandler {
	return &UserHandler{userUC: uc}
}

func (h *UserHandler) Invite(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.InviteUserRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Informe um email válido.")
		return
	}

	u, err := h.userUC.Invite(r.Context(), weddingID, req.Email, req.Name)
	if err != nil {
		if errors.Is(err, entity.ErrAlreadyExists) {
			respondError(w, http.StatusConflict, "Este email já está vinculado a um casamento.")
			return
		}
		respondInternalError(w, r, "user.handler.Invite", err, "Erro ao convidar usuário.")
		return
	}

	respondJSON(w, http.StatusCreated, toUserResponse(u))
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	users, err := h.userUC.List(r.Context(), weddingID)
	if err != nil {
		respondInternalError(w, r, "user.handler.List", err, "Erro ao listar usuários.")
		return
	}

	items := make([]dto.UserResponse, len(users))
	for i, u := range users {
		items[i] = toUserResponse(&u)
	}

	respondJSON(w, http.StatusOK, items)
}

func (h *UserHandler) Remove(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	userID := chi.URLParam(r, "id")

	if err := h.userUC.Remove(r.Context(), weddingID, userID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Usuário não encontrado.")
			return
		}
		if errors.Is(err, entity.ErrUnauthorized) {
			respondError(w, http.StatusForbidden, "Não é possível remover o último administrador do casamento.")
			return
		}
		respondInternalError(w, r, "user.handler.Remove", err, "Erro ao remover usuário.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toUserResponse(u *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		HasGoogle: u.GoogleID != "",
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
