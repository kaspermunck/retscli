package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kaspermunck/retscli/rets"
)

// writePrettyJSON pretty-prints raw JSON to stdout, falling back to raw output
// if indentation fails. Used by --raw across all subcommands.
func writePrettyJSON(raw []byte) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		_, _ = os.Stdout.Write(raw)
		fmt.Fprintln(os.Stdout)
		return nil
	}
	_, _ = os.Stdout.Write(pretty.Bytes())
	fmt.Fprintln(os.Stdout)
	return nil
}

func encodeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// parseYearNumber accepts "<year>/<number>" or an ELI URI like
// "/eli/lta/2018/502" or "eli/lta/2018/502" and returns (year, number).
func parseYearNumber(id string) (int, int, error) {
	s := strings.TrimSpace(id)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimPrefix(s, "eli/lta/")
	s = strings.TrimPrefix(s, "eli/")
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("expected <year>/<number> or an ELI URI, got %q", id)
	}
	// Take the last two segments — handles both "2018/502" and longer ELIs.
	year, err := parseInt(parts[len(parts)-2])
	if err != nil {
		return 0, 0, fmt.Errorf("year: %w", err)
	}
	number, err := parseInt(parts[len(parts)-1])
	if err != nil {
		return 0, 0, fmt.Errorf("number: %w", err)
	}
	return year, number, nil
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range strings.TrimSpace(s) {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// formatLawRow returns the columns used by search/list/recent tables.
func formatLawRow(l rets.LawSummary) []string {
	title := l.Title
	if l.PopularTitle != "" {
		title = l.PopularTitle
	}
	return []string{
		fmt.Sprintf("%d/%d", l.Year, l.Number),
		l.DocumentType,
		l.PublicationDate,
		rets.Truncate(title, 80),
	}
}
