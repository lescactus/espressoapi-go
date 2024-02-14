package migrate

import (
	"github.com/lescactus/espressoapi-go/cmd/app"
	sqlmigrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

// RedoCmd represents the redo command
var RedoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Reapply the last migration",
	Long: `Reapply the last migration. It is the equivalent of running "down" and then "up" as would
"sql-migrate redo" do.`,
	Run: func(cmd *cobra.Command, args []string) {
		source := sqlmigrate.EmbedFileSystemMigrationSource{
			FileSystem: *app.App.MigrationsFS,
			Root:       "migrations/sql/mysql",
		}

		migrations, _, err := sqlmigrate.PlanMigration(app.App.Db.DB, string(app.App.Cfg.DatabaseType), source, sqlmigrate.Down, 1)
		if err != nil {
			app.App.Logger.Fatal().Err(err).Msg("Failed to reapply migrations")
		} else if len(migrations) == 0 {
			app.App.Logger.Info().Msg("Nothing to do!")
			return
		}

		_, err = sqlmigrate.ExecMax(app.App.Db.DB, string(app.App.Cfg.DatabaseType), source, sqlmigrate.Down, 1)
		if err != nil {
			app.App.Logger.Fatal().Err(err).Msg("Migration (down) failed")
		}

		_, err = sqlmigrate.ExecMax(app.App.Db.DB, string(app.App.Cfg.DatabaseType), source, sqlmigrate.Up, 1)
		if err != nil {
			app.App.Logger.Fatal().Err(err).Msg("Migration (up) failed")
		}

		app.App.Logger.Info().Msgf("Successfully reapplied migration %s!", migrations[0].Id)
	},
}
