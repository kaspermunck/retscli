package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	listType    string
	listRessort string
	listYear    int
	listSort    string
	listLimit   int
	listSkip    int
	listRaw     bool
	listJSON    bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents filtered by type, year, or ministry",
	Long: `List Danish legal documents filtered by document type, year, or ministry.
Use --type to restrict to a single category (LOV, LBK, BEK, CIR, VEJ, ...);
combine with --year for a single year's output, or --ressort for a ministry.
Same underlying endpoint as 'search', but without a free-text query.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listType == "" && listYear == 0 && listRessort == "" {
			return fmt.Errorf("provide at least one of --type, --year, --ressort (or use 'retscli recent' / 'retscli search')")
		}
		client := rets.NewClient()
		opts := rets.SearchOpts{
			Year:         listYear,
			Ressort:      listRessort,
			DocumentType: listType,
			Sort:         listSort,
			Limit:        listLimit,
			Skip:         listSkip,
		}

		if listRaw {
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
		if listJSON {
			return encodeJSON(resp)
		}
		printLawList(resp, "")
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listType, "type", "", "exact type code: LOV, LOVH, LBK, LBKH, BEK, CIR, VEJ, SKR, KEN (consolidations live under LBKH for older entries, LBK for current)")
	listCmd.Flags().StringVar(&listRessort, "ressort", "", "filter by ministry (free text)")
	listCmd.Flags().IntVar(&listYear, "year", 0, "filter by publication year")
	listCmd.Flags().StringVar(&listSort, "sort", "publication_date", "sort field")
	listCmd.Flags().IntVar(&listLimit, "limit", 50, "max number of results")
	listCmd.Flags().IntVar(&listSkip, "skip", 0, "pagination offset")
	listCmd.Flags().BoolVar(&listRaw, "raw", false, "print the raw API response as JSON")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(listCmd)
}

// printLawList declared in search.go is reused here.
var _ = os.Stdout // keep import in case future helpers move
