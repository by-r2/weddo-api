package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerPort  int    `envconfig:"SERVER_PORT" default:"8080"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	JWTSecret          string `envconfig:"JWT_SECRET" required:"true"`
	JWTExpirationHours int    `envconfig:"JWT_EXPIRATION_HOURS" default:"24"`

	SeedWeddingSlug   string `envconfig:"SEED_WEDDING_SLUG" default:""`
	SeedWeddingTitle  string `envconfig:"SEED_WEDDING_TITLE" default:""`
	SeedWeddingDate   string `envconfig:"SEED_WEDDING_DATE" default:""`
	SeedAdminEmail    string `envconfig:"SEED_ADMIN_EMAIL" default:""`
	SeedAdminPassword string `envconfig:"SEED_ADMIN_PASSWORD" default:""`

	PaymentProvider string `envconfig:"PAYMENT_PROVIDER" default:""`

	MPAccessToken      string `envconfig:"MP_ACCESS_TOKEN" default:""`
	MPWebhookSecret    string `envconfig:"MP_WEBHOOK_SECRET" default:""`
	MPNotificationURL  string `envconfig:"MP_NOTIFICATION_URL" default:""`
	MPPixExpirationMin int    `envconfig:"MP_PIX_EXPIRATION_MINUTES" default:"30"`

	IPHandle      string `envconfig:"IP_HANDLE" default:""`
	IPRedirectURL string `envconfig:"IP_REDIRECT_URL" default:""`
	IPWebhookURL  string `envconfig:"IP_WEBHOOK_URL" default:""`

	GoogleOAuthClientID       string `envconfig:"GOOGLE_OAUTH_CLIENT_ID" default:""`
	GoogleOAuthClientSecret   string `envconfig:"GOOGLE_OAUTH_CLIENT_SECRET" default:""`
	GoogleOAuthRedirectURL    string `envconfig:"GOOGLE_OAUTH_REDIRECT_URL" default:""`
	GoogleOAuthTokenCipherKey string `envconfig:"GOOGLE_OAUTH_TOKEN_CIPHER_KEY" default:""`
	GoogleOAuthStateSecret    string `envconfig:"GOOGLE_OAUTH_STATE_SECRET" default:""`

	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"*"`
	LogLevel           string `envconfig:"LOG_LEVEL" default:"info"`
	LogFormat          string `envconfig:"LOG_FORMAT" default:"text"`
	InvitationCodeLen  int    `envconfig:"INVITATION_CODE_LENGTH" default:"5"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}
	return &cfg, nil
}
