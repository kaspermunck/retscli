package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	billsSearch    string
	billsStatus    string
	billsRessort   string
	billsLimit     int
	billsSkip      int
	billsEnacted   bool
	billsBillNum   string
	billsRaw       bool
	billsJSON      bool
)

var billsCmd = &cobra.Command{
	Use:   "bills [query]",
	Short: "Search Folketinget bills (lovforslag)",
	Long: `Search legislative bills (Folketinget lovforslag) by title or fetch one by
its short number (e.g. "L 33"). Bills are the parliamentary track behind every
enacted law; use 'retscli history' on a law to see which bills produced it.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := rets.NewClient()

		// Single-bill mode: --number or positional that looks like one.
		number := billsBillNum
		if number == "" && len(args) == 1 && looksLikeBillNumber(args[0]) {
			number = args[0]
		}
		if number != "" {
			if billsRaw {
				raw, err := client.BillRaw(number)
				if err != nil {
					return err
				}
				return writePrettyJSON(raw)
			}
			b, err := client.Bill(number)
			if err != nil {
				return err
			}
			if billsJSON {
				return encodeJSON(b)
			}
			printBill(*b)
			return nil
		}

		query := billsSearch
		if query == "" && len(args) == 1 {
			query = args[0]
		}
		opts := rets.BillSearchOpts{
			Query:   query,
			Status:  billsStatus,
			Ressort: billsRessort,
			Limit:   billsLimit,
			Skip:    billsSkip,
		}
		if cmd.Flags().Changed("enacted") {
			opts.Enacted = &billsEnacted
		}

		if billsRaw {
			raw, err := client.BillsRaw(opts)
			if err != nil {
				return err
			}
			return writePrettyJSON(raw)
		}
		resp, err := client.Bills(opts)
		if err != nil {
			return err
		}
		if billsJSON {
			return encodeJSON(resp)
		}
		printBillList(*resp)
		return nil
	},
}

func init() {
	billsCmd.Flags().StringVar(&billsBillNum, "number", "", "fetch a specific bill by its short number (e.g. \"L 33\")")
	billsCmd.Flags().StringVar(&billsSearch, "search", "", "free-text search in bill titles")
	billsCmd.Flags().StringVar(&billsStatus, "status", "", "filter by status (Vedtaget, Forkastet, ...)")
	billsCmd.Flags().StringVar(&billsRessort, "ressort", "", "filter by ministry")
	billsCmd.Flags().IntVar(&billsLimit, "limit", 20, "max number of results")
	billsCmd.Flags().IntVar(&billsSkip, "skip", 0, "pagination offset")
	billsCmd.Flags().BoolVar(&billsEnacted, "enacted", false, "filter to bills that were enacted (combine: --enacted=true|false)")
	billsCmd.Flags().BoolVar(&billsRaw, "raw", false, "print the raw API response as JSON")
	billsCmd.Flags().BoolVar(&billsJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(billsCmd)
}

func looksLikeBillNumber(s string) bool {
	// "L 33", "B 12", "F 5" — the Folketinget bill prefixes.
	if len(s) < 3 {
		return false
	}
	first := s[0]
	return (first == 'L' || first == 'B' || first == 'F') && s[1] == ' '
}

func printBillList(resp rets.BillList) {
	w := os.Stdout
	if len(resp.Data) == 0 {
		fmt.Fprintln(w, "No bills found.")
		return
	}
	headers := []string{"Number", "Status", "Title"}
	rows := make([][]string, len(resp.Data))
	for i, b := range resp.Data {
		title := b.Title
		if b.ShortTitle != "" {
			title = b.ShortTitle
		}
		rows[i] = []string{b.Number, b.Status, rets.Truncate(title, 80)}
	}
	rets.WriteTable(w, headers, rows)
	fmt.Fprintf(w, "\nShowing %d of %d. Use 'retscli bills <L NNN>' for the full record.\n", len(resp.Data), resp.Count)
}

func printBill(b rets.Bill) {
	w := os.Stdout
	fmt.Fprintln(w, "Bill")
	fmt.Fprintf(w, "  Number:    %s\n", b.Number)
	fmt.Fprintf(w, "  FT ID:     %d\n", b.FTID)
	fmt.Fprintf(w, "  Status:    %s\n", b.Status)
	if b.DecisionDate != "" {
		fmt.Fprintf(w, "  Decided:   %s\n", b.DecisionDate)
	}
	if b.PeriodeTitle != "" {
		fmt.Fprintf(w, "  Period:    %s\n", b.PeriodeTitle)
	}
	if b.Ressort != "" {
		fmt.Fprintf(w, "  Ressort:   %s\n", b.Ressort)
	}
	if b.EnactedLawID != "" {
		fmt.Fprintf(w, "  Enacted:   %s\n", b.EnactedLawID)
	}
	fmt.Fprintln(w, "\nTitle")
	fmt.Fprintln(w, "  "+b.Title)
	if b.ShortTitle != "" {
		fmt.Fprintln(w, "\nShort title")
		fmt.Fprintln(w, "  "+b.ShortTitle)
	}
	if b.Resume != "" {
		fmt.Fprintln(w, "\nResumé")
		fmt.Fprintln(w, "  "+b.Resume)
	}
}
