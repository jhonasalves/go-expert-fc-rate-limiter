package middleware

import (
	"net"
	"net/http"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/pkg/ratelimiter"
)

const HeaderAPIKey = "API_KEY"

type RateLimiterMiddleware struct {
	limiter *ratelimiter.RateLimiter
}

func NewRateLimiterMiddleware(l *ratelimiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{limiter: l}
}

func (rl *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ip := rl.getIP(r)
		token := rl.getToken(r)

		var key string

		if token != "" {
			key = token
		} else {
			key = ip
		}

		if !rl.limiter.Allow(ctx, key) {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}

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
