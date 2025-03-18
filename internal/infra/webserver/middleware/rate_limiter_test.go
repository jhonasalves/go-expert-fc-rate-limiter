package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/mocks"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/pkg/ratelimiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StorageMock struct {
	mock.Mock
}

func (m *StorageMock) IsBlocked(ctx context.Context, key string) (bool, time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Get(1).(time.Duration), args.Error(2)
}

func (m *StorageMock) IncrRequest(ctx context.Context, key string, window time.Duration) (int, time.Duration, error) {
	args := m.Called(ctx, key, window)
	return args.Int(0), args.Get(1).(time.Duration), args.Error(2)
}

func (m *StorageMock) BlockRequest(ctx context.Context, key string, duration time.Duration) error {
	args := m.Called(ctx, key, duration)
	return args.Error(0)
}

func (m *StorageMock) clearMocks() {
	m.ExpectedCalls = nil
}

func TestRateLimiterMiddleware_Handler(t *testing.T) {
	mockStorage := new(StorageMock)
	logger := &mocks.LoggerMock{}
	logger.On("NewLogger").Return(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	opts := ratelimiter.Options{
		MaxRequestIP:    5,
		MaxRequestToken: 10,
		WindowDuration:  time.Minute,
		BlockDuration:   time.Minute * 5,
	}
	rateLimiter := ratelimiter.NewRateLimiter(mockStorage, opts, logger.NewLogger())
	middleware := NewRateLimiterMiddleware(rateLimiter, logger.NewLogger())

	t.Run("should return OK for allowed request", func(t *testing.T) {
		defer mockStorage.clearMocks()

		ctx := context.Background()
		rk := ratelimiter.RateLimitKey{Key: "test-key", KeyType: ratelimiter.Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(1, time.Minute, nil)

		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderAPIKey, rk.Key)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, w.Result().Body.Close())
		mockStorage.AssertExpectations(t)
	})

	t.Run("should return rate limit exceeded when too many requests", func(t *testing.T) {
		defer mockStorage.clearMocks()

		ctx := context.Background()
		rk := ratelimiter.RateLimitKey{Key: "test-key", KeyType: ratelimiter.Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(11, time.Minute, nil)
		mockStorage.On("BlockRequest", ctx, rk.Key, opts.BlockDuration).Return(nil)

		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderAPIKey, rk.Key)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("should validate HTTP headers for rate limiting", func(t *testing.T) {
		defer mockStorage.clearMocks()

		ctx := context.Background()
		rk := ratelimiter.RateLimitKey{Key: "test-key", KeyType: ratelimiter.Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(3, time.Minute, nil)

		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderAPIKey, rk.Key)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "7", w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
		assert.NoError(t, w.Result().Body.Close())
		mockStorage.AssertExpectations(t)
	})

	t.Run("should validate HTTP headers when rate limit is exceeded", func(t *testing.T) {
		defer mockStorage.clearMocks()

		ctx := context.Background()
		rk := ratelimiter.RateLimitKey{Key: "test-key", KeyType: ratelimiter.Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(11, time.Minute, nil)
		mockStorage.On("BlockRequest", ctx, rk.Key, opts.BlockDuration).Return(nil)

		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderAPIKey, rk.Key)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
	})
}
