// Package main is the entry point for the getblobz CLI application.
// getblobz is a tool for synchronising files from Azure Blob Storage to local filesystem.
package main

import (
	"github.com/haepapa/getblobz/cmd"
)

// Build-time variables set via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info for cmd package
	cmd.SetVersion(version, commit, date)
	cmd.Execute()
}
