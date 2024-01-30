package migrate

import (
	"github.com/spf13/cobra"
)

// UpCmd represents the up command
var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrates the database to the most recent version available",
	Long: `Migrates the database to the most recent version available.
It is the equivalent of running "sql-migrate up".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Todo: implement
	},
}
