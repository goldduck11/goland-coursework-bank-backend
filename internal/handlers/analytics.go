package handlers

import (
	"net/http"
	"strconv"

	"banking-system/internal/service"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
	logger  *logrus.Logger
}

func NewAnalyticsHandler(s *service.AnalyticsService, l *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: s, logger: l}
}

func (h *AnalyticsHandler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	predictedBalance, err := h.service.PredictBalance(r.Context(), accountID, days, userID)
	if err != nil {
		h.logger.WithError(err).Error("Balance prediction failed")
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"account_id":        accountID,
		"prediction_days":   days,
		"predicted_balance": predictedBalance,
	})
}

func (h *AnalyticsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	stats, err := h.service.GetMonthlyStats(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Analytics failed")
		respondWithError(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}
	respondWithJSON(w, http.StatusOK, stats)
}
