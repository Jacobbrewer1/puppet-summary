package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/messages"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/Jacobbrewer1/puppet-summary/pkg/services/parser"
)

func (s service) UploadPuppetReport(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		slog.Warn("Request body is empty")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrBadRequest)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Get the request body as a byte slice.
	bdy, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("Error reading request body", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
	rep, err := parser.ParsePuppetReport(bdy)
	if err != nil {
		slog.Warn("Error parsing puppet report", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrBadRequest)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Generate the file path.
	rep.ReportFilePath()

	// Save the file to Files.
	err = dataaccess.Files.SaveFile(r.Context(), rep.YamlFile, bdy)
	if err != nil {
		slog.Error("Error saving file to Files", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Save the run to the database.
	err = s.r.SaveRun(r.Context(), rep)
	if errors.Is(err, dataaccess.ErrDuplicate) {
		slog.Warn("Duplicate run", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrDuplicate)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	} else if err != nil {
		slog.Error("Error saving run to database", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

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

	// Return the report in the response.
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
}
