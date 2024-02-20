package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type durationSuite struct {
	suite.Suite
}

func TestDurationSuite(t *testing.T) {
	suite.Run(t, new(durationSuite))
}

func (s *durationSuite) TestString() {
	d := Duration(0)
	s.Require().Equal("", d.String())

	d = Duration(1 * time.Second)
	s.Require().Equal("1s", d.String())

	d = Duration((12 * time.Second) + (3 * time.Minute) + (4 * time.Hour) + (5 * 24 * time.Hour))
	s.Require().Equal("124h3m12s", d.String())
}

func (s *durationSuite) TestScan() {
	d := Duration(0)
	err := d.Scan("1s")
	s.Require().NoError(err)
	s.Require().Equal(Duration(1*time.Second), d)

	err = d.Scan("1m")
	s.Require().NoError(err)
	s.Require().Equal(Duration(1*time.Minute), d)

	err = d.Scan("1h")
	s.Require().NoError(err)
	s.Require().Equal(Duration(1*time.Hour), d)

	err = d.Scan("7h30m")
	s.Require().NoError(err)
	s.Require().Equal(Duration((7*time.Hour)+(30*time.Minute)), d)
}
