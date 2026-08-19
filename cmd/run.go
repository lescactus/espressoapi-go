package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/cmd/app"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/lescactus/espressoapi-go/internal/controllers/rest"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/cobra"

	svcbean "github.com/lescactus/espressoapi-go/internal/services/bean"
	svcroaster "github.com/lescactus/espressoapi-go/internal/services/roaster"
	svcsheet "github.com/lescactus/espressoapi-go/internal/services/sheet"
	svcshot "github.com/lescactus/espressoapi-go/internal/services/shot"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"server", "srv", "http"},
	Short:   "Run the http api server",
	Long: `Run the http api server. Configuration is read either from config files or environment variables.
Available configuration files are:

- config.json
- config.yaml
- config.env

Refer to the documentation for more information about configuration files and environment variables.
`,
	Run: runCmdMain,
}

func runCmdMain(cmd *cobra.Command, args []string) {
	repositories, err := newRepositorySet(app.App.Cfg.DatabaseType, app.App.Db)
	if err != nil {
		log.Fatalf("unable to create repositories: %s", err)
	}

	svcSheet := svcsheet.New(repositories.sheet)
	svcRoaster := svcroaster.New(repositories.roaster)
	svcBean := svcbean.New(repositories.beans)
	svcShot := svcshot.New(repositories.shot)

	// Create handler and middleware chain
	h := rest.NewHandler(svcSheet, svcRoaster, svcBean, svcShot, app.App.Cfg.ServerMaxRequestSize)
	c := alice.New()

	// Logger fields
	*app.App.Logger = app.App.Logger.With().Str("svc", config.AppName).Logger()

	// Register logging middleware
	c = c.Append(hlog.NewHandler(*app.App.Logger))

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
	c = c.Append(h.MaxReqSize())

	// Build the route table and server
	r := newRouter(h, c)
	s := &http.Server{
		Addr:              app.App.Cfg.ServerAddr,
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       app.App.Cfg.ServerReadTimeout,
		ReadHeaderTimeout: app.App.Cfg.ServerReadHeaderTimeout,
		WriteTimeout:      app.App.Cfg.ServerWriteTimeout,
	}

	// Start server
	go func() {
		app.App.Logger.Info().Msg("Starting server ...")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.App.Logger.Fatal().Err(err).Msg("Could not start server on port " + app.App.Cfg.ServerAddr)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Blocking until receiving a shutdown signal
	sig := <-sigChan

	app.App.Logger.Info().Str("signal", sig.String()).Msg("Server received signal. Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	// Attempting to gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		app.App.Logger.Err(err).Msg("Failed to gracefully shutdown the server")
	}
}
