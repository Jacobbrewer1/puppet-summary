package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type indexSuite struct {
	suite.Suite

	// db is the database used for testing.
	db *dataaccess.MockDb
}

func TestIndexSuite(t *testing.T) {
	suite.Run(t, new(indexSuite))
}

func (s *indexSuite) SetupTest() {
	s.db = new(dataaccess.MockDb)
	dataaccess.DB = s.db
}

func (s *indexSuite) TearDownTest() {
	s.db.AssertExpectations(s.T())
	s.db = nil
}

func (s *indexSuite) TestIndexHandlerAPI() {
	m := s.db

	now := time.Now().UTC()

	runs := []*entities.PuppetRun{
		{
			Fqdn:     "test1",
			Env:      entities.EnvProduction,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
		{
			Fqdn:     "test2",
			Env:      entities.EnvStaging,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
		{
			Fqdn:     "test3",
			Env:      entities.EnvProduction,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
	}

	// Test the index handler with the API.
	m.On("GetRuns", mock.AnythingOfType("context.backgroundCtx")).Return(runs, nil).Once()
	m.On("GetHistory", mock.AnythingOfType("context.backgroundCtx"), entities.EnvAll).Return([]*entities.PuppetHistory{}, nil).Once()
	m.On("GetEnvironments", mock.AnythingOfType("context.backgroundCtx")).Return([]entities.Environment{entities.EnvProduction, entities.EnvStaging}, nil).Once()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/nodes", nil)

	indexHandler(w, r)

	s.Equal(200, w.Code)
	s.Equal("application/json", w.Header().Get("Content-Type"))

	// Compare the response.
	expected := "[{\"id\":\"\",\"fqdn\":\"test1\",\"env\":\"PRODUCTION\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"},{\"id\":\"\",\"fqdn\":\"test2\",\"env\":\"STAGING\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"},{\"id\":\"\",\"fqdn\":\"test3\",\"env\":\"PRODUCTION\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"}]\n"
	s.Require().Equal(expected, w.Body.String())
}
