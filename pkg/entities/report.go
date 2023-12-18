package entities

import (
	"log/slog"
	"path/filepath"
	"time"
)

// PuppetReportSummary is the structure used to represent a series
// of puppet-runs against a particular node.
type PuppetReportSummary struct {
	// ID is the ID of the node.
	ID string `json:"id" bson:"id"`

	// Fqdn is the FQDN of the node.
	Fqdn string `json:"fqdn" bson:"fqdn"`

	// Env is the environment of the node.
	Env Environment `json:"env" bson:"env"`

	// State is the state of the node.
	State State `json:"state" bson:"state"`

	// ExecTime is the time the puppet-run was completed. This is self-reported by the node, and copied almost literally.
	ExecTime Datetime `json:"exec_time" bson:"exec_time"`

	// TimeSince is the time since the puppet-run was completed.
	TimeSince Duration `json:"time_since" bson:"time_since"`

	// Runtime is the time the puppet-run took.
	Runtime Duration `json:"runtime" bson:"runtime"`

	// Failed is the number of resources which failed.
	Failed int `json:"failed" bson:"failed"`

	// Changed is the number of resources which changed.
	Changed int `json:"changed" bson:"changed"`

	// Skipped is the number of resources which were skipped.
	Skipped int `json:"skipped" bson:"skipped"`

	// Total is the total number of resources.
	Total int `json:"total" bson:"total"`

	// YamlFile is the file the report was read from.
	YamlFile string `json:"yamlFile" bson:"yamlFile"`
}

func (n *PuppetReportSummary) CalculateTimeSince() {
	if n.ExecTime.Time().IsZero() {
		slog.Warn("CalculateTimeSince called with zero ExecTime")
		return
	}
	n.TimeSince = Duration(time.Since(n.ExecTime.Time()))
}

func (n *PuppetReportSummary) ReportFilePath() string {
	path := filepath.Join("reports", n.Env.String(), n.Fqdn, n.ExecTime.Time().Format(time.RFC3339)+".yaml")
	n.YamlFile = path
	return path
}

// PuppetReport stores the details of a single run of puppet.
type PuppetReport struct {
	// ID is a hash of the report-body. This is used to create the file to store the report in on-disk,
	// and as a means of detecting duplication submissions.
	ID string `json:"id" bson:"id"`

	// Fqdn of the node.
	Fqdn string `json:"fqdn" bson:"fqdn"`

	// Env of the node.
	Env Environment `json:"env" bson:"env"`

	// State of the run. changed, unchanged, etc.
	State State `json:"state" bson:"state"`

	// ExecTime is the time the puppet-run was completed. This is self-reported by the node, and copied almost literally.
	ExecTime Datetime `json:"exec_time" bson:"exec_time"`

	// Runtime is the time the puppet-run took.
	Runtime Duration `json:"runtime" bson:"runtime"`

	// Failed is the number of resources which failed.
	Failed int64 `json:"failed" bson:"failed"`

	// Changed is the number of resources which changed.
	Changed int64 `json:"changed" bson:"changed"`

	// Skipped is the number of resources which were skipped.
	Skipped int64 `json:"skipped" bson:"skipped"`

	// Total is the total number of resources.
	Total int64 `json:"total" bson:"total"`

	// LogMessages are the messages logged by puppet.
	LogMessages []string `json:"log_messages" bson:"log_messages"`

	// ResourcesFailed are the resources which failed.
	ResourcesFailed []*PuppetResource `json:"resources_failed" bson:"resources_failed"`

	// ResourcesChanged are the resources which changed.
	ResourcesChanged []*PuppetResource `json:"resources_changed" bson:"resources_changed"`

	// ResourcesSkipped are the resources which were skipped.
	ResourcesSkipped []*PuppetResource `json:"resources_skipped" bson:"resources_skipped"`

	// ResourcesOK are the resources which were OK.
	ResourcesOK []*PuppetResource `json:"resources_ok" bson:"resources_ok"`

	// YamlFile is the file the report was read from.
	YamlFile string `json:"yamlFile" bson:"yamlFile"`
}

func (n *PuppetReport) ReportFilePath() string {
	path := filepath.Join("reports", n.Env.String(), n.Fqdn, n.ExecTime.Time().Format(time.RFC3339)+".yaml")
	n.YamlFile = path
	return path
}
