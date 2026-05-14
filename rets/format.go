package rets

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// WriteTable renders rows as an aligned table using tabwriter. Same shape as
// dstcli's helper so the suite reads consistently.
func WriteTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	sep := make([]string, len(headers))
	for i, h := range headers {
		sep[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}

// WriteCSV renders rows as CSV with a header row.
func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(headers); err != nil {
		return err
	}
	if err := cw.WriteAll(rows); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

// Truncate shortens s to max runes, appending "…" if truncated.
func Truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// FormatLawTitle returns a one-line label suitable for human output.
// Prefers the popular title (e.g. "Databeskyttelsesloven") when set.
func FormatLawTitle(l LawSummary) string {
	if l.PopularTitle != "" {
		return l.PopularTitle
	}
	return l.Title
}

// CollectParagraphs walks a law structure and returns every Paragraph in
// document order. Useful for the table view of a law and for finding a
// requested paragraph number without re-walking the tree manually.
func CollectParagraphs(s Structure) []Paragraph {
	out := []Paragraph{}
	for _, g := range s.ParagraphGroups {
		out = append(out, g.Paragraphs...)
	}
	for _, ch := range s.Chapters {
		for _, g := range ch.ParagraphGroups {
			out = append(out, g.Paragraphs...)
		}
	}
	return out
}
