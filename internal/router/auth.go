package router

import (
	"banking-system/internal/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterAuthRoutes(r *mux.Router, h *handlers.AuthHandler) {
	r.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", h.Login).Methods(http.MethodPost)
}
