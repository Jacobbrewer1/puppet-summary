package entities

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
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
	s.history.AddCount(StateChanged, 7)
	s.Equal(7, s.history.Changed)
	s.Equal(0, s.history.Unchanged)
	s.Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddUnchanged() {
	s.history.AddCount(StateUnchanged, 7)
	s.Equal(7, s.history.Unchanged)
	s.Equal(0, s.history.Changed)
	s.Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddFailed() {
	s.history.AddCount(StateFailed, 7)
	s.Equal(7, s.history.Failed)
	s.Equal(0, s.history.Changed)
	s.Equal(0, s.history.Unchanged)
}

func (s *historySuite) TestAddCount() {
	s.history.AddCount(StateChanged, 7)
	s.history.AddCount(StateUnchanged, 5)
	s.history.AddCount(StateFailed, 3)
	s.Equal(7, s.history.Changed)
	s.Equal(5, s.history.Unchanged)
	s.Equal(3, s.history.Failed)
}

func (s *historySuite) TestAddCountInvalidState() {
	s.history.AddCount(State("invalid"), 7)
	s.Equal(0, s.history.Changed)
	s.Equal(0, s.history.Unchanged)
	s.Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddCountNegativeCount() {
	s.history.AddCount(StateChanged, -7)
	s.Equal(-7, s.history.Changed)
	s.Equal(0, s.history.Unchanged)
	s.Equal(0, s.history.Failed)
}

func (s *historySuite) TestAddCountChanged() {
	s.history.AddCount(StateChanged, 7)
	s.Equal(7, s.history.Changed)
}
