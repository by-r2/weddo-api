package web

import (
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/handler"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/middleware"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/guest"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/invitation"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/rsvp"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/wedding"
)

type RouterDeps struct {
	WeddingUC    *wedding.UseCase
	RSVPUC       *rsvp.UseCase
	InvitationUC *invitation.UseCase
	GuestUC      *guest.UseCase
	WeddingRepo  repository.WeddingRepository
	JWTSecret    string
	CORSOrigins  string
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
	rsvpHandler := handler.NewRSVPHandler(deps.RSVPUC)
	invHandler := handler.NewInvitationHandler(deps.InvitationUC)
	guestHandler := handler.NewGuestHandler(deps.GuestUC)
	dashHandler := handler.NewDashboardHandler(deps.GuestUC)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.Health)

		// Endpoints públicos (tenant via UUID na URL)
		r.Route("/w/{weddingId}", func(r chi.Router) {
			r.Use(middleware.TenantResolver(deps.WeddingRepo))

			r.Post("/rsvp", rsvpHandler.Confirm)
			r.Get("/rsvp/invitation", rsvpHandler.LookupInvitation)
		})

		// Autenticação
		r.Post("/admin/auth", authHandler.Login)

		// Endpoints admin (tenant via JWT)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))

			r.Get("/dashboard", dashHandler.Get)

			r.Route("/invitations", func(r chi.Router) {
				r.Get("/", invHandler.List)
				r.Post("/", invHandler.Create)
				r.Get("/{id}", invHandler.GetByID)
				r.Put("/{id}", invHandler.Update)
				r.Delete("/{id}", invHandler.Delete)
				r.Post("/{id}/guests", invHandler.AddGuest)
			})

			r.Route("/guests", func(r chi.Router) {
				r.Get("/", guestHandler.List)
				r.Get("/{id}", guestHandler.GetByID)
				r.Put("/{id}", guestHandler.Update)
				r.Delete("/{id}", guestHandler.Delete)
			})
		})

		// Webhook (sem auth — validação via assinatura do provider)
		// r.Post("/payments/webhook", paymentHandler.Webhook) — Fase 3
	})

	return r
}
