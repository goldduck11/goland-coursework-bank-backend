package handlers

import (
	"net/http"

	"awesomeProject/internal/service"

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
