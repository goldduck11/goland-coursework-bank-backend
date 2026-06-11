package handlers

import (
	"encoding/json"
	"net/http"

	"awesomeProject/internal/service"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type CreditHandler struct {
	service *service.CreditService
	logger  *logrus.Logger
}

func NewCreditHandler(s *service.CreditService, l *logrus.Logger) *CreditHandler {
	return &CreditHandler{service: s, logger: l}
}

func (h *CreditHandler) ApplyForCredit(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	var req struct {
		AccountID  string  `json:"account_id"`
		Principal  float64 `json:"principal"`
		TermMonths int     `json:"term_months"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	creditID, err := h.service.ApplyForCredit(r.Context(), userID, req.AccountID, req.Principal, req.TermMonths)
	if err != nil {
		h.logger.WithError(err).Error("Credit application failed")
		respondWithError(w, http.StatusInternalServerError, "Failed to apply for credit")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"credit_id": creditID, "status": "approved"})
}

func (h *CreditHandler) GetPaymentSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID := vars["creditId"]

	schedule, err := h.service.GetPaymentSchedule(r.Context(), creditID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch payment schedule")
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch schedule")
		return
	}

	respondWithJSON(w, http.StatusOK, schedule)
}
