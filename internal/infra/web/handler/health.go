package handler

import (
	"net/http"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/dto"
)

func Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, dto.HealthResponse{Status: "ok"})
}
