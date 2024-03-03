package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/messages"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/gorilla/mux"
)

func middlewareHttp(handler http.Handler, authOption summary.AuthOption) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		cw := request.NewClientWriter(w)

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

		var path string
		route := mux.CurrentRoute(r)
		if route != nil { // The route may be nil if the request is not routed.
			var err error
			path, err = route.GetPathTemplate()
			if err != nil {
				// An error here is only returned if the route does not define a path.
				slog.Error("Error getting path template", slog.String(logging.KeyError, err.Error()))
				path = r.URL.Path // If the route does not define a path, use the URL path.
			}
		} else {
			path = r.URL.Path // If the route is nil, use the URL path.
		}

		reqSize := r.ContentLength

		switch authOption {
		case summary.AuthOptionNone:
		// Do nothing.
		case summary.AuthOptionRequired:
			if authToken != "" {
				// Check if the request has the correct token.
				token := r.Context().Value(summary.BearerAuthScopes)
				if token == nil || len(token.(string)) == 0 || token.(string) != authToken {
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

		httpTotalRequests.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Inc()
		httpRequestDuration.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Observe(time.Since(now).Seconds())
		httpRequestSize.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Observe(float64(reqSize))
	}
}
