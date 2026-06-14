package router

import (
	"banking-system/internal/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterCreditRoutes(api *mux.Router, h *handlers.CreditHandler) {
	api.HandleFunc("/credits", h.ApplyForCredit).Methods(http.MethodPost)
	api.HandleFunc("/credits/{creditId}/schedule", h.GetPaymentSchedule).Methods(http.MethodGet)
}
