package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

func (s service) GetAllNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.r.GetRuns(r.Context())
	if err != nil && !errors.Is(err, dataaccess.ErrNotFound) {
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting nodes")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	nodesMap := make(map[string]*entities.PuppetRun)
	for _, node := range nodes {
		node.CalculateTimeSince()

		// Create a key for the node. This should be the FQDN and the environment.
		key := fmt.Sprintf("%s-%s", node.Fqdn, node.Env)

		// Now check if the node is already in the map.
		if _, ok := nodesMap[key]; !ok {
			nodesMap[key] = node
		} else {
			// The node is already in the map, so we need to check if the node has a newer timestamp.
			if node.ExecTime.Time().After(nodesMap[key].ExecTime.Time()) {
				nodesMap[key] = node
			}
		}
	}

	// Create the response.
	mappedNodes := make([]summary.Node, 0, len(nodesMap))
	for _, node := range nodesMap {
		mappedNodes = append(mappedNodes, summary.Node{
			Env:      &node.Env,
			ExecTime: summary.Point(node.ExecTime.String()),
			Fqdn:     &node.Fqdn,
			Runtime:  summary.Point(node.Runtime.String()),
			State:    &node.State,
		})
	}

	nodesResponse := new(summary.NodesResponse)
	nodesResponse.Nodes = &mappedNodes

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(mappedNodes); err != nil {
		slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
	}
}

func (s service) GetAllNodesByEnvironment(w http.ResponseWriter, r *http.Request, env summary.Environment) {
	nodes, err := s.r.GetRuns(r.Context())
	if err != nil && !errors.Is(err, dataaccess.ErrNotFound) {
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting nodes")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	nodesMap := make(map[string]*entities.PuppetRun)
	for _, node := range nodes {
		if node.Env != env {
			continue
		}

		node.CalculateTimeSince()

		// Create a key for the node. This should be the FQDN and the environment.
		key := fmt.Sprintf("%s-%s", node.Fqdn, node.Env)

		// Now check if the node is already in the map.
		if _, ok := nodesMap[key]; !ok {
			nodesMap[key] = node
		} else {
			// The node is already in the map, so we need to check if the node has a newer timestamp.
			if node.ExecTime.Time().After(nodesMap[key].ExecTime.Time()) {
				nodesMap[key] = node
			}
		}
	}

	// Create the response.
	mappedNodes := make([]summary.Node, 0, len(nodesMap))
	for _, node := range nodesMap {
		mappedNodes = append(mappedNodes, summary.Node{
			Env:      &node.Env,
			ExecTime: summary.Point(node.ExecTime.String()),
			Fqdn:     &node.Fqdn,
			Runtime:  summary.Point(node.Runtime.String()),
			State:    &node.State,
		})
	}

	nodesResponse := new(summary.NodesResponse)
	nodesResponse.Nodes = &mappedNodes

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(mappedNodes); err != nil {
		slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
	}
}

func (s service) GetNodeByFqdn(w http.ResponseWriter, r *http.Request, fqdn string) {
	reps, err := s.r.GetReports(r.Context(), fqdn)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting reports")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	if len(reps) == 0 {
		// Respond with 404 not found.
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(request.NewMessage("No reports found for node %s", fqdn)); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Sort the reports by the time they were received.
	sort.Slice(reps, func(i, j int) bool {
		return reps[i].ExecTime.Time().Before(reps[j].ExecTime.Time())
	})

	resp := make([]*summary.PuppetReportSummary, 0, len(reps))
	for _, rep := range reps {
		resp = append(resp, &summary.PuppetReportSummary{
			Changed:  &rep.Changed,
			Env:      &rep.Env,
			ExecTime: summary.Point(rep.ExecTime.Time()),
			Failed:   &rep.Failed,
			Fqdn:     &rep.Fqdn,
			Id:       &rep.ID,
			Runtime:  summary.Point(rep.Runtime.String()),
			Skipped:  &rep.Skipped,
			State:    &rep.State,
			Total:    &rep.Total,
		})
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
	}
}
