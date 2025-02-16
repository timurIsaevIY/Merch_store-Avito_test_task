package main

import (
	authHandler "Merch_store-Avito_test_task/internal/pkg/auth/delivery/http"
	authRepo "Merch_store-Avito_test_task/internal/pkg/auth/repository"
	authUsecase "Merch_store-Avito_test_task/internal/pkg/auth/usecase"
	"Merch_store-Avito_test_task/internal/pkg/config"
	httpresponse "Merch_store-Avito_test_task/internal/pkg/httpresponses"
	"Merch_store-Avito_test_task/internal/pkg/jwt"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	paymentsHandler "Merch_store-Avito_test_task/internal/pkg/payments/delivery/http"
	paymentsRepo "Merch_store-Avito_test_task/internal/pkg/payments/repository"
	paymentsUsecase "Merch_store-Avito_test_task/internal/pkg/payments/usecase"
	serviceHandler "Merch_store-Avito_test_task/internal/pkg/service/delivery/http"
	serviceRepo "Merch_store-Avito_test_task/internal/pkg/service/repository"
	serviceUsecase "Merch_store-Avito_test_task/internal/pkg/service/usecase"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtHandler := jwt.NewJTW(jwtSecret, logger)

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Database.DbHost, cfg.Database.DbPort, cfg.Database.DbUser, cfg.Database.DbPass, cfg.Database.DbName))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(100)                 // Maximum number of open connections
	db.SetMaxIdleConns(50)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(10 * time.Minute) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(10 * time.Minute)
	log.Printf("DB_HOST: %s, DB_PORT: %d, DB_USER: %s, DB_PASS: %s, DB_NAME: %s", cfg.Database.DbHost, cfg.Database.DbPort, cfg.Database.DbUser, cfg.Database.DbPass, cfg.Database.DbName)
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	authRepo := authRepo.NewAuthRepositoryImpl(db)
	authUsecase := authUsecase.NewAuthUsecase(authRepo)
	authHandler := authHandler.NewAuthHandler(authUsecase, logger, jwtHandler)

	paymentsRepo := paymentsRepo.NewAuthRepositoryImpl(db)
	paymentsUsecase := paymentsUsecase.NewPaymentsUsecase(paymentsRepo)
	paymentsHandler := paymentsHandler.NewPaymentsHandler(paymentsUsecase, logger)

	serviceRepo := serviceRepo.NewServiceRepo(db)
	serviceUsecase := serviceUsecase.NewServiceUsecase(serviceRepo)
	serviceHandler := serviceHandler.NewServiceHandler(serviceUsecase, logger)

	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := httpresponse.Response{
			Message: "Not found",
		}
		httpresponse.SendJSONResponse(r.Context(), w, response, http.StatusNotFound, logger)
	})
	r.HandleFunc("/healthcheck", healthcheckHandler).Methods(http.MethodGet)

	r.HandleFunc("/auth", authHandler.Login).Methods(http.MethodPost)
	r.Handle("/sendCoin", middleware.AuthMiddleware(jwtHandler, http.HandlerFunc(paymentsHandler.SendCoins), logger)).Methods(http.MethodPost)
	r.Handle("/buy/{item}", middleware.AuthMiddleware(jwtHandler, http.HandlerFunc(paymentsHandler.BuyItem), logger)).Methods(http.MethodGet)
	r.Handle("/info", middleware.AuthMiddleware(jwtHandler, http.HandlerFunc(serviceHandler.GetUserInfo), logger)).Methods(http.MethodGet)

	httpSrv := &http.Server{Handler: r, Addr: fmt.Sprintf(":%d", cfg.HttpServer.Address)}
	go func() {
		logger.Info(fmt.Sprintf("HTTP server listening on :%d", cfg.HttpServer.Address))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to serve HTTP", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	logger.Info("Shutting down HTTP server...")
	if err := httpSrv.Shutdown(context.Background()); err != nil {
		logger.Error("HTTP server shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("HTTP server gracefully stopped")
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	logger := &slog.Logger{}

	_, err := fmt.Fprintf(w, "STATUS: OK")
	if err != nil {
		logger.Error("Failed to write healthcheck response", slog.Any("error", err))
		response := httpresponse.Response{
			Message: "Invalid request",
		}
		httpresponse.SendJSONResponse(r.Context(), w, response, http.StatusBadRequest, logger)
	}
}
