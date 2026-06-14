package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"banking-system/internal/handlers"
	"banking-system/internal/middleware"
)

// Config объединяет все хендлеры и зависимости для роутера
type Config struct {
	Logger           *logrus.Logger
	JWTSecret        string
	AuthHandler      *handlers.AuthHandler
	AccountHandler   *handlers.AccountHandler
	CardHandler      *handlers.CardHandler
	TransferHandler  *handlers.TransferHandler
	CreditHandler    *handlers.CreditHandler
	AnalyticsHandler *handlers.AnalyticsHandler
}

// New создает и конфигурирует итоговый роутер приложения
func New(cfg Config) *mux.Router {
	r := mux.NewRouter()

	// Глобальное middleware для логирования
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			cfg.Logger.Infof("[%s] %s", req.Method, req.URL.Path)
			next.ServeHTTP(w, req)
		})
	})

	// 1. Подключаем публичные маршруты (не требуют JWT)
	RegisterAuthRoutes(r, cfg.AuthHandler)

	// 2. Создаем изолированную зону для защищенных маршрутов
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.JWTMiddleware(cfg.JWTSecret, cfg.Logger))

	// Подключаем защищенные маршруты из других файлов этого же пакета
	RegisterAccountRoutes(api, cfg.AccountHandler, cfg.CardHandler, cfg.TransferHandler, cfg.AnalyticsHandler)
	RegisterCreditRoutes(api, cfg.CreditHandler)

	return r
}
