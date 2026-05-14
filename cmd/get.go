package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kaspermunck/retscli/rets"
	"github.com/spf13/cobra"
)

var (
	getParagraph string
	getMarkdown  bool
	getInclude   string
	getRaw       bool
	getJSON      bool
)

var getCmd = &cobra.Command{
	Use:   "get <year>/<number>",
	Short: "Fetch a Danish law (or one paragraph) by year/number or ELI URI",
	Long: `Fetch a specific law version by its (year, number) pair. Accepts either
"2018/502" or the ELI URI "/eli/lta/2018/502".

By default prints the law metadata and the full body as plain text. Use
--paragraph <n> to extract a single section, --markdown for the rendered
markdown export from the API, --raw for the upstream JSON, or --json for the
parsed struct.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		year, number, err := parseYearNumber(args[0])
		if err != nil {
			return fmt.Errorf("%w (try 'retscli search' or 'retscli resolve' to find the right ID)", err)
		}
		client := rets.NewClient()

		// Paragraph-only fast path.
		if getParagraph != "" {
			if getRaw {
				raw, err := client.GetParagraphRaw(year, number, getParagraph)
				if err != nil {
					return err
				}
				return writePrettyJSON(raw)
			}
			p, err := client.GetParagraph(year, number, getParagraph)
			if err != nil {
				return err
			}
			if getJSON {
				return encodeJSON(p)
			}
			printParagraph(*p)
			return nil
		}

		// Markdown export — the API returns markdown, not JSON.
		if getMarkdown {
			md, err := client.GetMarkdown(year, number, "", "")
			if err != nil {
				return err
			}
			_, _ = os.Stdout.Write(md)
			fmt.Fprintln(os.Stdout)
			return nil
		}

		if getRaw {
			raw, err := client.GetRaw(year, number, getInclude)
			if err != nil {
				return err
			}
			return writePrettyJSON(raw)
		}

		law, err := client.Get(year, number, getInclude)
		if err != nil {
			return err
		}
		if getJSON {
			return encodeJSON(law)
		}
		printLaw(*law)
		return nil
	},
}

func init() {
	getCmd.Flags().StringVar(&getParagraph, "paragraph", "", "extract only this paragraph (e.g. 7, 5a)")
	getCmd.Flags().BoolVar(&getMarkdown, "markdown", false, "fetch the law rendered as markdown (uses /markdown endpoint)")
	getCmd.Flags().StringVar(&getInclude, "include", "", "extra includes: case | actors | timeline | full")
	getCmd.Flags().BoolVar(&getRaw, "raw", false, "print the raw API response as JSON")
	getCmd.Flags().BoolVar(&getJSON, "json", false, "print the parsed result as JSON")
	rootCmd.AddCommand(getCmd)
}

func printLaw(l rets.Law) {
	w := os.Stdout
	title := l.Title
	if l.PopularTitle != "" {
		title = fmt.Sprintf("%s (%s)", l.Title, l.PopularTitle)
	}
	fmt.Fprintln(w, title)
	fmt.Fprintln(w, strings.Repeat("=", min(80, len(title))))

	fmt.Fprintf(w, "ID:         %d/%d\n", l.Year, l.Number)
	if l.ShortName != "" {
		fmt.Fprintf(w, "Reference:  %s\n", l.ShortName)
	}
	if l.DocumentType != "" {
		fmt.Fprintf(w, "Type:       %s\n", l.DocumentType)
	}
	if l.Ministry != "" {
		fmt.Fprintf(w, "Ressort:    %s\n", l.Ministry)
	} else if l.Ressort != "" {
		fmt.Fprintf(w, "Ressort:    %s\n", l.Ressort)
	}
	if l.SignatureDate != "" {
		fmt.Fprintf(w, "Signed:     %s\n", l.SignatureDate)
	}
	if l.PublicationDate != "" {
		fmt.Fprintf(w, "Published:  %s\n", l.PublicationDate)
	}
	if l.EffectiveDate != "" {
		fmt.Fprintf(w, "In force:   %s\n", l.EffectiveDate)
	}
	if l.ELIURI != "" {
		fmt.Fprintf(w, "ELI:        %s\n", l.ELIURI)
	}
	if l.IsConsolidated {
		fmt.Fprintln(w, "Status:     Consolidated (includes amendments)")
	} else if l.Historical {
		fmt.Fprintln(w, "Status:     Historical (superseded — see 'retscli history' for the current version)")
	}

	if len(l.Structure.Preamble) > 0 {
		fmt.Fprintln(w)
		for _, line := range l.Structure.Preamble {
			fmt.Fprintln(w, line)
		}
	}

	// Walk chapters, then top-level paragraph groups.
	for _, ch := range l.Structure.Chapters {
		fmt.Fprintln(w)
		head := strings.TrimSpace(ch.ChapterNumber + " " + ch.ChapterTitle)
		fmt.Fprintln(w, head)
		fmt.Fprintln(w, strings.Repeat("-", min(80, len(head))))
		for _, g := range ch.ParagraphGroups {
			if g.Heading != "" {
				fmt.Fprintln(w)
				fmt.Fprintln(w, g.Heading)
			}
			for _, p := range g.Paragraphs {
				printParagraph(p)
			}
		}
	}
	for _, g := range l.Structure.ParagraphGroups {
		if g.Heading != "" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, g.Heading)
		}
		for _, p := range g.Paragraphs {
			printParagraph(p)
		}
	}
}

func printParagraph(p rets.Paragraph) {
	w := os.Stdout
	fmt.Fprintln(w)
	fmt.Fprintln(w, p.Number)
	for _, s := range p.Stk {
		// Single-stk paragraphs often have an empty/blank stk number — drop the prefix in that case.
		prefix := s.Number
		if prefix != "" {
			fmt.Fprintf(w, "  %s %s\n", prefix, s.Text)
		} else {
			fmt.Fprintf(w, "  %s\n", s.Text)
		}
		for _, l := range s.Litra {
			fmt.Fprintf(w, "    %s) %s\n", l.Letter, l.Text)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
