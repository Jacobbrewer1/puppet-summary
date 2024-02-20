package main

import (
	"github.com/smallfish/simpleyaml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ParsePuppetReportSuite struct {
	suite.Suite

	report *entities.PuppetReport

	yamlContent []byte
	sy          *simpleyaml.Yaml
}

func TestParsePuppetReportSuite(t *testing.T) {
	suite.Run(t, new(ParsePuppetReportSuite))
}

func (s *ParsePuppetReportSuite) SetupTest() {
	s.report = &entities.PuppetReport{}

	// Read the valid YAML content from the file in the data directory
	pwd, err := os.Getwd()
	s.Require().NoError(err, "Unexpected error getting working directory")
	s.Require().NotEmpty(pwd, "Expected a non-empty working directory")

	if !strings.Contains(pwd, "cmd/summary") {
		pwd += filepath.Join(pwd, "cmd/summary")
		s.Require().DirExists(pwd, "Expected a data directory")
	}

	validYAMLFile := filepath.Join(pwd, "data/example.yaml")
	s.Require().FileExists(validYAMLFile, "Expected a valid YAML file")

	s.yamlContent, err = os.ReadFile(validYAMLFile)
	s.Require().NoError(err, "Unexpected error reading valid YAML file")
	s.Require().NotEmpty(s.yamlContent, "Expected non-empty YAML content")

	// Create a new simpleyaml object
	s.sy, err = simpleyaml.NewYaml(s.yamlContent)
	s.Require().NoError(err, "Unexpected error creating simpleyaml object")
	s.Require().NotNil(s.sy, "Expected a non-nil simpleyaml object")
}

func (s *ParsePuppetReportSuite) TearDownTest() {
	s.yamlContent = nil
	s.report = nil
}

func (s *ParsePuppetReportSuite) TestParsePuppetReport_InvalidYAML() {
	// Define your invalid YAML content for testing
	invalidYAML := []byte("invalid YAML content")

	// Call the parsePuppetReport function
	report, err := parsePuppetReport(invalidYAML)

	// Assertions
	s.Error(err, "Expected an error for invalid YAML")
	s.Nil(report, "Expected a nil report for invalid YAML")
}

func (s *ParsePuppetReportSuite) TestParseHost() {
	err := parseHost(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing host")
	s.Equal("example-host", s.report.Fqdn)
}

func (s *ParsePuppetReportSuite) TestParsePuppetVersion() {
	err := parsePuppetVersion(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing puppet version")
	s.Equal(8.4, s.report.PuppetVersion)
}

func (s *ParsePuppetReportSuite) TestParseEnvironment() {
	err := parseEnvironment(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing environment")
	s.Equal(entities.EnvProduction, s.report.Env)
}

func (s *ParsePuppetReportSuite) TestParseTime() {
	err := parseTime(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing time")
	exeTime, err := time.Parse(time.RFC3339, "2024-02-17T02:00:09+00:00")
	s.Require().NoError(err, "Unexpected error parsing time")
	s.Equal(exeTime.UTC(), s.report.ExecTime.Time().UTC())
}

func (s *ParsePuppetReportSuite) TestParseStatus() {
	err := parseStatus(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing status")
	s.Equal(entities.StateChanged, s.report.State)
}

func (s *ParsePuppetReportSuite) TestParseRuntime() {
	err := parseRuntime(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing runtime")

	runtime, err := time.ParseDuration("26.67511224s")
	s.Require().NoError(err, "Unexpected error parsing runtime")
	s.Equal(entities.Duration(runtime), s.report.Runtime)
}

func (s *ParsePuppetReportSuite) TestParseResources() {
	err := parseResources(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing resources")
	s.Equal(int64(0), s.report.Failed)
	s.Equal(int64(6), s.report.Changed)
	s.Equal(int64(0), s.report.Skipped)
	s.Equal(int64(67), s.report.Total)
}

func (s *ParsePuppetReportSuite) TestParseLogs() {
	err := parseLogs(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing logs")

	expectedLogs := []string{
		"/Stage[main]/Default_config/Exec[example-command1]/returns : Testing if example-command1 is already installed",
		"/Stage[main]/Default_config/Exec[example-command2]/returns : executed successfully",
		"Puppet : Applied catalog in 26.67 seconds",
	}
	s.Equal(expectedLogs, s.report.LogMessages)
}

func (s *ParsePuppetReportSuite) TestParseResults() {
	err := parseResults(s.sy, s.report)
	s.NoError(err, "Unexpected error parsing results")
	s.Equal(1, len(s.report.ResourcesChanged))
	s.Equal(0, len(s.report.ResourcesSkipped))
	s.Equal(0, len(s.report.ResourcesOK))
}
