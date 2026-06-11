package handlers

import (
	"encoding/json"
	"net/http"
)

// respondWithError отправляет стандартизированный JSON с ошибкой
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON сериализует данные в JSON и отправляет клиенту
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

// getUserIDFromContext безопасно достает ID пользователя, сохраненный JWT-middleware
func getUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value("userID").(string); ok {
		return userID
	}
	return ""
}
