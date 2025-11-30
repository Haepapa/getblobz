// Package cmd implements the CLI commands for getblobz.
package cmd

import (
	"fmt"
	"os"

	"github.com/haepapa/getblobz/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     *config.Config

	// Version information
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion sets the version information from build-time variables
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "getblobz",
	Short: "Azure Blob Storage sync tool",
	Long: `getblobz is a CLI tool for synchronising files from Azure Blob Storage to local filesystem.

It supports multiple authentication methods, incremental sync, continuous monitoring,
and adaptive performance tuning for diverse hardware platforms.`,
	Version: "dev",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./getblobz.yaml or ~/.config/getblobz/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "log format (text, json)")

	if err := viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind log-level flag: %v\n", err)
	}
	if err := viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind log-format flag: %v\n", err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfg = config.Default()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		discoveredPath := config.GetConfigPath("")
		if discoveredPath != "" {
			viper.SetConfigFile(discoveredPath)
		} else {
			viper.AddConfigPath(".")
			viper.AddConfigPath("$HOME/.config/getblobz")
			viper.SetConfigName("getblobz")
			viper.SetConfigType("yaml")
		}
	}

	viper.SetEnvPrefix("getblobz")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		}
	}
}
