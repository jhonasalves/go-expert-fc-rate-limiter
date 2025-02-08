package webserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/configs"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/handlers"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/middleware/ratelimiter"
	"golang.org/x/time/rate"
)

type Server struct {
	Router *chi.Mux
}

func NewServer() *Server {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	rl := ratelimiter.NewRateLimiter(rate.Limit(configs.RateLimiterMaxIPRequests), configs.RateLimiterBurst)
	r.Use(rl.RateLimitMiddleware)

	r.Get("/", handlers.HomeHandler)

	return &Server{Router: r}
}
