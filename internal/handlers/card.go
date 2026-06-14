package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"banking-system/internal/repository/postgres"
	"banking-system/internal/service"

	"github.com/sirupsen/logrus"
)

type CardHandler struct {
	service *service.CardService
	logger  *logrus.Logger
}

func NewCardHandler(s *service.CardService, l *logrus.Logger) *CardHandler {
	return &CardHandler{service: s, logger: l}
}

func (h *CardHandler) IssueCard(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	var req struct {
		AccountID string `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.IssueCard(r.Context(), userID, req.AccountID)
	if err != nil {
		if errors.Is(err, postgres.ErrAccessDenied) || errors.Is(err, postgres.ErrAccountNotFound) {
			respondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		h.logger.WithError(err).Error("Card issuance failed")
		respondWithError(w, http.StatusInternalServerError, "Failed to issue card")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success", "message": "Card issued successfully"})
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	cards, err := h.service.GetCards(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch cards")
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch cards")
		return
	}

	respondWithJSON(w, http.StatusOK, cards)
}

func (h *CardHandler) PayWithCard(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	var req struct {
		CardID string  `json:"card_id"`
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

	err := h.service.PayWithCard(r.Context(), userID, req.CardID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrInsufficientFunds):
			respondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, postgres.ErrAccessDenied):
			respondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.logger.WithError(err).Error("Card payment failed")
			respondWithError(w, http.StatusInternalServerError, "Payment failed")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Payment completed"})
}
