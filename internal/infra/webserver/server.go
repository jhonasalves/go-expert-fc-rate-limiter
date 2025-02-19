package webserver

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/configs"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/infra/handlers"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/pkg/ratelimiter"

	mdLimiter "github.com/jhonasalves/go-expert-fc-rate-limiter/internal/infra/webserver/middleware"
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

	limiter := ratelimiter.NewRateLimiter(configs.RateLimiterMaxIPRequests, time.Minute)

	rl := mdLimiter.NewRateLimiterMiddleware(limiter)

	r.Handle("/", rl.Handler(http.HandlerFunc(handlers.HomeHandler)))

	return &Server{Router: r}
}
