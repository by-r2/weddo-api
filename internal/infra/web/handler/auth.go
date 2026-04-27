package handler

import (
	"errors"
	"net/http"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/gateway"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/usecase/wedding"
)

type AuthHandler struct {
	weddingUC    *wedding.UseCase
	googleVerify gateway.GoogleAuthVerifier
}

func NewAuthHandler(weddingUC *wedding.UseCase, gv gateway.GoogleAuthVerifier) *AuthHandler {
	return &AuthHandler{weddingUC: weddingUC, googleVerify: gv}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique os campos obrigatórios.")
		return
	}

	result, err := h.weddingUC.Register(r.Context(), wedding.RegisterInput{
		Partner1Name: req.Partner1Name,
		Partner2Name: req.Partner2Name,
		Email:        req.Email,
		Password:     req.Password,
		Date:         req.Date,
		Slug:         req.Slug,
	})
	if err != nil {
		if errors.Is(err, entity.ErrAlreadyExists) {
			respondError(w, http.StatusConflict, "Já existe uma conta cadastrada com este email.")
			return
		}
		respondInternalError(w, r, "auth.handler.Register", err, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusCreated, toAuthResponse(result))
}

func (h *AuthHandler) RegisterGoogle(w http.ResponseWriter, r *http.Request) {
	if h.googleVerify == nil {
		respondError(w, http.StatusServiceUnavailable, "Login com Google não está configurado.")
		return
	}

	var req dto.RegisterGoogleRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique os campos obrigatórios.")
		return
	}

	info, err := h.googleVerify.Verify(r.Context(), req.IDToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token do Google inválido.")
		return
	}

	result, err := h.weddingUC.RegisterGoogle(r.Context(), wedding.RegisterGoogleInput{
		Partner1Name: req.Partner1Name,
		Partner2Name: req.Partner2Name,
		Date:         req.Date,
		Slug:         req.Slug,
		GoogleID:     info.GoogleID,
		Email:        info.Email,
		Name:         info.Name,
		Picture:      info.Picture,
	})
	if err != nil {
		if errors.Is(err, entity.ErrAlreadyExists) {
			respondError(w, http.StatusConflict, "Já existe uma conta cadastrada com este email ou conta Google.")
			return
		}
		respondInternalError(w, r, "auth.handler.RegisterGoogle", err, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusCreated, toAuthResponse(result))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique email e senha.")
		return
	}

	result, err := h.weddingUC.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, entity.ErrUnauthorized) {
			respondError(w, http.StatusUnauthorized, "Email ou senha incorretos.")
			return
		}
		respondInternalError(w, r, "auth.handler.Login", err, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toAuthResponse(result))
}

func (h *AuthHandler) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	if h.googleVerify == nil {
		respondError(w, http.StatusServiceUnavailable, "Login com Google não está configurado.")
		return
	}

	var req dto.GoogleAuthRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	info, err := h.googleVerify.Verify(r.Context(), req.IDToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token do Google inválido.")
		return
	}

	result, err := h.weddingUC.AuthenticateGoogle(r.Context(), info.GoogleID, info.Email, info.Name, info.Picture)
	if err != nil {
		if errors.Is(err, entity.ErrUnauthorized) {
			respondError(w, http.StatusUnauthorized, "Conta não encontrada. Registre-se primeiro ou peça um convite ao administrador do casamento.")
			return
		}
		respondInternalError(w, r, "auth.handler.LoginGoogle", err, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toAuthResponse(result))
}

func toAuthResponse(r *wedding.AuthResult) dto.AuthResponse {
	return dto.AuthResponse{
		Token: r.Token,
		Wedding: dto.WeddingSummary{
			ID:    r.Wedding.ID,
			Slug:  r.Wedding.Slug,
			Title: r.Wedding.Title,
		},
		User: dto.UserSummary{
			ID:        r.User.ID,
			Name:      r.User.Name,
			Email:     r.User.Email,
			AvatarURL: r.User.AvatarURL,
		},
	}
}
