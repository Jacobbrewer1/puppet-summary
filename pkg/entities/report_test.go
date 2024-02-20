package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ReportSummarySuite struct {
	suite.Suite

	// reportSummary is the instance of the report summary to be used in the tests.
	reportSummary *PuppetReportSummary
}

func TestReportSummarySuite(t *testing.T) {
	suite.Run(t, new(ReportSummarySuite))
}

func (s *ReportSummarySuite) SetupTest() {
	s.reportSummary = &PuppetReportSummary{
		ID:       "test-id",
		Fqdn:     "test-fqdn",
		Env:      "test-env",
		State:    "test-state",
		ExecTime: Datetime(time.Now()),
		Runtime:  Duration(10 * time.Second),
		Failed:   0,
		Changed:  0,
		Skipped:  0,
		Total:    0,
		YamlFile: "test-yaml-file",
	}
}

func (s *ReportSummarySuite) TestPuppetReportSummary_CalculateTimeSince() {
	s.reportSummary.CalculateTimeSince()

	s.Require().NotZero(s.reportSummary.TimeSince)
	s.Require().NotZero(s.reportSummary.ExecTime)
}

func (s *ReportSummarySuite) TestPuppetReportSummary_ReportFilePath() {
	filePath := s.reportSummary.ReportFilePath()

	s.Require().NotEmpty(filePath)

	s.Require().Equal("reports/test-env/test-fqdn/"+s.reportSummary.ExecTime.Time().Format(time.RFC3339)+".yaml", filePath)
}
