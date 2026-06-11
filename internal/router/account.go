package router

import (
	"awesomeProject/internal/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterAccountRoutes(
	api *mux.Router,
	account *handlers.AccountHandler,
	card *handlers.CardHandler,
	transfer *handlers.TransferHandler,
	analytics *handlers.AnalyticsHandler,
) {
	// Счета
	api.HandleFunc("/accounts", account.CreateAccount).Methods(http.MethodPost)
	api.HandleFunc("/accounts/{accountId}/predict", analytics.PredictBalance).Methods(http.MethodGet)

	// Карты
	api.HandleFunc("/cards", card.IssueCard).Methods(http.MethodPost)
	api.HandleFunc("/cards", card.GetCards).Methods(http.MethodGet)

	// Переводы
	api.HandleFunc("/transfer", transfer.ExecuteTransfer).Methods(http.MethodPost)

	// Общая статистика
	api.HandleFunc("/analytics", analytics.GetMonthlyStats).Methods(http.MethodGet)
}
