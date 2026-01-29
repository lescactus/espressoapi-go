package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/cmd/app"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/lescactus/espressoapi-go/internal/controllers/rest"
	"github.com/lescactus/espressoapi-go/internal/controllers/web"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/cobra"

	mysqlbean "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/bean"
	mysqlroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/roaster"
	mysqlsheet "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/sheet"
	mysqlshot "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/shot"

	svcbean "github.com/lescactus/espressoapi-go/internal/services/bean"
	svcroaster "github.com/lescactus/espressoapi-go/internal/services/roaster"
	svcsheet "github.com/lescactus/espressoapi-go/internal/services/sheet"
	svcshot "github.com/lescactus/espressoapi-go/internal/services/shot"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"server", "srv", "http"},
	Short:   "Run the http api and web server",
	Long: `Run the http api and web server. Configuration is read either from config files or environment variables.
Available configuration files are:

- config.json
- config.yaml
- config.env

Refer to the documentation for more information about configuration files and environment variables.
`,
	Run: runCmdMain,
}

func runCmdMain(cmd *cobra.Command, args []string) {
	dbSheet := mysqlsheet.New(app.App.Db)
	dbRoaster := mysqlroaster.New(app.App.Db)
	dbBean := mysqlbean.New(app.App.Db)
	dbShot := mysqlshot.New(app.App.Db)

	svcSheet := svcsheet.New(dbSheet)
	svcRoaster := svcroaster.New(dbRoaster)
	svcBean := svcbean.New(dbBean)
	svcShot := svcshot.New(dbShot)

	// Create http router, server and handler controller
	r := httprouter.New()
	rest := rest.NewHandler(svcSheet, svcRoaster, svcBean, svcShot, app.App.Cfg.ServerMaxRequestSize)
	web := web.NewHandler(svcSheet, svcRoaster, svcBean, svcShot)
	c := alice.New()
	s := &http.Server{
		Addr:              app.App.Cfg.ServerAddr,
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       app.App.Cfg.ServerReadTimeout,
		ReadHeaderTimeout: app.App.Cfg.ServerReadHeaderTimeout,
		WriteTimeout:      app.App.Cfg.ServerWriteTimeout,
	}

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
	c = c.Append(rest.IdParameterLoggerHandler("id"))
	c = c.Append(rest.MaxReqSize())

	r.Handler(http.MethodGet, "/ping", c.ThenFunc(rest.Ping))

	// REST API handlers
	r.Handler(http.MethodPost, "/rest/v1/sheets", c.ThenFunc(rest.CreateSheet))
	r.Handler(http.MethodGet, "/rest/v1/sheets/:id", c.ThenFunc(rest.GetSheetById))
	r.Handler(http.MethodGet, "/rest/v1/sheets", c.ThenFunc(rest.GetAllSheets))
	r.Handler(http.MethodPut, "/rest/v1/sheets/:id", c.ThenFunc(rest.UpdateSheetById))
	r.Handler(http.MethodDelete, "/rest/v1/sheets/:id", c.ThenFunc(rest.DeleteSheetById))

	r.Handler(http.MethodPost, "/rest/v1/roasters", c.ThenFunc(rest.CreateRoaster))
	r.Handler(http.MethodGet, "/rest/v1/roasters/:id", c.ThenFunc(rest.GetRoasterById))
	r.Handler(http.MethodGet, "/rest/v1/roasters", c.ThenFunc(rest.GetAllRoasters))
	r.Handler(http.MethodPut, "/rest/v1/roasters/:id", c.ThenFunc(rest.UpdateRoasterById))
	r.Handler(http.MethodDelete, "/rest/v1/roasters/:id", c.ThenFunc(rest.DeleteRoasterById))

	r.Handler(http.MethodPost, "/rest/v1/beans", c.ThenFunc(rest.CreateBeans))
	r.Handler(http.MethodGet, "/rest/v1/beans/:id", c.ThenFunc(rest.GetBeansById))
	r.Handler(http.MethodGet, "/rest/v1/beans", c.ThenFunc(rest.GetAllBeans))
	r.Handler(http.MethodPut, "/rest/v1/beans/:id", c.ThenFunc(rest.UpdateBeanById))
	r.Handler(http.MethodDelete, "/rest/v1/beans/:id", c.ThenFunc(rest.DeleteBeansById))

	r.Handler(http.MethodPost, "/rest/v1/shots", c.ThenFunc(rest.CreateShot))
	r.Handler(http.MethodGet, "/rest/v1/shots/:id", c.ThenFunc(rest.GetShotById))
	r.Handler(http.MethodGet, "/rest/v1/shots", c.ThenFunc(rest.GetAllShots))
	r.Handler(http.MethodPut, "/rest/v1/shots/:id", c.ThenFunc(rest.UpdateShotById))
	r.Handler(http.MethodDelete, "/rest/v1/shots/:id", c.ThenFunc(rest.DeleteShotById))

	// Swagger handlers
	redocOpts := middleware.RedocOpts{Path: "redoc", SpecURL: "swagger.json"}
	swaggerUiOpts := middleware.SwaggerUIOpts{Path: "swagger", SpecURL: "swagger.json"}
	r.Handler(http.MethodGet, "/redoc", middleware.Redoc(redocOpts, nil))
	r.Handler(http.MethodGet, "/swagger", middleware.SwaggerUI(swaggerUiOpts, nil))
	r.Handler(http.MethodGet, "/swagger.json", c.ThenFunc(rest.Swagger))

	// Web handlers
	r.Handler(http.MethodGet, "/", c.ThenFunc(web.Index))

	r.Handler(http.MethodGet, "/roasters/get/:id", c.ThenFunc(web.GetRoasterById))

	r.Handler(http.MethodGet, "/roasters/update/:id", c.ThenFunc(web.UpdateRoasterById))
	r.Handler(http.MethodPut, "/roasters/update/:id", c.ThenFunc(web.UpdateRoasterById))

	r.Handler(http.MethodGet, "/roasters", c.ThenFunc(web.GetRoasters))

	r.Handler(http.MethodPost, "/roasters", c.ThenFunc(web.CreateRoaster))
	r.Handler(http.MethodGet, "/roasters/add", c.ThenFunc(web.CreateRoaster))

	r.Handler(http.MethodDelete, "/roasters/delete/:id", c.ThenFunc(web.DeleteRoasterById))

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
