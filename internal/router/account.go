package router

import (
	"net/http"

	"banking-system/internal/handlers"

	"github.com/gorilla/mux"
)

func RegisterAccountRoutes(
	api *mux.Router,
	account *handlers.AccountHandler,
	card *handlers.CardHandler,
	transfer *handlers.TransferHandler,
	analytics *handlers.AnalyticsHandler,
) {
	api.HandleFunc("/accounts", account.CreateAccount).Methods(http.MethodPost)
	api.HandleFunc("/accounts", account.GetAccounts).Methods(http.MethodGet)
	api.HandleFunc("/accounts/{accountId}/deposit", account.Deposit).Methods(http.MethodPost)
	api.HandleFunc("/accounts/{accountId}/withdraw", account.Withdraw).Methods(http.MethodPost)
	api.HandleFunc("/accounts/{accountId}/predict", analytics.PredictBalance).Methods(http.MethodGet)

	api.HandleFunc("/cards", card.IssueCard).Methods(http.MethodPost)
	api.HandleFunc("/cards", card.GetCards).Methods(http.MethodGet)
	api.HandleFunc("/cards/pay", card.PayWithCard).Methods(http.MethodPost)

	api.HandleFunc("/transfer", transfer.ExecuteTransfer).Methods(http.MethodPost)
	api.HandleFunc("/analytics", analytics.GetMonthlyStats).Methods(http.MethodGet)
}
