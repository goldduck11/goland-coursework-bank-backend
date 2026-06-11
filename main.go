package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"awesomeProject/internal/handlers"
	"awesomeProject/internal/repository/postgres"
	"awesomeProject/internal/router"
	"awesomeProject/internal/service"
	"awesomeProject/internal/service/external"
)

var initSQL string

func main() {
	// 1. Настройка логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Info("Starting Banking Service...")

	// 2. Строгое чтение переменных окружения (БЕЗ ХАРДКОДА)
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		logger.Fatal("CRITICAL: DATABASE_URL environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatal("CRITICAL: JWT_SECRET environment variable is not set")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		logger.Fatal("CRITICAL: SERVER_PORT environment variable is not set")
	}

	// 3. Создаем контекст для отлова системных сигналов
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Подключение к БД
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		logger.WithError(err).Fatal("Failed to open database")
	}
	defer db.Close()

	// Автоматический запуск скрипта инициализации таблиц
	logger.Info("Running database initialization script...")

	if _, err := db.Exec(initSQL); err != nil {
		logger.WithError(err).Fatal("Failed to initialize database schema")
	}
	logger.Infof("Database schema is up to date! (Executed %d bytes of SQL)", len(initSQL))

	db.SetMaxOpenConns(25)
	if err := db.Ping(); err != nil {
		logger.WithError(err).Fatal("Database is unreachable")
	}

	// 4. Инициализация слоев (Dependency Injection)
	cbrClient := external.NewCBRClient()

	userRepo := postgres.NewUserRepository(db)
	accountRepo := postgres.NewAccountRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	creditRepo := postgres.NewCreditRepository(db)

	authService := service.NewAuthService(userRepo, jwtSecret)
	accountService := service.NewAccountService(accountRepo)
	cardService := service.NewCardService(cardRepo)
	creditService := service.NewCreditService(creditRepo, accountRepo, cbrClient)
	analyticsService := service.NewAnalyticsService(db)

	// 5. Запуск фонового шедулера (каждые 12 часов)
	scheduler := service.NewCreditScheduler(db, logger)
	scheduler.Start(ctx)

	// 6. Сборка маршрутизатора из нашего нового пакета router
	mainRouter := router.New(router.Config{
		Logger:           logger,
		JWTSecret:        jwtSecret,
		AuthHandler:      handlers.NewAuthHandler(authService, logger),
		AccountHandler:   handlers.NewAccountHandler(accountService, logger),
		CardHandler:      handlers.NewCardHandler(cardService, logger),
		TransferHandler:  handlers.NewTransferHandler(accountService, logger),
		CreditHandler:    handlers.NewCreditHandler(creditService, logger),
		AnalyticsHandler: handlers.NewAnalyticsHandler(analyticsService, logger),
	})

	// 7. Конфигурация и запуск HTTP-сервера
	srv := &http.Server{
		Addr:         serverPort,
		Handler:      mainRouter,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Запускаем сервер в фоне, чтобы он не блокировал ожидание системных сигналов
	serverErrors := make(chan error, 1)
	go func() {
		logger.Infof("Server is listening on port %s", serverPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 8. Ожидание: либо сервер упал с ошибкой, либо юзер нажал Ctrl+C
	select {
	case err := <-serverErrors:
		logger.WithError(err).Fatal("HTTP server critical error")
	case <-ctx.Done():
		logger.Info("Shutdown signal received, stopping gracefully...")

		// Даем серверу 5 секунд на завершение текущих транзакций/запросов
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.WithError(err).Error("Server forced to shutdown")
		}
	}

	logger.Info("Banking Service stopped safely.")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
