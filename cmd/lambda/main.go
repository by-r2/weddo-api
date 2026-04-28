package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/by-r2/weddo-api/internal/domain/gateway"
	"github.com/by-r2/weddo-api/internal/infra/config"
	"github.com/by-r2/weddo-api/internal/infra/database"
	infragateway "github.com/by-r2/weddo-api/internal/infra/gateway"
	infraGoogle "github.com/by-r2/weddo-api/internal/infra/google"
	"github.com/by-r2/weddo-api/internal/infra/security"
	infraSheets "github.com/by-r2/weddo-api/internal/infra/sheets"
	"github.com/by-r2/weddo-api/internal/infra/web"
	giftuc "github.com/by-r2/weddo-api/internal/usecase/gift"
	"github.com/by-r2/weddo-api/internal/usecase/guest"
	"github.com/by-r2/weddo-api/internal/usecase/invitation"
	paymentuc "github.com/by-r2/weddo-api/internal/usecase/payment"
	"github.com/by-r2/weddo-api/internal/usecase/rsvp"
	sheetsuc "github.com/by-r2/weddo-api/internal/usecase/sheets"
	"github.com/by-r2/weddo-api/internal/usecase/user"
	"github.com/by-r2/weddo-api/internal/usecase/wedding"
)

var proxyHandler *httpadapter.HandlerAdapterV2

func init() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel, cfg.LogFormat)

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	// Limita conexões por instância Lambda para não esgotar db.t4g.micro (~112 max_connections).
	// Com ReservedConcurrency=10 e MaxOpenConns=5: máximo 50 conexões simultâneas.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := database.RunMigrations(db, "migrations"); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	}

	proxyHandler = httpadapter.NewV2(buildRouter(cfg, db))
	slog.Info("lambda initialized")
}

func main() {
	lambda.Start(proxyHandler.ProxyWithContext)
}

func buildRouter(cfg *config.Config, db *sql.DB) http.Handler {
	weddingRepo := database.NewWeddingRepository(db)
	invitationRepo := database.NewInvitationRepository(db)
	guestRepo := database.NewGuestRepository(db)
	giftRepo := database.NewGiftRepository(db)
	paymentRepo := database.NewPaymentRepository(db)
	googleIntegrationRepo := database.NewGoogleIntegrationRepository(db)
	userRepo := database.NewUserRepository(db)

	ensureCashTpl := func(ctx context.Context, weddingID string) error {
		return giftuc.EnsureCashTemplate(ctx, giftRepo, weddingID)
	}
	weddingUC := wedding.NewUseCase(weddingRepo, userRepo, cfg.JWTSecret, cfg.JWTExpirationHours, ensureCashTpl)
	rsvpUC := rsvp.NewUseCase(guestRepo, invitationRepo)
	invitationUC := invitation.NewUseCase(invitationRepo, guestRepo)
	guestUC := guest.NewUseCase(guestRepo, invitationRepo)
	giftUC := giftuc.NewUseCase(giftRepo, paymentRepo)
	userUC := user.NewUseCase(userRepo)

	var paymentUC *paymentuc.UseCase
	switch strings.ToLower(cfg.PaymentProvider) {
	case "infinitepay":
		ipGateway := infragateway.NewInfinitePayGateway(cfg.IPHandle, cfg.IPRedirectURL, cfg.IPWebhookURL)
		paymentUC = paymentuc.NewUseCase(paymentRepo, giftRepo, ipGateway)
	case "mercadopago":
		mpGateway, err := infragateway.NewMercadoPagoGateway(
			cfg.MPAccessToken, cfg.MPNotificationURL, cfg.MPPixExpirationMin,
		)
		if err != nil {
			slog.Error("failed to init mercado pago gateway", "error", err)
			os.Exit(1)
		}
		paymentUC = paymentuc.NewUseCase(paymentRepo, giftRepo, mpGateway)
	}

	var googleVerifier gateway.GoogleAuthVerifier
	if cfg.GoogleOAuthClientID != "" {
		googleVerifier = infraGoogle.NewAuthVerifier(cfg.GoogleOAuthClientID)
	}

	var sheetsUC *sheetsuc.UseCase
	if cfg.GoogleOAuthClientID != "" && cfg.GoogleOAuthClientSecret != "" &&
		cfg.GoogleOAuthRedirectURL != "" && cfg.GoogleOAuthTokenCipherKey != "" {
		oauthProvider := infraSheets.NewOAuthProvider(
			cfg.GoogleOAuthClientID, cfg.GoogleOAuthClientSecret, cfg.GoogleOAuthRedirectURL,
		)
		tokenCipher, err := security.NewCipher(cfg.GoogleOAuthTokenCipherKey)
		if err != nil {
			slog.Error("failed to init token cipher", "error", err)
			os.Exit(1)
		}
		stateSecret := cfg.GoogleOAuthStateSecret
		if stateSecret == "" {
			stateSecret = cfg.JWTSecret
		}
		sheetsUC = sheetsuc.NewUseCase(
			invitationRepo, guestRepo, giftRepo, paymentRepo,
			weddingRepo, googleIntegrationRepo,
			oauthProvider, tokenCipher, stateSecret,
		)
	}

	return web.NewRouter(web.RouterDeps{
		WeddingUC:      weddingUC,
		RSVPUC:         rsvpUC,
		InvitationUC:   invitationUC,
		GuestUC:        guestUC,
		GiftUC:         giftUC,
		PaymentUC:      paymentUC,
		SheetsUC:       sheetsUC,
		UserUC:         userUC,
		WeddingRepo:    weddingRepo,
		GoogleVerifier: googleVerifier,
		JWTSecret:      cfg.JWTSecret,
		CORSOrigins:    cfg.CORSAllowedOrigins,
	})
}

func setupLogger(level, format string) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
