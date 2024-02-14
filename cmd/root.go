package cmd

import (
	"embed"
	"io/fs"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/cmd/app"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/lescactus/espressoapi-go/internal/logger"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "espressoapi-go",
	Short: "Small API server used to keep track and take notes of pulling espresso shots.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(migrationFs *embed.FS, swaggerFs fs.FS) {
	app.App.MigrationsFS = migrationFs
	app.App.SwaggerFS = swaggerFs

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	app.App = &app.Application{}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(migrateCmd)

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("unable to build a new app config: %s", err)
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
			log.Fatalf("unable to connect to %s: %s", config.DatabaseTypeMySQL, err)
		}

	// Using mysql by default
	default:
		sqlxdb, err = sqlx.Connect(string(config.DatabaseTypeMySQL), cfg.DatabaseDatasourceName)
		if err != nil {
			log.Fatalf("unable to connect to %s: %s", config.DatabaseTypeMySQL, err)
		}
	}

	app.App.Db = sqlxdb
	app.App.Cfg = cfg
	app.App.Logger = logger
}
