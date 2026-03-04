package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const WeddingIDKey contextKey = "wedding_id"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				unauthorized(w, "Token não fornecido")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				unauthorized(w, "Formato de token inválido")
				return
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				unauthorized(w, "Token inválido ou expirado")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				unauthorized(w, "Token inválido")
				return
			}

			weddingID, ok := claims["wedding_id"].(string)
			if !ok || weddingID == "" {
				unauthorized(w, "Token sem identificação do casamento")
				return
			}

			ctx := context.WithValue(r.Context(), WeddingIDKey, weddingID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetWeddingID(ctx context.Context) string {
	id, _ := ctx.Value(WeddingIDKey).(string)
	return id
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
