package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "subscription-aggregator/docs"
)

// NewRouter собирает HTTP-маршруты API.
func NewRouter(logger *slog.Logger, svc SubscriptionService) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(slogMiddleware(logger))

	r.Get("/health", Health)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	subscriptions := NewSubscriptionsHandler(svc)
	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Get("/cost", subscriptions.Cost)
		r.Post("/", subscriptions.Create)
		r.Get("/", subscriptions.List)
		r.Get("/{id}", subscriptions.Get)
		r.Patch("/{id}", subscriptions.Update)
		r.Delete("/{id}", subscriptions.Delete)
	})

	return r
}

func slogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("запрос",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", middleware.GetReqID(r.Context()),
			)
			next.ServeHTTP(w, r)
		})
	}
}
