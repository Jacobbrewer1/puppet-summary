package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/messages"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/gorilla/mux"
)

type Controller func(w http.ResponseWriter, r *http.Request)

type AuthOption int

const (
	// AuthOptionNone is the option for no authentication.
	AuthOptionNone AuthOption = iota

	// AuthOptionInternal is the option to only allow internal traffic.
	AuthOptionInternal

	// AuthOptionRequired is the option for required authentication.
	AuthOptionRequired
)

func middlewareHttp(handler Controller, authOption AuthOption) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// We want to check if the auth type is internal, and if so, check if the request is internal. This is because
		// the 404 will mess up the metrics.
		if authOption == AuthOptionInternal && !request.IsInternal(r) {
			// If the request is not internal, return a 404.
			slog.Debug("Request is not internal", slog.String("remote_addr", r.RemoteAddr),
				slog.String("headers", fmt.Sprintf("%+v", r.Header)))
			request.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		now := time.Now().UTC()
		cw := request.NewClientWriter(w)

		// Recover from any panics that occur in the handler.
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("Panic in handler",
					slog.String(logging.KeyError, rec.(error).Error()),
					slog.String("stack", string(debug.Stack())),
				)
				cw.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(cw).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
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

		defer func() {
			// Run the deferred function after the request has been handled, as the status code will not be available until then.
			httpTotalRequests.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Inc()
			httpRequestDuration.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Observe(time.Since(now).Seconds())
			httpRequestSize.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Observe(float64(r.ContentLength))
		}()

		switch authOption {
		case AuthOptionNone, AuthOptionInternal, AuthOptionRequired:
			// Do nothing.
		default:
			slog.Error("Invalid auth option", slog.Int("auth_option", int(authOption)))
			cw.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(cw).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
				slog.Warn("Failed to write response", slog.String(logging.KeyError, err.Error()))
			}
		}

		handler(cw, r)
	}
}
