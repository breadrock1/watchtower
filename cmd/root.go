package cmd

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var serviceConfig *Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "./watchtower",
	Short: "StartConsuming internal service to load endpoints from watcher directory",
	Long: `
		StartConsuming internal service to load endpoints from watcher directory.
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		var parseErr error

		dotEnvEnabled, err := cmd.Flags().GetBool("dotenv")
		if err == nil && dotEnvEnabled {
			err = godotenv.Load(".env")
			if err != nil {
				slog.Warn("failed to load .env file")
			}
		}

		serviceConfig, parseErr = InitConfig()
		if parseErr != nil {
			log.Fatalf("launch failed: %s", parseErr.Error())
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() *Config {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	return serviceConfig
}

func init() {
	flags := rootCmd.Flags()
	flags.BoolP("dotenv", "d", false, "load environment vars using dotenv")
}
