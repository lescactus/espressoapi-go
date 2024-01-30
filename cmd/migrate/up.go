package migrate

import (
	"github.com/lescactus/espressoapi-go/cmd/app"
	sqlmigrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

// UpCmd represents the up command
var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrates the database to the most recent version available",
	Long: `Migrates the database to the most recent version available.
It is the equivalent of running "sql-migrate up".`,
	Run: func(cmd *cobra.Command, args []string) {
		n, err := execMigrations(sqlmigrate.Up)
		if err != nil {
			app.App.Logger.Fatal().Err(err).Msg("Failed to apply (up) migrations")
		}
		app.App.Logger.Info().Msgf("Successfully applied (up) %d migration(s)!", n)
	},
}

func init() {
	UpCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of migrations (0 = unlimited)")
	UpCmd.Flags().Int64VarP(&version, "version", "v", -1, "Run migrate up to a specific version, eg: the version number of migration 1_initial.sql is 1")
}
