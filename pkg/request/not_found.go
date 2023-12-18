package request

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
)

// NotFoundHandler returns a handler that returns a 404 response.
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := NewMessage("Not found")
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
	}
}

// MethodNotAllowedHandler returns a handler that returns a 405 response.
func MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := NewMessage("Method not allowed")
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		}
	}
}
