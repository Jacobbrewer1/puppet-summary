package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

const ctxEnvKey = "env"

func (svc webService) GetAllNodesByEnvironment(w http.ResponseWriter, r *http.Request, env summary.Environment) {
	ctx := r.Context()

	entEnv := entities.Environment(env)
	if !entEnv.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Invalid environment")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	ctx = context.WithValue(ctx, ctxEnvKey, entEnv)

	svc.GetAllNodes(w, r.WithContext(ctx))
}

func (svc webService) GetAllNodes(w http.ResponseWriter, r *http.Request) {
	env := entities.EnvAll
	if r.Context().Value(ctxEnvKey) != nil {
		env = r.Context().Value(ctxEnvKey).(entities.Environment)
	}

	nodes, err := dataaccess.DB.GetRuns(r.Context())
	if err != nil && !errors.Is(err, dataaccess.ErrNotFound) {
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting nodes")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Filter the nodes by environment.
	filteredNodesMap := make(map[string]*entities.PuppetRun)
	for _, node := range nodes {
		// Check that the node is in the environment.
		if env == entities.EnvAll || node.Env.IsIn(env) {
			node.CalculateTimeSince()

			// Create a key for the node. This should be the FQDN and the environment.
			key := fmt.Sprintf("%s-%s", node.Fqdn, node.Env)

			// Now check if the node is already in the map.
			if _, ok := filteredNodesMap[key]; !ok {
				filteredNodesMap[key] = node
			} else {
				// The node is already in the map, so we need to check if the node has a newer timestamp.
				if node.ExecTime.Time().After(filteredNodesMap[key].ExecTime.Time()) {
					filteredNodesMap[key] = node
				}
			}
		}
	}

	filteredNodes := make([]*entities.PuppetRun, 0, len(nodes))
	for _, node := range filteredNodesMap {
		filteredNodes = append(filteredNodes, node)
	}

	// Sort the nodes by the time since the last puppet-run. This will put the nodes with the newest puppet-runs at
	// the top.
	sort.Slice(filteredNodes, func(i, j int) bool {
		return filteredNodes[i].TimeSince.Time() < filteredNodes[j].TimeSince.Time()
	})

	history, err := dataaccess.DB.GetHistory(r.Context(), env)
	if err != nil && !errors.Is(err, dataaccess.ErrNotFound) {
		slog.Error("Error getting history", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting history")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// If the history is not empty, then sort them by timestamp.
	if len(history) > 0 {
		sort.Slice(history, func(i, j int) bool {
			// Parse the timestamps.
			iTime, err := time.Parse(time.DateOnly, history[i].Date)
			if err != nil {
				slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
				return false
			}

			jTime, err := time.Parse(time.DateOnly, history[j].Date)
			if err != nil {
				slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
				return false
			}

			// Compare the timestamps.
			return iTime.Before(jTime)
		})
	}

	envs, err := dataaccess.DB.GetEnvironments(r.Context())
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			// Respond with 404 not found.
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(request.NewMessage("No environments found")); err != nil {
				slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
			}
			return
		}

		slog.Error("Error getting environments", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting environments")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Remove ID's from the nodes.
	for i := range filteredNodes {
		filteredNodes[i].ID = ""
	}

	// If the path prefix is "api", then we want to just encode the nodes as JSON and return them.
	if strings.HasPrefix(r.URL.Path, "/api") {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(filteredNodes); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	type PageData struct {
		Graph        []*entities.PuppetHistory
		Nodes        []*entities.PuppetRun
		Environment  entities.Environment
		Environments []entities.Environment
		URLPrefix    string
	}

	pd := &PageData{
		Graph:        history,
		Nodes:        filteredNodes,
		Environment:  env,
		Environments: envs,
		URLPrefix:    "",
	}

	// Read the page template from the file.
	page, err := os.OpenFile("assets/index.gohtml", os.O_RDONLY, os.ModePerm)
	if err != nil {
		slog.Error("Error opening page file", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error reading page template")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	defer func() {
		if err := page.Close(); err != nil {
			slog.Error("Error closing page file", slog.String(logging.KeyError, err.Error()))
		}
	}()

	pt, err := io.ReadAll(page)
	if err != nil {
		slog.Error("Error reading page file", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error reading page template")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Parse the template.
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"prettyDuration": prettyDuration,
	}).Parse(string(pt)))

	// Execute the template.
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, pd); err != nil {
		slog.Warn("Error executing template", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error executing template")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
}

func prettyDuration(d *entities.Duration) string {
	if d == nil {
		return ""
	}

	// Get the duration as a string.
	str := d.PrettyString()

	// Add a space between each unit.
	str = strings.ReplaceAll(str, "d", "d ")
	str = strings.ReplaceAll(str, "h", "h ")
	str = strings.ReplaceAll(str, "m", "m ")
	str = strings.ReplaceAll(str, "s", "s ")

	// Remove the last space.
	str = strings.TrimSuffix(str, " ")

	return str
}
