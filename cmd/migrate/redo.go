package migrate

import (
	"github.com/spf13/cobra"
)

// RedoCmd represents the redo command
var RedoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Reapply the last migration",
	Long: `Reapply the last migration. It is the equivalent of running "down" and then "up" as would
"sql-migrate redo" do.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Todo: implement
	},
}
