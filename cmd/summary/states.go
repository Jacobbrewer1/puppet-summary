package main

//func stateHandler(w http.ResponseWriter, r *http.Request) {
//	stateStr, ok := mux.Vars(r)["state_id"]
//	if !ok {
//		w.WriteHeader(http.StatusBadRequest)
//		if err := json.NewEncoder(w).Encode(request.NewMessage("%s is not a recognized state", stateStr)); err != nil {
//			slog.Warn("Error encoding response", slog.String("error", err.Error()))
//		}
//		return
//	}
//
//	state := entities.State(strings.ToUpper(stateStr))
//	if !state.Valid() {
//		w.WriteHeader(http.StatusBadRequest)
//		if err := json.NewEncoder(w).Encode(request.NewMessage("%s is not a recognized state", state)); err != nil {
//			slog.Warn("Error encoding response", slog.String("error", err.Error()))
//		}
//		return
//	}
//
//	// Get the state from the database.
//	runs, err := dataaccess.DB.GetRunsByState(r.Context(), state)
//	if err != nil {
//		if !errors.Is(err, context.Canceled) {
//			slog.Error("Error getting runs from database", slog.String("error", err.Error()))
//		}
//		w.WriteHeader(http.StatusInternalServerError)
//		if err := json.NewEncoder(w).Encode(request.NewMessage("Error getting runs from database")); err != nil {
//			slog.Warn("Error encoding response", slog.String("error", err.Error()))
//		}
//		return
//	}
//
//	if err := json.NewEncoder(w).Encode(runs); err != nil {
//		slog.Warn("Error encoding response", slog.String("error", err.Error()))
//		w.WriteHeader(http.StatusInternalServerError)
//		if err := json.NewEncoder(w).Encode(request.NewMessage("Error encoding response")); err != nil {
//			slog.Warn("Error encoding response", slog.String("error", err.Error()))
//		}
//		return
//	}
//}
