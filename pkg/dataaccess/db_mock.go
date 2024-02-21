package dataaccess

import (
	"context"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/stretchr/testify/mock"
)

type MockDb struct {
	mock.Mock
}

func (m *MockDb) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDb) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDb) SaveRun(ctx context.Context, run *entities.PuppetReport) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

func (m *MockDb) GetRuns(ctx context.Context) ([]*entities.PuppetRun, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.PuppetRun), args.Error(1)
}

func (m *MockDb) GetRunsByState(ctx context.Context, state entities.State) ([]*entities.PuppetRun, error) {
	args := m.Called(ctx, state)
	return args.Get(0).([]*entities.PuppetRun), args.Error(1)
}

func (m *MockDb) GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error) {
	args := m.Called(ctx, fqdn)
	return args.Get(0).([]*entities.PuppetReportSummary), args.Error(1)
}

func (m *MockDb) GetReport(ctx context.Context, id string) (*entities.PuppetReport, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.PuppetReport), args.Error(1)
}

func (m *MockDb) GetHistory(ctx context.Context, environment entities.Environment) ([]*entities.PuppetHistory, error) {
	args := m.Called(ctx, environment)
	return args.Get(0).([]*entities.PuppetHistory), args.Error(1)
}

func (m *MockDb) GetEnvironments(ctx context.Context) ([]entities.Environment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.Environment), args.Error(1)
}

func (m *MockDb) Purge(ctx context.Context, from time.Time) (int, error) {
	args := m.Called(ctx, from)
	return args.Int(0), args.Error(1)
}
