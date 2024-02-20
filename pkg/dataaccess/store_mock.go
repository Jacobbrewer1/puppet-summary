package dataaccess

import (
	"context"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveFile(ctx context.Context, filePath string, file []byte) error {
	args := m.Called(ctx, filePath, file)
	return args.Error(0)
}

func (m *MockStorage) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	args := m.Called(ctx, filePath)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStorage) DeleteFile(ctx context.Context, filePath string) error {
	args := m.Called(ctx, filePath)
	return args.Error(0)
}

func (m *MockStorage) Purge(ctx context.Context, from time.Time) (int, error) {
	args := m.Called(ctx, from)
	return args.Int(0), args.Error(1)
}
