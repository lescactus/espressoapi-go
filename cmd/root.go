package cmd

import (
	"os"

	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/spf13/cobra"
)

// Application configuration
var cfg *config.App

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "espressoapi-go",
	Short: "Small API server used to keep track and take notes of pulling espresso shots.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(migrateCmd)

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var err error
	cfg, err = config.New()
	if err != nil {
		panic(err)
	}
}
