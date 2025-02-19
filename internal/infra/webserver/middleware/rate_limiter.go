package middleware

import (
	"net"
	"net/http"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/pkg/ratelimiter"
)

type RateLimiterMiddleware struct {
	limiter *ratelimiter.RateLimiter
}

func NewRateLimiterMiddleware(l *ratelimiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{limiter: l}
}

func (rl *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getIP(r)

		if !rl.limiter.Allow(ip) {
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
