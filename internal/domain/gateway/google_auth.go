package gateway

import "context"

// GoogleTokenInfo contém os dados extraídos de um Google ID Token verificado.
type GoogleTokenInfo struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

// GoogleAuthVerifier valida um Google ID Token e extrai informações do usuário.
type GoogleAuthVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleTokenInfo, error)
}
