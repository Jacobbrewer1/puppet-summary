package entities

import (
	"log/slog"
	"path/filepath"
	"sort"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
)

// PuppetReportSummary is the structure used to represent a series
// of puppet-runs against a particular node.
type PuppetReportSummary struct {
	// ID is the ID of the node.
	ID string `json:"id" bson:"id"`

	// Fqdn is the FQDN of the node.
	Fqdn string `json:"fqdn" bson:"fqdn"`

	// Env is the environment of the node.
	Env summary.Environment `json:"env" bson:"env"`

	// State is the state of the node.
	State summary.State `json:"state" bson:"state"`

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
	YamlFile string `json:"-" bson:"yamlFile"`
}

func (n *PuppetReportSummary) CalculateTimeSince() {
	if n.ExecTime.Time().IsZero() {
		slog.Warn("CalculateTimeSince called with zero ExecTime")
		return
	}
	n.TimeSince = Duration(time.Since(n.ExecTime.Time()))
}

func (n *PuppetReportSummary) ReportFilePath() string {
	path := filepath.Join("reports", string(n.Env), n.Fqdn, n.ExecTime.Time().Format(time.RFC3339)+".yaml")
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

	// PuppetVersion is the version of puppet used to generate the report.
	PuppetVersion float64 `json:"puppet_version" bson:"puppet_version"`

	// Env of the node.
	Env summary.Environment `json:"env" bson:"env"`

	// State of the run. changed, unchanged, etc.
	State summary.State `json:"state" bson:"state"`

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
	YamlFile string `json:"-" bson:"yamlFile"`
}

func (n *PuppetReport) ReportFilePath() string {
	path := filepath.Join("reports", string(n.Env), n.Fqdn, n.ExecTime.Time().Format(time.RFC3339)+".yaml")
	n.YamlFile = path
	return path
}

func (n *PuppetReport) SortResources() {
	// Sort the resources.
	n.sortResource(n.ResourcesFailed)
	n.sortResource(n.ResourcesChanged)
	n.sortResource(n.ResourcesSkipped)
	n.sortResource(n.ResourcesOK)
}

func (n *PuppetReport) sortResource(resources []*PuppetResource) {
	sort.Slice(resources, func(i, j int) bool {
		// Sort by resource file.
		if resources[i].File != resources[j].File {
			return resources[i].File < resources[j].File
		}

		// Sort by resource line.
		if resources[i].Line != resources[j].Line {
			return resources[i].Line < resources[j].Line
		}

		// Sort by resource type.
		if resources[i].Type != resources[j].Type {
			return resources[i].Type < resources[j].Type
		}

		// Sort by resource name.
		if resources[i].Name != resources[j].Name {
			return resources[i].Name < resources[j].Name
		}

		// Resources are equal.
		return false
	})
}
