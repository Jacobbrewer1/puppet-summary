package api

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
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
			State:    summary.State_SKIPPED,
		},
		{
			Fqdn:     "test2",
			Env:      summary.Environment_STAGING,
			ExecTime: entities.Datetime(now.Add(10 * time.Second)),
			Runtime:  entities.Duration(10 * time.Second),
			State:    summary.State_UNCHANGED,
		},
		{
			Fqdn:     "test3",
			Env:      summary.Environment_DEVELOPMENT,
			ExecTime: entities.Datetime(now.Add(20 * time.Second)),
			Runtime:  entities.Duration(10 * time.Second),
			State:    summary.State_CHANGED,
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/nodes", nil)

	// Test the index handler with the API.
	m.On("GetRuns", r.Context()).Return(runs, nil).Once()

	s.svc.GetAllNodes(w, r)

	s.Equal(200, w.Code)
	s.Equal("application/json", w.Header().Get("Content-Type"))

	// Compare the response.
	expected := "[{\"env\":\"PRODUCTION\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"fqdn\":\"test1\",\"runtime\":\"10s\",\"state\":\"SKIPPED\"},{\"env\":\"STAGING\",\"exec_time\":\"" + now.Add(10*time.Second).Format(time.RFC3339) + "\",\"fqdn\":\"test2\",\"runtime\":\"10s\",\"state\":\"UNCHANGED\"},{\"env\":\"DEVELOPMENT\",\"exec_time\":\"" + now.Add(20*time.Second).Format(time.RFC3339) + "\",\"fqdn\":\"test3\",\"runtime\":\"10s\",\"state\":\"CHANGED\"}]\n"
	s.Require().Equal(expected, w.Body.String())

	s.db.AssertExpectations(s.T())
}
