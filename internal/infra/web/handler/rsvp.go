package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	"github.com/by-r2/weddo-api/internal/usecase/rsvp"
)

type RSVPHandler struct {
	rsvpUC *rsvp.UseCase
}

func NewRSVPHandler(uc *rsvp.UseCase) *RSVPHandler {
	return &RSVPHandler{rsvpUC: uc}
}

func (h *RSVPHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.RSVPRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Informe code e name.")
		return
	}

	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	guest, inv, alreadyConfirmed, err := h.rsvpUC.Confirm(r.Context(), weddingID, code, name)
	if err != nil {
		switch {
		case errors.Is(err, rsvp.ErrInvitationNotFound):
			respondError(w, http.StatusNotFound, "Convite não encontrado.")
			return
		case errors.Is(err, rsvp.ErrGuestNotFoundOnInvitation):
			respondError(w, http.StatusNotFound, "Convidado não encontrado neste convite. Verifique o nome.")
			return
		case errors.Is(err, rsvp.ErrGuestStatusTransitionNotAllowed):
			respondError(w, http.StatusConflict, "Este convidado recusou o convite e não pode confirmar novamente.")
			return
		}
		respondInternalError(w, r, "rsvp.handler.Confirm", err, "Erro interno do servidor.")
		return
	}

	msg := "Presença confirmada com sucesso!"
	status := http.StatusOK
	if alreadyConfirmed {
		msg = "Presença já estava confirmada."
		status = http.StatusConflict
	}

	respondJSON(w, status, dto.RSVPResponse{
		Guest:      toGuestSummary(guest),
		Invitation: dto.InvitationSummary{Label: inv.Label},
		Message:    msg,
	})
}

func (h *RSVPHandler) GetInvitationByCode(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		respondError(w, http.StatusBadRequest, "Parâmetro 'code' é obrigatório.")
		return
	}

	inv, guests, err := h.rsvpUC.FindInvitationByCode(r.Context(), weddingID, code)
	if err != nil {
		if errors.Is(err, rsvp.ErrInvitationNotFound) {
			respondError(w, http.StatusNotFound, "Convite não encontrado.")
			return
		}
		respondInternalError(w, r, "rsvp.handler.GetInvitationByCode", err, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toRSVPInvitationResponse(inv, guests))
}

func toRSVPInvitationResponse(inv *entity.Invitation, guests []entity.Guest) dto.RSVPInvitationResponse {
	guestsPublic := make([]dto.GuestPublic, len(guests))
	for i, g := range guests {
		guestsPublic[i] = dto.GuestPublic{ID: g.ID, Name: g.Name, Status: string(g.Status)}
	}
	return dto.RSVPInvitationResponse{
		Invitation: dto.InvitationSummary{Label: inv.Label, MaxGuests: inv.MaxGuests},
		Guests:     guestsPublic,
	}
}
