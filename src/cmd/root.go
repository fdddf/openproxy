package cmd

import (
	"fmt"
	"os"

	"github.com/fdddf/openproxy/pkg/version"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	showVersion bool

	rootCmd = &cobra.Command{
		Use:   "gptproxy",
		Short: "GPT Proxy server and utilities",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if showVersion {
				cmd.Printf("gptproxy %s\n", version.Info())
				os.Exit(0)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), configPath)
		},
	}
)

// Execute runs the root cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Print version information and exit")

	rootCmd.AddCommand(
		newServeCmd(),
		newMigrateCmd(),
		newPasswordCmd(),
		newGenDAOCmd(),
		newRefreshTokenCmd(),
	)
}
