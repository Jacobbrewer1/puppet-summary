package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

func radiatorHandler(w http.ResponseWriter, r *http.Request) {
	// Get the states for all environments.
	states, err := getStates(r.Context(), entities.EnvProduction)
	if err != nil {
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting states")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	total := 0
	for _, state := range states {
		total += state.Count
	}

	// Add in the total count of nodes.
	states = append(states, &entities.PuppetState{
		State:   "Total",
		Count:   total,
		Percent: 100,
	})

	// PageData is the data for the page.
	type PageData struct {
		States    []*entities.PuppetState
		URLPrefix string
	}

	// Create the page data.
	pageData := &PageData{
		States:    states,
		URLPrefix: "/",
	}

	// Parse the template.
	tmpl := template.Must(template.ParseFiles("assets/radiator.gohtml"))

	// Execute the template.
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, pageData); err != nil {
		slog.Warn("Error executing template", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error executing template")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
}

func getStates(ctx context.Context, env entities.Environment) ([]*entities.PuppetState, error) {
	if !env.Valid() {
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	// Get the nodes from the database.
	nodes, err := dataaccess.DB.GetRuns(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting nodes from database: %w", err)
	}

	// Get the states from the nodes.
	states := make(map[entities.State]int)
	total := 0
	for _, node := range nodes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			total++
			switch env {
			case entities.EnvProduction:
				if node.Env.IsIn(entities.EnvProduction) {
					states[node.State]++
				}
			case entities.EnvDevelopment:
				if node.Env.IsIn(entities.EnvDevelopment) {
					states[node.State]++
				}
			}
		}
	}

	// Get the distinct states.
	distinctStates := make([]string, 0)
	for state := range states {
		distinctStates = append(distinctStates, string(state))
	}

	// Sort the states.
	sort.Strings(distinctStates)

	// Create the PuppetStates.
	puppetStates := make([]*entities.PuppetState, 0)
	for _, state := range distinctStates {
		ps := &entities.PuppetState{
			State: state,
			Count: states[entities.State(state)],
		}
		ps.Percent = float64(ps.Count) / float64(total) * 100

		// If the total is 0, then the percent will be NaN. Set it to 0.
		if total == 0 {
			ps.Percent = 0
		}

		puppetStates = append(puppetStates, ps)
	}

	return puppetStates, nil
}
