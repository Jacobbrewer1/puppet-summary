package dataaccess

import (
	"context"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/stretchr/testify/mock"
)

type DbMock struct {
	mock.Mock
}

func (d *DbMock) Ping(ctx context.Context) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *DbMock) Close(ctx context.Context) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *DbMock) SaveRun(ctx context.Context, run *entities.PuppetReport) error {
	args := d.Called(ctx, run)
	return args.Error(0)
}

func (d *DbMock) GetRuns(ctx context.Context) ([]*entities.PuppetRun, error) {
	args := d.Called(ctx)
	return args.Get(0).([]*entities.PuppetRun), args.Error(1)
}

func (d *DbMock) GetRunsByState(ctx context.Context, state entities.State) ([]*entities.PuppetRun, error) {
	args := d.Called(ctx, state)
	return args.Get(0).([]*entities.PuppetRun), args.Error(1)
}

func (d *DbMock) GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error) {
	args := d.Called(ctx, fqdn)
	return args.Get(0).([]*entities.PuppetReportSummary), args.Error(1)
}

func (d *DbMock) GetReport(ctx context.Context, id string) (*entities.PuppetReport, error) {
	args := d.Called(ctx, id)
	return args.Get(0).(*entities.PuppetReport), args.Error(1)
}

func (d *DbMock) GetHistory(ctx context.Context, environment entities.Environment) ([]*entities.PuppetHistory, error) {
	args := d.Called(ctx, environment)
	return args.Get(0).([]*entities.PuppetHistory), args.Error(1)
}

func (d *DbMock) GetEnvironments(ctx context.Context) ([]entities.Environment, error) {
	args := d.Called(ctx)
	return args.Get(0).([]entities.Environment), args.Error(1)
}

func (d *DbMock) Purge(ctx context.Context, from time.Time) (int, error) {
	args := d.Called(ctx, from)
	return args.Int(0), args.Error(1)
}
