package middleware

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/pkg/ratelimiter"
)

const HeaderAPIKey = "API_KEY"

type RateLimiterMiddleware struct {
	limiter *ratelimiter.RateLimiter
	logger  *slog.Logger
}

type RateLimitErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Limit      int    `json:"limit"`
	Remaining  int    `json:"remaining"`
	ResetAfter int    `json:"reset_after"`
}

func NewRateLimiterMiddleware(l *ratelimiter.RateLimiter, logger *slog.Logger) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: l,
		logger:  logger,
	}
}

func (rl *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ip := rl.getIP(r)
		token := rl.getToken(r)

		var key string
		var keyType ratelimiter.KeyType

		if token != "" {
			key = token
			keyType = ratelimiter.Token
		} else {
			key = ip
			keyType = ratelimiter.API
		}

		resp, err := rl.limiter.Allow(ctx, key, keyType)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !resp.Allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", resp.RetryAfter.Format(http.TimeFormat))
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(resp.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(0))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(resp.RetryAfter.Sub(resp.RetryAfter).Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)

			response := RateLimitErrorResponse{
				Error:      "rate_limit_exceeded",
				Message:    "you have reached the maximum number of requests or actions allowed within a certain time frame",
				Limit:      resp.Limit,
				Remaining:  0,
				ResetAfter: int(resp.RetryAfter.Sub(resp.RetryAfter).Seconds()),
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(resp.Limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(0))
		w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(resp.RetryAfter.Sub(resp.RetryAfter).Seconds())))

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiterMiddleware) getIP(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func (rl *RateLimiterMiddleware) getToken(r *http.Request) string {
	return r.Header.Get(HeaderAPIKey)
}
