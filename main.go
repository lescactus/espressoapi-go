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
	"github.com/lescactus/espressoapi-go/internal/logger"
	mysqlroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/roaster"
	mysqlsheet "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/sheet"

	svcroaster "github.com/lescactus/espressoapi-go/internal/services/roaster"
	svcsheet "github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/rs/zerolog/hlog"
)

func main() {
	// Get application configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("unable to build a new app config: %v", err)
	}

	logger := logger.New(
		cfg.LoggerLogLevel,
		cfg.LoggerDurationFieldUnit,
		cfg.LoggerFormat,
	)

	var sqlxdb *sqlx.DB
	switch cfg.DatabaseType {
	case config.DatabaseTypeMySQL:
		sqlxdb, err = sqlx.Connect(string(config.DatabaseTypeMySQL), cfg.DatabaseDatasourceName)
		if err != nil {
			logger.Fatal().Err(err).Msgf("unable to connect to %s", config.DatabaseTypeMySQL)
		}
	// Using mysql by default
	default:
		sqlxdb, err = sqlx.Connect(string(config.DatabaseTypeMySQL), cfg.DatabaseDatasourceName)
		if err != nil {
			logger.Fatal().Err(err).Msgf("unable to connect to %s", config.DatabaseTypeMySQL)
		}
	}

	dbSheet := mysqlsheet.New(sqlxdb)
	dbRoaster := mysqlroaster.New(sqlxdb)

	svcSheet := svcsheet.New(dbSheet)
	svcRoaster := svcroaster.New(dbRoaster)

	// Create http router, server and handler controller
	r := httprouter.New()
	h := controllers.NewHandler(svcSheet, svcRoaster, cfg.ServerMaxRequestSize)
	c := alice.New()
	s := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       cfg.ServerReadTimeout,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
	}

	// logger fields
	*logger = logger.With().Str("svc", config.AppName).Logger()

	// Register logging middleware
	c = c.Append(hlog.NewHandler(*logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.ProtoHandler("proto"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RemoteAddrHandler("remote_client"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RequestIDHandler("req_id", "X-Request-ID"))
	c = c.Append(h.IdParameterLoggerHandler("id"))
	//c = c.Append(h.IdParameterLoggerHandler(h, "id"))
	c = c.Append(h.MaxReqSize())

	r.Handler(http.MethodGet, "/ping", c.ThenFunc(h.Ping))
	r.Handler(http.MethodPost, "/rest/v1/sheets", c.ThenFunc(h.CreateSheet))
	r.Handler(http.MethodGet, "/rest/v1/sheets/:id", c.ThenFunc(h.GetSheetById))
	r.Handler(http.MethodGet, "/rest/v1/sheets", c.ThenFunc(h.GetAllSheets))
	r.Handler(http.MethodPut, "/rest/v1/sheets/:id", c.ThenFunc(h.UpdateSheetById))
	r.Handler(http.MethodDelete, "/rest/v1/sheets/:id", c.ThenFunc(h.DeleteSheetById))

	r.Handler(http.MethodPost, "/rest/v1/roasters", c.ThenFunc(h.CreateRoaster))
	r.Handler(http.MethodGet, "/rest/v1/roasters/:id", c.ThenFunc(h.GetRoasterById))
	r.Handler(http.MethodGet, "/rest/v1/roasters", c.ThenFunc(h.GetAllRoasters))
	r.Handler(http.MethodPut, "/rest/v1/roasters/:id", c.ThenFunc(h.UpdateRoasterById))
	r.Handler(http.MethodDelete, "/rest/v1/roasters/:id", c.ThenFunc(h.DeleteRoasterById))

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
