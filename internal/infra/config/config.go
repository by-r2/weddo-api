package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerPort int    `envconfig:"SERVER_PORT" default:"8080"`
	DatabasePath string `envconfig:"DATABASE_PATH" default:"./data/wedding.db"`

	JWTSecret          string `envconfig:"JWT_SECRET" required:"true"`
	JWTExpirationHours int    `envconfig:"JWT_EXPIRATION_HOURS" default:"24"`

	SeedWeddingSlug  string `envconfig:"SEED_WEDDING_SLUG" default:""`
	SeedWeddingTitle string `envconfig:"SEED_WEDDING_TITLE" default:""`
	SeedWeddingDate  string `envconfig:"SEED_WEDDING_DATE" default:""`
	SeedAdminEmail   string `envconfig:"SEED_ADMIN_EMAIL" default:""`
	SeedAdminPassword string `envconfig:"SEED_ADMIN_PASSWORD" default:""`

	MPAccessToken       string `envconfig:"MP_ACCESS_TOKEN" default:""`
	MPWebhookSecret     string `envconfig:"MP_WEBHOOK_SECRET" default:""`
	MPNotificationURL   string `envconfig:"MP_NOTIFICATION_URL" default:""`
	MPPixExpirationMin  int    `envconfig:"MP_PIX_EXPIRATION_MINUTES" default:"30"`

	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"*"`
	LogLevel           string `envconfig:"LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}
	return &cfg, nil
}
