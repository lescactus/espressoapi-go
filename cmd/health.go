package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/lescactus/espressoapi-go/cmd/app"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:     "healthz",
	Aliases: []string{"health", "healthcheck", "check", "status", "ping"},
	Short:   "Run a healthcheck on the http api server",
	Long: `Run a healthcheck on the http api server. This command will check if the server is running and if the database connection is healthy.
This command is useful for healthchecks in container orchestration platforms like Kubernetes.
`,
	Run: healthCmdMain,
}

func healthCmdMain(cmd *cobra.Command, args []string) {
	*app.App.Logger = app.App.Logger.With().Str("svc", config.AppName).Logger()

	// Check if the database connection is healthy
	err := app.App.Db.Ping()
	if err != nil {
		app.App.Logger.Error().Err(err).Msg("database connection is not healthy")
		os.Exit(1)
	}

	app.App.Logger.Info().Msg("database connection is healthy")

	// Check if the server is running by sending a http request to the /ping endpoint
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", app.App.Cfg.ServerAddr))
	if err != nil {
		app.App.Logger.Error().Err(err).Msg("server is not healthy")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		app.App.Logger.Error().Int("status_code", resp.StatusCode).Msg("server is not returning 200 OK")
		os.Exit(1)
	}

	app.App.Logger.Info().Msg("server is running")
}
