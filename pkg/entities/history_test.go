package entities

import (
	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type historySuite struct {
	suite.Suite

	// history is the history to test.
	history *PuppetHistory
}

func TestHistorySuite(t *testing.T) {
	suite.Run(t, new(historySuite))
}

func (s *historySuite) SetupTest() {
	now := time.Now().Add(-time.Hour * 24)
	s.history = &PuppetHistory{
		Date:      now.Format(time.DateOnly),
		Changed:   0,
		Unchanged: 0,
		Failed:    0,
	}
}

func (s *historySuite) TearDownTest() {
	s.history = nil
}

func (s *historySuite) TestAddChanged() {
	s.history.AddCount(summary.State_CHANGED, 7)
	s.Require().Equal(7, s.history.Changed)
	s.Require().Equal(0, s.history.Unchanged)
	s.Require().Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddUnchanged() {
	s.history.AddCount(summary.State_UNCHANGED, 7)
	s.Require().Equal(7, s.history.Unchanged)
	s.Require().Equal(0, s.history.Changed)
	s.Require().Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddFailed() {
	s.history.AddCount(summary.State_FAILED, 7)
	s.Require().Equal(7, s.history.Failed)
	s.Require().Equal(0, s.history.Changed)
	s.Require().Equal(0, s.history.Unchanged)
}

func (s *historySuite) TestAddCount() {
	s.history.AddCount(summary.State_CHANGED, 7)
	s.history.AddCount(summary.State_UNCHANGED, 5)
	s.history.AddCount(summary.State_FAILED, 3)
	s.Require().Equal(7, s.history.Changed)
	s.Require().Equal(5, s.history.Unchanged)
	s.Require().Equal(3, s.history.Failed)
}

func (s *historySuite) TestAddCountInvalidState() {
	s.history.AddCount(summary.State("invalid"), 7)
	s.Require().Equal(0, s.history.Changed)
	s.Require().Equal(0, s.history.Unchanged)
	s.Require().Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddCountNegativeCount() {
	s.history.AddCount(summary.State_CHANGED, -7)
	s.Require().Equal(-7, s.history.Changed)
	s.Require().Equal(0, s.history.Unchanged)
	s.Require().Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddCountChanged() {
	s.history.AddCount(summary.State_CHANGED, 7)
	s.Require().Equal(7, s.history.Changed)
}
