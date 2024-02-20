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

func (s *ReportSummarySuite) TearDownTest() {
	s.reportSummary = nil
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

type ReportSuite struct {
	suite.Suite

	// report is the instance of the report to be used in the tests.
	report *PuppetReport
}

func TestReportSuite(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}

func (s *ReportSuite) SetupTest() {
	s.report = &PuppetReport{
		ID:              "test-id",
		Fqdn:            "test-fqdn",
		PuppetVersion:   8,
		Env:             "test-env",
		State:           "test-state",
		ExecTime:        Datetime(time.Now()),
		Runtime:         Duration(10 * time.Second),
		Failed:          0,
		Changed:         1,
		Skipped:         3,
		Total:           4,
		LogMessages:     nil,
		ResourcesFailed: nil,
		ResourcesChanged: []*PuppetResource{
			{
				Name: "TestResource1",
				Type: "TestType1",
				File: "TestFile1",
				Line: "TestLine1",
			},
		},
		ResourcesSkipped: []*PuppetResource{
			{
				Name: "TestResource1",
				Type: "TestType1",
				File: "TestFile1",
				Line: "TestLine1",
			},
			{
				Name: "TestResource2",
				Type: "TestType2",
				File: "TestFile2",
				Line: "TestLine2",
			},
			{
				Name: "TestResource3",
				Type: "TestType3",
				File: "TestFile3",
				Line: "TestLine3",
			},
		},
		ResourcesOK: nil,
		YamlFile:    "",
	}
}

func (s *ReportSuite) TearDownTest() {
	s.report = nil
}

func (s *ReportSuite) TestPuppetReport_ReportFilePath() {
	filePath := s.report.ReportFilePath()
	s.Require().NotEmpty(filePath)
	s.Require().Equal("reports/test-env/test-fqdn/"+s.report.ExecTime.Time().Format(time.RFC3339)+".yaml", filePath)
}

func (s *ReportSuite) TestPuppetReport_SortResources() {
	s.report.SortResources()
	s.Require().Nil(s.report.ResourcesFailed)
	s.Require().NotNil(s.report.ResourcesChanged)
	s.Require().NotNil(s.report.ResourcesSkipped)
	s.Require().Nil(s.report.ResourcesOK)
}

func (s *ReportSuite) TestPuppetReport_SortResource() {
	resources := []*PuppetResource{
		{
			Name: "TestResource2",
			Type: "TestType2",
			File: "TestFile2",
			Line: "TestLine2",
		},
		{
			Name: "TestResource1",
			Type: "TestType1",
			File: "TestFile1",
			Line: "TestLine1",
		},
		{
			Name: "TestResource3",
			Type: "TestType3",
			File: "TestFile3",
			Line: "TestLine3",
		},
	}
	s.report.sortResource(resources)
	s.Require().Equal("TestResource1", resources[0].Name)
	s.Require().Equal("TestResource2", resources[1].Name)
	s.Require().Equal("TestResource3", resources[2].Name)
}
