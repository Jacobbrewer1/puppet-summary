package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/Jacobbrewer1/puppet-summary/pkg/services/parser"
)

func (s service) GetReportById(w http.ResponseWriter, r *http.Request, id string) {
	// Get the report from the database.
	rep, err := s.r.GetReport(r.Context(), id)
	if errors.Is(err, dataaccess.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Report not found")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	} else if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Error("Error getting report", slog.String(logging.KeyError, err.Error()))
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting report")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Check if the report exists.
	if rep == nil {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Report not found")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Get the Yaml report from Files.
	file, err := dataaccess.Files.DownloadFile(r.Context(), rep.ReportFilePath())
	if err != nil {
		slog.Error("Error downloading yaml file", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error downloading yaml file")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Parse the parser file.
	report, err := parser.ParsePuppetReport(file)
	if err != nil {
		slog.Error("Error parsing yaml file", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error parsing yaml file")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Sort the report resources.
	report.SortResources()
	rep = report

	resp := summary.PuppetReport{
		Changed:          summary.Point(int(rep.Changed)),
		Env:              &rep.Env,
		ExecTime:         summary.Point(rep.ExecTime.Time()),
		Failed:           summary.Point(int(rep.Failed)),
		Fqdn:             &rep.Fqdn,
		Id:               &rep.ID,
		LogMessages:      &rep.LogMessages,
		PuppetVersion:    summary.Point(float32(rep.PuppetVersion)),
		ResourcesChanged: nil, // Map later.
		ResourcesFailed:  nil, // Map later.
		ResourcesOk:      nil, // Map later.
		ResourcesSkipped: nil, // Map later.
		Runtime:          summary.Point(rep.Runtime.String()),
		Skipped:          summary.Point(int(rep.Skipped)),
		State:            &rep.State,
		Total:            summary.Point(int(rep.Total)),
	}

	// Map the resources.
	changed := make([]summary.Resource, 0, len(rep.ResourcesChanged))
	for _, change := range rep.ResourcesChanged {
		changed = append(changed, summary.Resource{
			File: &change.File,
			Line: &change.Line,
			Name: &change.Name,
			Type: &change.Type,
		})
	}

	failed := make([]summary.Resource, 0, len(rep.ResourcesFailed))
	for _, fail := range rep.ResourcesFailed {
		failed = append(failed, summary.Resource{
			File: &fail.File,
			Line: &fail.Line,
			Name: &fail.Name,
			Type: &fail.Type,
		})
	}

	ok := make([]summary.Resource, 0, len(rep.ResourcesOK))
	for _, o := range rep.ResourcesOK {
		ok = append(ok, summary.Resource{
			File: &o.File,
			Line: &o.Line,
			Name: &o.Name,
			Type: &o.Type,
		})
	}

	skipped := make([]summary.Resource, 0, len(rep.ResourcesSkipped))
	for _, skip := range rep.ResourcesSkipped {
		skipped = append(skipped, summary.Resource{
			File: &skip.File,
			Line: &skip.Line,
			Name: &skip.Name,
			Type: &skip.Type,
		})
	}

	if len(changed) > 0 {
		resp.ResourcesChanged = &changed
	}
	if len(failed) > 0 {
		resp.ResourcesFailed = &failed
	}
	if len(ok) > 0 {
		resp.ResourcesOk = &ok
	}
	if len(skipped) > 0 {
		resp.ResourcesSkipped = &skipped
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
	}
}
