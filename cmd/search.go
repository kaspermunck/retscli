package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	searchType    string
	searchRessort string
	searchYear    int
	searchSort    string
	searchLimit   int
	searchSkip    int
	searchHist    bool
	searchRaw     bool
	searchJSON    bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Full-text search across Danish legal documents",
	Long: `Search across all Danish legal documents on Retsinformation by free text in
the title. Combine with --type (LOV, LBK, BEK, CIR, VEJ, ...), --year, and
--ressort to narrow the result set. Results are sorted by publication date
(newest first) unless --sort is given.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		for i, a := range args {
			if i > 0 {
				query += " "
			}
			query += a
		}
		client := rets.NewClient()
		opts := rets.SearchOpts{
			Query:        query,
			Year:         searchYear,
			Ressort:      searchRessort,
			DocumentType: searchType,
			Sort:         searchSort,
			Limit:        searchLimit,
			Skip:         searchSkip,
		}
		if cmd.Flags().Changed("historical") {
			opts.Historical = &searchHist
		}

		if searchRaw {
			raw, err := client.SearchRaw(opts)
			if err != nil {
				return err
			}
			return writePrettyJSON(raw)
		}

		resp, err := client.Search(opts)
		if err != nil {
			return err
		}
		if Envelope {
			return encodeJSON(rets.Wrap("LawSearch", resp))
		}
		if searchJSON {
			return encodeJSON(resp)
		}

		printLawList(resp, query)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchType, "type", "", "exact document type code: LOV (act), LOVH (historical act), LBK (consolidation), LBKH (historical consolidation), BEK, CIR, VEJ, SKR, KEN")
	searchCmd.Flags().StringVar(&searchRessort, "ressort", "", "filter by ministry (free text)")
	searchCmd.Flags().IntVar(&searchYear, "year", 0, "filter by publication year")
	searchCmd.Flags().StringVar(&searchSort, "sort", "", "sort field: signature_date | publication_date | effective_date | title | year | number")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 20, "max number of results")
	searchCmd.Flags().IntVar(&searchSkip, "skip", 0, "pagination offset")
	searchCmd.Flags().BoolVar(&searchHist, "historical", false, "if set, filter by historical=true/false (use --historical=false for current only)")
	searchCmd.Flags().BoolVar(&searchRaw, "raw", false, "print the raw API response as JSON")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(searchCmd)
}

func printLawList(resp *rets.LawList, query string) {
	if len(resp.Data) == 0 {
		fmt.Fprintln(os.Stdout, "No matches.")
		return
	}
	if resp.CanonicalMatch != nil && resp.CanonicalMatch.Number != 0 {
		m := resp.CanonicalMatch
		fmt.Fprintf(os.Stdout, "Canonical match: %d/%d  %s — %s (confidence %.2f)\n\n",
			m.Year, m.Number, m.DocumentType, m.ShortName, m.Confidence)
	}
	headers := []string{"ID", "Type", "Published", "Title"}
	rows := make([][]string, len(resp.Data))
	for i, l := range resp.Data {
		rows[i] = formatLawRow(l)
	}
	rets.WriteTable(os.Stdout, headers, rows)
	fmt.Fprintf(os.Stdout, "\nShowing %d of %d total matches", len(resp.Data), resp.Count)
	if query != "" {
		fmt.Fprintf(os.Stdout, " for %q", query)
	}
	fmt.Fprintln(os.Stdout, ".")
	fmt.Fprintln(os.Stdout, "Use 'retscli get <year>/<number>' for the full text.")
}
