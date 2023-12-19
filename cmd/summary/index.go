package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/gorilla/mux"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// See if the environment has been provided in the URL.
	envStr, ok := mux.Vars(r)["env"]
	var env entities.Environment
	if !ok {
		env = entities.EnvAll
	} else {
		envStr = strings.ToUpper(envStr)
		env = entities.Environment(envStr)
		if !env.Valid() {
			// Respond with 400 bad request.
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(request.NewMessage("Invalid environment provided")); err != nil {
				slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
			}
			return
		}
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
			// Now check if the node is already in the map.
			if _, ok := filteredNodesMap[node.Fqdn]; !ok {
				filteredNodesMap[node.Fqdn] = node
			} else {
				// The node is already in the map, so we need to check if the node has a newer timestamp.
				if node.ExecTime.Time().After(filteredNodesMap[node.Fqdn].ExecTime.Time()) {
					filteredNodesMap[node.Fqdn] = node
				}
			}
		}
	}

	filteredNodes := make([]*entities.PuppetRun, 0, len(nodes))
	for _, node := range filteredNodesMap {
		filteredNodes = append(filteredNodes, node)
	}

	history, err := dataaccess.DB.GetHistory(r.Context(), env)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			// Respond with 404 not found.
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(request.NewMessage("No history found")); err != nil {
				slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
			}
			return
		}

		slog.Error("Error getting history", slog.String(logging.KeyError, err.Error()))
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting history")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
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
	str := d.String()

	// Add a space between each unit.
	str = strings.ReplaceAll(str, "d", "d ")
	str = strings.ReplaceAll(str, "h", "h ")
	str = strings.ReplaceAll(str, "m", "m ")
	str = strings.ReplaceAll(str, "s", "s ")

	// Remove the last space.
	str = strings.TrimSuffix(str, " ")

	return str
}
