package migrate

import (
	"github.com/spf13/cobra"
)

// DownCmd represents the down command
var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Undo a database migration",
	Long: `Undo a database migration. 
It is the equivalent of running "sql-migrate down".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Todo: implement
	},
}
