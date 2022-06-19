package handlers

import "github.com/go-chi/chi/v5"

// GetRoutes объявим роуты, используя маршрутизатор chi
func (h *handler) setRoutes() {
	h.r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.Get("/logout", h.logout)
		r.Get("/welcome", h.welcome)

		r.Post("/orders", h.postOrders)
		r.Get("/orders", h.getOrders)

		r.Get("/balance", h.getBalance)
		r.Post("/balance/withdraw", h.postWithdraw)
		r.Get("/balance/withdrawals", h.getWithdrawals)
	})
}
