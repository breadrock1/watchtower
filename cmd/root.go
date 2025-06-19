package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"watchtower/internal/infrastructure/config"
)

var serviceConfig *config.Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "./watchtower",
	Short: "Consume internal service to load endpoints from watcher directory",
	Long: `
		Consume internal service to load endpoints from watcher directory.
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		var parseErr error
		filePath, _ := cmd.Flags().GetString("config")
		serviceConfig, parseErr = config.FromFile(filePath)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() *config.Config {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	return serviceConfig
}

func init() {
	flags := rootCmd.Flags()
	flags.StringP("config", "c", "./configs/config.toml", "Parse options from config file.")
}
