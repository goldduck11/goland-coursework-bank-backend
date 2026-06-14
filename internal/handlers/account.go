package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"banking-system/internal/repository/postgres"
	"banking-system/internal/service"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AccountHandler struct {
	service *service.AccountService
	logger  *logrus.Logger
}

func NewAccountHandler(s *service.AccountService, l *logrus.Logger) *AccountHandler {
	return &AccountHandler{service: s, logger: l}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	accountID, err := h.service.CreateAccount(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create account")
		respondWithError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"account_id": accountID, "status": "created"})
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	accounts, err := h.service.GetAccounts(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list accounts")
		respondWithError(w, http.StatusInternalServerError, "Failed to list accounts")
		return
	}

	respondWithJSON(w, http.StatusOK, accounts)
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	accountID := mux.Vars(r)["accountId"]

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than zero")
		return
	}

	if err := h.service.Deposit(r.Context(), accountID, req.Amount, userID); err != nil {
		h.handleAccountError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Account topped up"})
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	accountID := mux.Vars(r)["accountId"]

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than zero")
		return
	}

	if err := h.service.Withdraw(r.Context(), accountID, req.Amount, userID); err != nil {
		h.handleAccountError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Funds withdrawn"})
}

func (h *AccountHandler) handleAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, postgres.ErrInsufficientFunds):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, postgres.ErrAccountNotFound):
		respondWithError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, postgres.ErrAccessDenied):
		respondWithError(w, http.StatusForbidden, err.Error())
	default:
		h.logger.WithError(err).Error("Account operation failed")
		respondWithError(w, http.StatusInternalServerError, "Operation failed")
	}
}
