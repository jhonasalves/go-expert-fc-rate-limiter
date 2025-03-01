package webserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/configs"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/infra/database"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/infra/handlers"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/pkg/logger"
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

	logger := logger.NewLogger()

	redisDB := database.NewRedisDatabase(configs)
	storage := ratelimiter.NewRedisStorage(redisDB.Client, logger)
	limiter := ratelimiter.NewRateLimiter(
		storage,
		ratelimiter.Options{
			MaxRequestIP:    configs.RateLimiterMaxIPRequests,
			MaxRequestToken: configs.RateLimiterMaxTokenRequests,
			WindowDuration:  configs.RateLimiterWindowDuration,
			BlockDuration:   configs.RateLimiterBlockDuration,
		},
		logger,
	)

	rl := mdLimiter.NewRateLimiterMiddleware(limiter, logger)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/", rl.Handler(http.HandlerFunc(handlers.HomeHandler)))

	return &Server{Router: r}
}
