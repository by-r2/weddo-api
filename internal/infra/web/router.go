package web

import (
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/handler"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/middleware"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/wedding"
)

type RouterDeps struct {
	WeddingUC   *wedding.UseCase
	WeddingRepo repository.WeddingRepository
	JWTSecret   string
	CORSOrigins string
}

func NewRouter(deps RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	origins := strings.Split(deps.CORSOrigins, ",")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RealIP)

	authHandler := handler.NewAuthHandler(deps.WeddingUC)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.Health)

		// Endpoints públicos (requerem weddingId na URL)
		r.Route("/w/{weddingId}", func(r chi.Router) {
			r.Use(middleware.TenantResolver(deps.WeddingRepo))
			// RSVP e gifts serão registrados aqui nas próximas fases
		})

		// Autenticação
		r.Post("/admin/auth", authHandler.Login)

		// Endpoints admin (requerem JWT)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			// CRUD de invitations, guests, gifts, payments serão registrados aqui
		})

		// Webhook (sem auth — validação via assinatura do provider)
		// r.Post("/payments/webhook", paymentHandler.Webhook) — Fase 3
	})

	return r
}
