package handlers

import (
	"encoding/json"
	"net/http"

	"awesomeProject/internal/service"

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

	// Отдаем как есть (массив зашифрованных в БД карт)
	respondWithJSON(w, http.StatusOK, cards)
}
