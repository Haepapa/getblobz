package cmd

import (
	"fmt"
	"os"

	"github.com/Haepapa/getblobz/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg	 	*config.Config
)

var rootCmd = &cobra.Command{
	Use:   "getblobz",
	Short: "getblobz is a fast, stateful Azure Blob sync tool",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Printf("✓ Config Loaded: Container=%s, Workers=%d\n", cfg.ContainerName, cfg.Workers)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// global flag for config file if users want to specify a path
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./getblobz.yaml)")
}