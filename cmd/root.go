package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Envelope is the persistent flag controlling shared-envelope output across
// every subcommand. When true, results are wrapped in
// {source,kind,version,data,fetchedAt} regardless of --json/--raw.
var Envelope bool

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

func init() {
	rootCmd.PersistentFlags().BoolVar(&Envelope, "envelope", false, "wrap output in the shared {source,kind,version,data,fetchedAt} envelope (supersedes --json)")
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
