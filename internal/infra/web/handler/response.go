package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/by-r2/weddo-api/internal/dto"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("respondJSON: encode failed", "error", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, dto.ErrorResponse{Error: message})
}

// respondInternalError registra o erro para observabilidade (mensagens de log em inglês)
// e devolve ao cliente uma mensagem segura em português (sem detalhes internos).
func respondInternalError(w http.ResponseWriter, r *http.Request, logOperation string, err error, userMessage string) {
	slog.Error(logOperation,
		"error", err,
		"method", r.Method,
		"path", r.URL.Path,
	)
	respondError(w, http.StatusInternalServerError, userMessage)
}
