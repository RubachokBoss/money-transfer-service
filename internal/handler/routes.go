package handler

import (
	"github.com/go-chi/chi"
	"money-transfer-service/internal/middleware"
	"money-transfer-service/internal/repository"
)

func (h *Handler) Routes(authHandler *AuthHandler, repo *repository.Repository) chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(repo))

		r.Get("/api/balance", h.GetBalance)
		r.Post("/api/transfer", h.TransferMoney)
		r.Get("/api/transfers", h.GetTransfersHistory)
	})

	return r
}
