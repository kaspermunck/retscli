package cmd

import (
	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	recentType  string
	recentLimit int
	recentRaw   bool
	recentJSON  bool
)

var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "List the most recently published documents",
	Long: `List the most recently published Danish legal documents, sorted by
publication_date descending. Use --type to narrow to a single category
(LOV, LBK, BEK, CIR, VEJ, ...).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := rets.NewClient()
		opts := rets.SearchOpts{
			DocumentType: recentType,
			Sort:         "-publication_date", // leading minus = DESC; API otherwise sorts ASC
			Limit:        recentLimit,
		}
		if recentRaw {
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
		if recentJSON {
			return encodeJSON(resp)
		}
		printLawList(resp, "")
		return nil
	},
}

func init() {
	recentCmd.Flags().StringVar(&recentType, "type", "", "exact type code: LOV, LOVH, LBK, LBKH, BEK, CIR, VEJ, SKR, KEN")
	recentCmd.Flags().IntVar(&recentLimit, "limit", 20, "max number of results")
	recentCmd.Flags().BoolVar(&recentRaw, "raw", false, "print the raw API response as JSON")
	recentCmd.Flags().BoolVar(&recentJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(recentCmd)
}
