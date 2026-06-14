package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"banking-system/internal/repository/postgres"
	"banking-system/internal/service"

	"github.com/sirupsen/logrus"
)

type TransferHandler struct {
	service *service.AccountService
	logger  *logrus.Logger
}

func NewTransferHandler(s *service.AccountService, l *logrus.Logger) *TransferHandler {
	return &TransferHandler{service: s, logger: l}
}

func (h *TransferHandler) ExecuteTransfer(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	var req struct {
		SenderAccountID   string  `json:"sender_account_id"`
		ReceiverAccountID string  `json:"receiver_account_id"`
		Amount            float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than zero")
		return
	}

	err := h.service.Transfer(r.Context(), req.SenderAccountID, req.ReceiverAccountID, req.Amount, userID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrInsufficientFunds):
			respondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, postgres.ErrAccountNotFound):
			respondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, postgres.ErrAccessDenied):
			respondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.WithError(err).Error("Transfer operation failed")
			respondWithError(w, http.StatusInternalServerError, "Transaction failed")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Funds transferred successfully"})
}
