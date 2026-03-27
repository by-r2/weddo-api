package google

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"

	"github.com/by-r2/weddo-api/internal/domain/gateway"
)

type authVerifier struct {
	clientID string
}

func NewAuthVerifier(clientID string) gateway.GoogleAuthVerifier {
	return &authVerifier{clientID: clientID}
}

func (v *authVerifier) Verify(ctx context.Context, rawToken string) (*gateway.GoogleTokenInfo, error) {
	payload, err := idtoken.Validate(ctx, rawToken, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("google.Verify: %w", err)
	}

	info := &gateway.GoogleTokenInfo{
		GoogleID: fmt.Sprintf("%v", payload.Subject),
	}

	if email, ok := payload.Claims["email"].(string); ok {
		info.Email = email
	}
	if name, ok := payload.Claims["name"].(string); ok {
		info.Name = name
	}
	if picture, ok := payload.Claims["picture"].(string); ok {
		info.Picture = picture
	}

	return info, nil
}
