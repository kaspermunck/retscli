package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "retscli",
	Short: "CLI for querying Danish legal information via the Retsinformation API",
	Long: `CLI for querying Danish legal documents (love, bekendtgørelser, cirkulærer,
vejledninger) and Folketinget bills via the Retsinformation REST API at
retsinformation-api.dk.

Commands: search, get, list, recent, history, resolve, bills, query.

The default output is a human-readable table. Every data-returning command
supports --raw (raw upstream JSON body) and --json (parsed struct, pretty JSON).
No authentication is required.`,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
