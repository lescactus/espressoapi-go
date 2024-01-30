package migrate

import (
	"github.com/spf13/cobra"
)

// SkipCmd represents the skip command
var SkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Sets the database level to the most recent version available, without running the migrations",
	Long: `Sets the database level to the most recent version available, without running the migrations. 
It is the equivalent of running "sql-migrate skip".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Todo: implement
	},
}
