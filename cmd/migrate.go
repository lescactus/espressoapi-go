package cmd

import (
	"github.com/lescactus/espressoapi-go/cmd/migrate"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute database migrations",
	Long:  `Apply, undo, skip or show the status of database migrations.`,
}

func init() {
	migrateCmd.AddCommand(migrate.UpCmd)
	migrateCmd.AddCommand(migrate.DownCmd)
	migrateCmd.AddCommand(migrate.RedoCmd)
	migrateCmd.AddCommand(migrate.SkipCmd)
	migrateCmd.AddCommand(migrate.StatusCmd)
}
