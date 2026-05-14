package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	historyParagraph string
	historyRaw       bool
	historyJSON      bool
)

var historyCmd = &cobra.Command{
	Use:   "history <year>/<number>",
	Short: "Show the consolidation/amendment chain for a law",
	Long: `Show the unified history of a law across every consolidation (LBK) and
amendment. Use --paragraph to narrow the timeline to changes affecting a
specific section.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		year, number, err := parseYearNumber(args[0])
		if err != nil {
			return err
		}
		client := rets.NewClient()

		if historyRaw {
			raw, err := client.HistoryRaw(year, number, historyParagraph)
			if err != nil {
				return err
			}
			return writePrettyJSON(raw)
		}
		h, err := client.History(year, number, historyParagraph)
		if err != nil {
			return err
		}
		if historyJSON {
			return encodeJSON(h)
		}
		printHistory(*h)
		return nil
	},
}

func init() {
	historyCmd.Flags().StringVar(&historyParagraph, "paragraph", "", "narrow the history to one paragraph (e.g. 7)")
	historyCmd.Flags().BoolVar(&historyRaw, "raw", false, "print the raw API response as JSON")
	historyCmd.Flags().BoolVar(&historyJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(historyCmd)
}

func printHistory(h rets.History) {
	w := os.Stdout
	fmt.Fprintln(w, h.LawTitle)
	fmt.Fprintf(w, "%d entries (%d consolidations, %d amendments)", h.TotalEntries, h.ConsolidationCount, h.AmendmentCount)
	if h.FirstDate != "" || h.LastDate != "" {
		fmt.Fprintf(w, ", %s → %s", h.FirstDate, h.LastDate)
	}
	if h.FilteredByParagraf != "" {
		fmt.Fprintf(w, ", filtered by %s", h.FilteredByParagraf)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	headers := []string{"#", "Valid from", "Valid to", "ID", "Type", "Entry", "Title"}
	rows := make([][]string, len(h.Entries))
	for i, e := range h.Entries {
		validUntil := e.ValidUntil
		if validUntil == "" {
			validUntil = "(current)"
		}
		rows[i] = []string{
			fmt.Sprintf("%d", e.Sequence),
			e.ValidFrom,
			validUntil,
			fmt.Sprintf("%d/%d", e.Year, e.Number),
			e.DocumentType,
			e.EntryType,
			rets.Truncate(e.Title, 60),
		}
	}
	rets.WriteTable(w, headers, rows)
}
