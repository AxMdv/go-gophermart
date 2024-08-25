package router

import (
	"github.com/AxMdv/go-gophermart/internal/handlers"
	mw "github.com/AxMdv/go-gophermart/internal/middleware"
	"github.com/go-chi/chi"
)

func New(h *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.RegisterUser)
		r.Post("/login", h.LoginUser)
		r.Post("/orders", mw.ValidateUserMiddleware(mw.GzipMiddleware((h.CreateOrder))))
		r.Get("/orders", mw.ValidateUserMiddleware(mw.GzipMiddleware((h.GetOrdersInfo))))
		r.Get("/withdrawals", h.GetWithdrawalsInfo)

		r.Route("/balance", func(r chi.Router) {
			r.Get("/", h.GetUserBalance)
			r.Post("/withdraw", h.CreateWithdraw)
		})
	})
	// r.Get("/api/orders/{number}", h.Asdas)
	return r
}
