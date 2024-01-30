package migrate

import (
	"github.com/spf13/cobra"
)

// StatusCmd represents the statuss command
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long: `Show migration status. 
It is the equivalent of running "sql-migrate status".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Todo: implement
	},
}
