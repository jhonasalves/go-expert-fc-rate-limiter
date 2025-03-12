package mocks

import (
	"log/slog"

	"github.com/stretchr/testify/mock"
)

type LoggerMock struct {
	mock.Mock
}

func (m *LoggerMock) NewLogger() *slog.Logger {
	args := m.Called()
	return args.Get(0).(*slog.Logger)
}
