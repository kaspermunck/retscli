package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var queryParams []string

var queryCmd = &cobra.Command{
	Use:   "query <path>",
	Short: "Raw passthrough to any /v1/ endpoint (power users)",
	Long: `Pass through to an arbitrary endpoint under /v1/ of the Retsinformation API.
Use this to reach endpoints not yet covered by a first-class subcommand
(e.g. /lovgivning/{year}/{number}/timeline, /lovgivning/cases/, actor
relationships, paragraph stk navigation, etc.). The response body is
pretty-printed as JSON.

Examples:
  retscli query /lovgivning/2018/502/timeline
  retscli query /lovgivning/cases/ --param search=klima --param limit=5
  retscli query /lovgivning/keywords/ --param limit=10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		v := url.Values{}
		for _, p := range queryParams {
			parts := strings.SplitN(p, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --param %q: expected key=value", p)
			}
			v.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
		client := rets.NewClient()
		raw, err := client.QueryRaw(path, v)
		if err != nil {
			return err
		}
		return writePrettyJSON(raw)
	},
}

func init() {
	queryCmd.Flags().StringArrayVar(&queryParams, "param", nil, "query parameter key=value (repeatable)")
	rootCmd.AddCommand(queryCmd)
}
