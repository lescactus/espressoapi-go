package migrate

import (
	"github.com/lescactus/espressoapi-go/cmd/app"
	sqlmigrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

// DownCmd represents the down command
var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Undo a database migration",
	Long: `Undo a database migration. 
It is the equivalent of running "sql-migrate down".`,
	Run: func(cmd *cobra.Command, args []string) {
		n, err := execMigrations(sqlmigrate.Up)
		if err != nil {
			app.App.Logger.Fatal().Err(err).Msg("Failed to apply (down) migrations")
		}
		app.App.Logger.Info().Msgf("Successfully applied (down) %d migration(s)!", n)
	},
}

func init() {
	DownCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of migrations (0 = unlimited)")
	DownCmd.Flags().Int64VarP(&version, "version", "v", -1, "Run migrate up to a specific version, eg: the version number of migration 1_initial.sql is 1")
}
