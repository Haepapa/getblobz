package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "getblobz",
	Short: "getblobz is a fast, stateful Azure Blob sync tool",
	Long:  `A CLI tool for syncing files from Azure Blob Storage to local storage efficiently.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}