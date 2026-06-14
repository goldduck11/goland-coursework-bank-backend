package handlers

import (
	"encoding/json"
	"net/http"

	"banking-system/internal/middleware"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func getUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
		return userID
	}
	return ""
}
