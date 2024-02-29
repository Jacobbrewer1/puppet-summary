package main

import (
	"net/http"
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
		//	now := time.Now().UTC()
		//	cw := request.NewClientWriter(w)
		//
		//	// Recover from any panics that occur in the handler.
		//	defer func() {
		//		if rec := recover(); rec != nil {
		//			slog.Error("Panic in handler",
		//				slog.String(logging.KeyError, rec.(error).Error()),
		//				slog.String("stack", string(debug.Stack())),
		//			)
		//			cw.WriteHeader(http.StatusInternalServerError)
		//			if err := json.NewEncoder(cw).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
		//				slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		//			}
		//		}
		//	}()
		//
		//	var path string
		//	route := mux.CurrentRoute(r)
		//	if route != nil { // The route may be nil if the request is not routed.
		//		var err error
		//		path, err = route.GetPathTemplate()
		//		if err != nil {
		//			// An error here is only returned if the route does not define a path.
		//			slog.Error("Error getting path template", slog.String(logging.KeyError, err.Error()))
		//			path = r.URL.Path // If the route does not define a path, use the URL path.
		//		}
		//	} else {
		//		path = r.URL.Path // If the route is nil, use the URL path.
		//	}
		//
		//	switch authOption {
		//	case AuthOptionNone:
		//	// Do nothing.
		//	case AuthOptionRequired:
		//		if authToken != "" {
		//			// Check the token.
		//			if r.Header.Get("Authorization") != "Bearer "+authToken {
		//				slog.Warn("Invalid upload token", slog.String("token", r.Header.Get("Authorization")))
		//				request.UnauthorizedHandler().ServeHTTP(w, r)
		//				return
		//			}
		//		}
		//	case AuthOptionInternal:
		//		// Check if the request is internal.
		//		if !request.IsInternal(r) {
		//			cw.WriteHeader(http.StatusUnauthorized)
		//			if err := json.NewEncoder(cw).Encode(request.NewMessage(messages.ErrUnauthorized)); err != nil {
		//				slog.Error("Error encoding response", slog.String(logging.KeyError, err.Error()))
		//			}
		//			return
		//		}
		//	default:
		//		slog.Error("Invalid auth option", slog.Int("auth_option", int(authOption)))
		//		cw.WriteHeader(http.StatusInternalServerError)
		//		if err := json.NewEncoder(cw).Encode(request.NewMessage(messages.ErrInternalServer)); err != nil {
		//			slog.Warn("Failed to write response", slog.String(logging.KeyError, err.Error()))
		//		}
		//	}
		//
		//	handler(cw, r)
	}
}
