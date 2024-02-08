package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/messages"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

// uploadHandler takes the uploaded file and stores it in the database.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if uploadToken != "" {
		// Check the token.
		if r.Header.Get("Authorization") != "Bearer "+uploadToken {
			slog.Warn("Invalid upload token", slog.String("token", r.Header.Get("Authorization")))
			request.UnauthorizedHandler().ServeHTTP(w, r)
			return
		}
	}

	if r.Body == http.NoBody {
		slog.Warn("Request body is empty")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrBadRequest)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Get the request body as a byte slice.
	bdy, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("Error reading request body", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
	rep, err := parsePuppetReport(bdy)
	if err != nil {
		slog.Warn("Error parsing puppet report", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrBadRequest)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
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
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Save the run to the database.
	err = dataaccess.DB.SaveRun(r.Context(), rep)
	if errors.Is(err, dataaccess.ErrDuplicate) {
		slog.Warn("Duplicate run", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrDuplicate)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	} else if err != nil {
		slog.Error("Error saving run to database", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}

	// Return the report in the response.
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		slog.Warn("Error encoding response", slog.String(logging.KeyError, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
		return
	}
}
