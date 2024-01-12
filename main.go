package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/lescactus/espressoapi-go/internal/controllers"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/mysql"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

func main() {
	// Get application configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("unable to build a new app config: %v", err)
	}

	var sqlxdb *sqlx.DB
	switch cfg.DatabaseType {
	case config.DatabaseTypeMySQL:
		sqlxdb, err = sqlx.Connect(string(config.DatabaseTypeMySQL), cfg.DatabaseDatasourceName)
		if err != nil {
			log.Fatalf("unable to connect to %s: %s", config.DatabaseTypeMySQL, err.Error())
		}
	// Using mysql by default
	default:
		sqlxdb, err = sqlx.Connect(string(config.DatabaseTypeMySQL), cfg.DatabaseDatasourceName)
		if err != nil {
			log.Fatalf("unable to connect to %s: %s", config.DatabaseTypeMySQL, err.Error())
		}
	}

	db := mysql.New(sqlxdb)

	sheetService := sheet.New(db)

	// Create http router, server and handler controller
	r := httprouter.New()
	h := controllers.NewHandler(sheetService, cfg.ServerMaxRequestSize)
	c := alice.New()
	s := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       cfg.ServerReadTimeout,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
	}

	c = c.Append(h.MaxReqSize())

	r.Handler(http.MethodGet, "/rest/v1/ping", c.ThenFunc(h.Ping))
	r.Handler(http.MethodPost, "/rest/v1/sheets", c.ThenFunc(h.CreateSheet))
	r.Handler(http.MethodGet, "/rest/v1/sheets/:id", c.ThenFunc(h.GetSheetById))
	r.Handler(http.MethodGet, "/rest/v1/sheets", c.ThenFunc(h.GetAllSheets))
	r.Handler(http.MethodPut, "/rest/v1/sheets/:id", c.ThenFunc(h.UpdateSheetById))
	r.Handler(http.MethodDelete, "/rest/v1/sheets/:id", c.ThenFunc(h.DeleteSheetById))

	// Start server
	go func() {
		log.Printf("Starting server ...\n")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Blocking until receiving a shutdown signal
	sig := <-sigChan

	log.Printf("Server received %s signal. Shutting down...\n", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	// Attempting to gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		log.Println("Failed to gracefully shutdown the server")
	}
}
