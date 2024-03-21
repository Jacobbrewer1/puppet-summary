package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

func (s service) PurgePuppetReports(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		slog.Warn("missing request body")

		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("missing request body")); err != nil {
			slog.Warn("failed to encode response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Decode the request.
	req := new(summary.PurgePuppetReportsJSONBody)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		slog.Warn("failed to decode request", slog.String(logging.KeyError, err.Error()))

		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("failed to decode request")); err != nil {
			slog.Warn("failed to encode response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Ensure that the date is in the past.
	if req.Date.IsZero() || req.Date.After(time.Now().UTC()) {
		slog.Warn("invalid date", slog.String("date", req.Date.String()))

		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("invalid date")); err != nil {
			slog.Warn("failed to encode response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Calculate the number of days between the date and now.
	days := int(time.Since(req.Date.Time).Hours() / 24)

	// Purge the reports.
	if err := s.purger.PurgePuppetReports(days); err != nil {
		slog.Error("failed to purge reports", slog.String(logging.KeyError, err.Error()))

		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("failed to purge reports")); err != nil {
			slog.Warn("failed to encode response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(request.NewMessage("reports purged")); err != nil {
		slog.Warn("failed to encode response", slog.String(logging.KeyError, err.Error()))
	}
}
