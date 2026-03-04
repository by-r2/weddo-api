package handler

import (
	"net/http"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/dto"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/middleware"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/guest"
)

type DashboardHandler struct {
	guestUC *guest.UseCase
}

func NewDashboardHandler(guestUC *guest.UseCase) *DashboardHandler {
	return &DashboardHandler{guestUC: guestUC}
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	stats, err := h.guestUC.Dashboard(r.Context(), weddingID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao buscar estatísticas.")
		return
	}

	var rate float64
	if stats.TotalGuests > 0 {
		rate = float64(stats.Confirmed) / float64(stats.TotalGuests) * 100
	}

	respondJSON(w, http.StatusOK, dto.DashboardResponse{
		RSVP: dto.RSVPStats{
			TotalInvitations: stats.TotalInvitations,
			TotalGuests:      stats.TotalGuests,
			Confirmed:        stats.Confirmed,
			Pending:          stats.Pending,
			Declined:         stats.Declined,
			ConfirmationRate: rate,
		},
	})
}
