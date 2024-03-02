package web

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
	"strconv"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/Jacobbrewer1/puppet-summary/pkg/services/parser"
	"github.com/gorilla/mux"
)

func (s service) reportIDHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["report_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("No report ID provided")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	} else if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage("Invalid report ID provided")); err != nil {
			slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Get the report from the database.
	rep, err := s.db.GetReport(r.Context(), id)
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

	// Parse the yaml file.
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

	type PageData struct {
		Report    *entities.PuppetReport
		URLPrefix string
	}

	pd := &PageData{
		Report:    rep,
		URLPrefix: "",
	}

	// Read the page template from the file.
	page, err := os.OpenFile("assets/report.gohtml", os.O_RDONLY, os.ModePerm)
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
		"truncate": func(s string) string {
			f, _ := strconv.ParseFloat(s, 64)
			s = fmt.Sprintf("%.2f", f)
			return s
		},
		"prettyTime": prettyTime,
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

func prettyTime(t entities.Datetime) string {
	return t.Time().Format(time.DateTime)
}
