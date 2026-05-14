package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	resolveRaw  bool
	resolveJSON bool
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <name>",
	Short: "Resolve a popular law name to its (year, number) ID",
	Long: `Resolve a popular Danish law name (e.g. "databeskyttelsesloven",
"forvaltningsloven", "købeloven") to the canonical (year, number) ID. Useful
for chaining: 'retscli get $(retscli resolve --short databeskyttelsesloven)'.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		client := rets.NewClient()

		if resolveRaw {
			raw, err := client.ResolveRaw(query)
			if err != nil {
				return err
			}
			return writePrettyJSON(raw)
		}
		m, err := client.Resolve(query)
		if err != nil {
			return err
		}
		if resolveJSON {
			return encodeJSON(m)
		}
		if m == nil || m.Number == 0 {
			fmt.Fprintf(os.Stdout, "No match for %q. Try 'retscli search %q'.\n", query, query)
			return nil
		}
		fmt.Fprintf(os.Stdout, "%d/%d  %s — %s",
			m.Year, m.Number, m.DocumentType, m.ShortName)
		if m.PopularTitle != "" {
			fmt.Fprintf(os.Stdout, "  (%s)", m.PopularTitle)
		}
		fmt.Fprintf(os.Stdout, "  [confidence %.2f]\n", m.Confidence)
		return nil
	},
}

func init() {
	resolveCmd.Flags().BoolVar(&resolveRaw, "raw", false, "print the raw API response as JSON")
	resolveCmd.Flags().BoolVar(&resolveJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(resolveCmd)
}

var _ rets.ResolveMatch // keep types import lively
