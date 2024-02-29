package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/messages"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
)

func middlewareHttp(handler http.Handler, authOption summary.AuthOption) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Recover from any panics that occur in the handler.
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("Panic in handler",
					slog.String(logging.KeyError, rec.(error).Error()),
					slog.String("stack", string(debug.Stack())),
				)
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
					slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
				}
			}
		}()

		switch authOption {
		case summary.AuthOptionNone:
		// Do nothing.
		case summary.AuthOptionRequired:
			if authToken != "" {
				// Check if the request has the correct token.
				token := r.Context().Value(summary.BearerAuthToken)
				if token == nil || token.(string) != authToken {
					w.WriteHeader(http.StatusUnauthorized)
					if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrUnauthorized)); err != nil {
						slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
					}
					return
				}
			}
		case summary.AuthOptionInternal:
			// Check if the request is internal.
			if !request.IsInternal(r) {
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrUnauthorized)); err != nil {
					slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
				}
				return
			}
		default:
			slog.Error("Invalid auth option", slog.Int("auth_option", int(authOption)))
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
				slog.Warn("Failed to write response", slog.String(logging.KeyError, err.Error()))
			}
		}

		handler.ServeHTTP(w, r)
	}
}
