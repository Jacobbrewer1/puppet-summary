package api

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GetAllNodesSuite struct {
	suite.Suite

	// db is the database used for testing.
	db *dataaccess.MockDb

	svc *service
}

func TestIndexSuite(t *testing.T) {
	suite.Run(t, new(GetAllNodesSuite))
}

func (s *GetAllNodesSuite) SetupTest() {
	s.db = new(dataaccess.MockDb)
	s.svc = &service{
		r: s.db,
	}
}

func (s *GetAllNodesSuite) TearDownTest() {
	s.db = nil
}

func (s *GetAllNodesSuite) TestGetAllNodes() {
	m := s.db

	now := time.Now().UTC()

	runs := []*entities.PuppetRun{
		{
			Fqdn:     "test1",
			Env:      summary.Environment_PRODUCTION,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
		{
			Fqdn:     "test2",
			Env:      summary.Environment_STAGING,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
		{
			Fqdn:     "test3",
			Env:      summary.Environment_PRODUCTION,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
	}

	// Test the index handler with the API.
	m.On("GetRuns", mock.AnythingOfType("context.backgroundCtx")).Return(runs, nil).Once()
	m.On("GetHistory", mock.AnythingOfType("context.backgroundCtx"), []summary.Environment(nil)).Return([]*entities.PuppetHistory{}, nil).Once()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/nodes", nil)

	s.svc.GetAllNodes(w, r)

	s.Equal(200, w.Code)
	s.Equal("application/json", w.Header().Get("Content-Type"))

	// Compare the response.
	expected := "[{\"fqdn\":\"test1\",\"env\":\"PRODUCTION\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"},{\"fqdn\":\"test2\",\"env\":\"STAGING\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"},{\"fqdn\":\"test3\",\"env\":\"PRODUCTION\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"}]\n"
	s.Require().Equal(expected, w.Body.String())

	s.db.AssertExpectations(s.T())
}
