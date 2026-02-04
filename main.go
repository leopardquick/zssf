package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/leopardquick/zssf/handler"
	"github.com/leopardquick/zssf/setup"
	"github.com/leopardquick/zssf/store"

	_ "github.com/lib/pq"
)

const (
	serverAddr     = ":2080"
	shutdownTimout = 10 * time.Second
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	dsn := setup.DatabaseDSN()
	if dsn == "" {
		logger.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open(setup.DatabaseDriver(), dsn)
	if err != nil {
		logger.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(ctxPing); err != nil {
		logger.Fatalf("failed to ping database: %v", err)
	}

	requestLogStore := store.NewSQLRequestLogStore(db)
	accountStore := store.NewSQLAccountStore(db)
	apiHandler := handler.New(&http.Client{Timeout: 15 * time.Second}, requestLogStore, accountStore)
	controlNumberHandler := handler.NewControlNumberHandler(&http.Client{Timeout: 40 * time.Second}, requestLogStore, accountStore)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello from chi"))
	})

	router.Post("/account-balance", apiHandler.AccountBalance)
	router.Post("/control-number/enquire", controlNumberHandler.Enquire)
	router.Post("/control-number/payment", controlNumberHandler.PaymentPost)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("server listening on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Printf("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
		if err := server.Close(); err != nil {
			logger.Printf("server close failed: %v", err)
		}
	}

	logger.Printf("server stopped")
}
