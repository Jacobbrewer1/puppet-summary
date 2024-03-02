package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

func (s service) GetAllNodesByState(w http.ResponseWriter, r *http.Request, state summary.State) {
	// Get the state from the database.
	runs, err := s.r.GetRunsByState(r.Context(), state)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Error("Error getting runs from database", slog.String("error", err.Error()))
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting runs from database")); err != nil {
			slog.Warn("Error encoding response", slog.String("error", err.Error()))
		}
		return
	}

	nodes := make([]*summary.Node, 0, len(runs))
	for _, run := range runs {
		nodes = append(nodes, &summary.Node{
			Env:      &run.Env,
			ExecTime: summary.Point(run.ExecTime.String()),
			Fqdn:     &run.Fqdn,
			Runtime:  summary.Point(run.Runtime.String()),
			State:    &run.State,
		})
	}

	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		slog.Warn("Error encoding response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error encoding response")); err != nil {
			slog.Warn("Error encoding response", slog.String("error", err.Error()))
		}
		return
	}
}
