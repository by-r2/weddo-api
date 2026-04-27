package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		}

		// Convenção alinhada ao ecossistema Go (ex.: go-chi/httplog, docs Gin): 5xx = erro do servidor,
		// 4xx = problema de cliente; sucesso/redirects em Debug para não encher o log com LOG_LEVEL=info.
		switch {
		case ww.status >= 500:
			slog.Error("request", attrs...)
		case ww.status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Debug("request", attrs...)
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
