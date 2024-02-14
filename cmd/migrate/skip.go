package migrate

import (
	"github.com/lescactus/espressoapi-go/cmd/app"
	sqlmigrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

// SkipCmd represents the skip command
var SkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Sets the database level to the most recent version available, without running the migrations",
	Long: `Sets the database level to the most recent version available, without running the migrations. 
It is the equivalent of running "sql-migrate skip".`,
	Run: func(cmd *cobra.Command, args []string) {
		source := sqlmigrate.EmbedFileSystemMigrationSource{
			FileSystem: *app.App.MigrationsFS,
			Root:       "migrations/sql/mysql",
		}

		n, err := sqlmigrate.SkipMax(app.App.Db.DB, string(app.App.Cfg.DatabaseType), source, sqlmigrate.Up, limit)
		if err != nil {
			app.App.Logger.Fatal().Err(err).Msg("Failed to apply (down) migrations")
		}

		switch n {
		case 0:
			app.App.Logger.Info().Msg("All migrations have already been applied")
		case 1:
			app.App.Logger.Info().Msg("Skipped 1 migration")
		default:
			app.App.Logger.Info().Msgf("Skipped %d migrations", n)
		}
	},
}

func init() {
	SkipCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of migrations (0 = unlimited)")
}
