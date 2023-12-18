package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error parsing form")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Get the query term from the form.
	query := r.Form.Get("term")

	// Validate the query.
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Invalid query")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Get the nodes from the database.
	nodes, err := dataaccess.DB.GetRuns(r.Context())
	if err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("Error getting nodes from database", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting nodes from database")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Filter the nodes by the query.
	filteredNodes := make([]*entities.PuppetRun, 0, len(nodes))
	for _, node := range nodes {
		if strings.Contains(node.Fqdn, query) {
			filteredNodes = append(filteredNodes, node)
		}
	}

	type PageData struct {
		Nodes     []*entities.PuppetRun
		Term      string
		URLPrefix string
	}

	// Create the page data.
	pageData := &PageData{
		Nodes:     filteredNodes,
		Term:      query,
		URLPrefix: "/",
	}

	// Parse the template.
	tmpl := template.Must(template.ParseFiles("assets/results.gohtml"))

	// Execute the template.
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, pageData); err != nil {
		slog.Warn("Error executing template", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error executing template")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
}
