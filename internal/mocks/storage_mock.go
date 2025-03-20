package mocks

import (
	"context"
	"time"

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

func (m *StorageMock) ClearMocks() {
	m.ExpectedCalls = nil
}
