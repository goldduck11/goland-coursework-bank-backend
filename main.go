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

	"banking-system/internal/config"
	"banking-system/internal/core/logger"
	"banking-system/internal/handlers"
	"banking-system/internal/repository/postgres"
	"banking-system/internal/router"
	"banking-system/internal/service"
	"banking-system/internal/service/external"
	"banking-system/internal/utils"
)

//go:embed scripts/init.sql
var initSQL string

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	log := logger.New(cfg.LogLevel)
	log.Info("Starting Banking Service...")

	if err := utils.InitPGPKeys(cfg.PGPPublicKey, cfg.PGPPrivateKey, cfg.PGPPassphrase); err != nil {
		log.WithError(err).Fatal("Failed to initialize PGP keys")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to open database")
	}
	defer db.Close()

	log.Info("Running database initialization script...")
	if _, err := db.Exec(initSQL); err != nil {
		log.WithError(err).Fatal("Failed to initialize database schema")
	}
	log.Infof("Database schema is up to date! (Executed %d bytes of SQL)", len(initSQL))

	db.SetMaxOpenConns(25)
	if err := db.Ping(); err != nil {
		log.WithError(err).Fatal("Database is unreachable")
	}

	cbrClient := external.NewCBRClient()
	emailService := external.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom, log)

	userRepo := postgres.NewUserRepository(db)
	accountRepo := postgres.NewAccountRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	creditRepo := postgres.NewCreditRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, emailService)
	accountService := service.NewAccountService(accountRepo)
	cardService := service.NewCardService(cardRepo, accountRepo, cfg.HMACSecret)
	creditService := service.NewCreditService(creditRepo, accountRepo, cbrClient)
	analyticsService := service.NewAnalyticsService(db, transactionRepo)

	scheduler := service.NewCreditScheduler(db, userRepo, emailService, log)
	scheduler.Start(ctx)

	mainRouter := router.New(router.Config{
		Logger:           log,
		JWTSecret:        cfg.JWTSecret,
		AuthHandler:      handlers.NewAuthHandler(authService, log),
		AccountHandler:   handlers.NewAccountHandler(accountService, log),
		CardHandler:      handlers.NewCardHandler(cardService, log),
		TransferHandler:  handlers.NewTransferHandler(accountService, log),
		CreditHandler:    handlers.NewCreditHandler(creditService, log),
		AnalyticsHandler: handlers.NewAnalyticsHandler(analyticsService, log),
	})

	srv := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      mainRouter,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Infof("Server is listening on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		log.WithError(err).Fatal("HTTP server critical error")
	case <-ctx.Done():
		log.Info("Shutdown signal received, stopping gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.WithError(err).Error("Server forced to shutdown")
		}
	}

	log.Info("Banking Service stopped safely.")
}
