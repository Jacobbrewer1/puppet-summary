package web

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
)

func (s service) nodeFqdnHandler(w http.ResponseWriter, r *http.Request) {
	nodeFqdn, ok := mux.Vars(r)["node_fqdn"]
	if !ok {
		// Respond with 400 bad request.
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("No node fqdn provided")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	} else if nodeFqdn == "" {
		// Respond with 400 bad request.
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("No node fqdn provided")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	reps, err := s.db.GetReports(r.Context(), nodeFqdn)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Respond with 500 internal server error.
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting reports")); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	if len(reps) == 0 {
		// Respond with 404 not found.
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(request.NewMessage("No reports found for node %s", nodeFqdn)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Sort the reports by the time they were received.
	sort.Slice(reps, func(i, j int) bool {
		return reps[i].ExecTime.Time().Before(reps[j].ExecTime.Time())
	})

	type PageData struct {
		Fqdn      string
		Nodes     []*entities.PuppetReportSummary
		URLPrefix string
	}

	pd := &PageData{
		Fqdn:      nodeFqdn,
		Nodes:     reps,
		URLPrefix: "",
	}

	// Read the page template from the file.
	page, err := os.OpenFile("assets/node.gohtml", os.O_RDONLY, os.ModePerm)
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
	tmpl := template.Must(template.New("node").Funcs(template.FuncMap{
		"inc": func(i int) string {
			return strconv.Itoa(i + 1)
		},
		"graphConvert":   graphRuntime,
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

// graphRuntime takes in the duration of time to graph, and returns the duration as seconds as an int.
func graphRuntime(d entities.Duration) int {
	return int(d.Time().Seconds())
}
