package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
	"github.com/yherda-opensource/yherda-cmd/internal/export"
)

var expressionCmd = &cobra.Command{
	Use:   "expression",
	Short: "Manage expressions",
	Long:  "An expression is an idea rendered into a specific template — the manuscript, screenplay, or one-sheet you actually export and read. Expressions belong to an idea.",
}

var expressionIdeaID string

var expressionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List expressions for an idea",
	Long:  "Lists expressions for an idea, showing each one's template name. Uses the active idea unless --idea is passed.",
	Example: `  yherda expression list
  yherda expression list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := expressionIdeaID
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			fmt.Println("No active idea — showing ideas instead. Run 'yherda idea use <id>' to select one.")
			return ideasListCmd.RunE(cmd, args)
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/storyline/"+ideaID+"/expressions/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tTEMPLATE\tLANGUAGE\tCREATED")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				strField(row, "id"),
				strField(row, "template_name"),
				strField(row, "language"),
				strField(row, "created"),
			)
		}
		w.Flush()
		printContext()
		return nil
	},
}

var expressionShowCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show an expression",
	Example: `  yherda expression show 4`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/expression/"+args[0]+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "idea", "template", "template_name", "language", "created"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		return nil
	},
}

// expressionPrintCmd is kept as a convenience alias for export --format stdout.
var expressionPrintCmd = &cobra.Command{
	Use:   "print <id>",
	Short: "Print expression content in reading order (equivalent to export --format stdout)",
	Long:  "Walks the expression's segment tree in reading order and prints each segment's content. A quick way to read a draft without exporting a file.",
	Example: `  yherda expression print 4
  yherda expression print 4 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roots, _, err := fetchSegmentTree(args[0])
		if err != nil {
			return err
		}
		if jsonOutput {
			type segEntry struct {
				Type    string `json:"type"`
				Number  string `json:"number"`
				Depth   int    `json:"depth"`
				Content string `json:"content"`
			}
			var entries []segEntry
			var walk func(nodes []export.SegmentNode, depth int)
			walk = func(nodes []export.SegmentNode, depth int) {
				for _, n := range nodes {
					entries = append(entries, segEntry{
						Type:    n.TypeName,
						Number:  n.Number,
						Depth:   depth,
						Content: n.Content,
					})
					walk(n.Children, depth+1)
				}
			}
			walk(roots, 0)
			printJSON(entries)
			return nil
		}
		e, _ := export.Get("stdout")
		if err := e.Export("", roots, ""); err != nil {
			return err
		}
		printContext()
		return nil
	},
}

var (
	exportFormat     string
	exportExprID     string
	exportOutputPath string
)

var expressionExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export an expression to a file format",
	Long: `Exports an expression's segment tree to a file in the given format.
Output defaults to <expression-id>.<ext> in the current directory unless
--output is passed.`,
	Example: `  yherda expression export --expression 4 --format scriv
  yherda expression export --expression 4 --format scriv --output draft.scrivx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if exportFormat == "" {
			return fmt.Errorf("--format is required (available: %s)", strings.Join(export.Formats(), ", "))
		}
		e, ok := export.Get(exportFormat)
		if !ok {
			return fmt.Errorf("unknown format %q (available: %s)", exportFormat, strings.Join(export.Formats(), ", "))
		}

		exprID := exportExprID
		if exprID == "" {
			return fmt.Errorf("no expression specified — use --expression <id>")
		}

		roots, title, err := fetchSegmentTree(exprID)
		if err != nil {
			return err
		}

		output := exportOutputPath
		if output == "" && e.DefaultExt() != "" {
			output = exprID + "." + e.DefaultExt()
		}

		return e.Export(title, roots, output)
	},
}

// fetchSegmentTree fetches an expression and its full nested segment tree,
// returning the parsed tree and the expression title.
func fetchSegmentTree(exprID string) ([]export.SegmentNode, string, error) {
	client := mustClient()

	var expr map[string]any
	if err := client.Get("/expression/"+exprID+"/", &expr); err != nil {
		return nil, "", err
	}

	templateID := strField(expr, "template")
	if templateID == "" {
		return nil, "", fmt.Errorf("expression has no template")
	}

	title := strField(expr, "id")
	if t := strField(expr, "title"); t != "" {
		title = t
	}

	var allSegments []map[string]any
	if err := client.Get("/expressiontemplate/"+templateID+"/segments/?root=true", &allSegments); err != nil {
		return nil, "", fmt.Errorf("could not fetch segments: %w", err)
	}

	return export.BuildTree(allSegments), title, nil
}

func init() {
	expressionListCmd.Flags().StringVar(&expressionIdeaID, "idea", "", "Idea ID (overrides active context)")
	expressionExportCmd.Flags().StringVar(&exportFormat, "format", "", "Export format (required): "+strings.Join(export.Formats(), ", "))
	expressionExportCmd.Flags().StringVar(&exportExprID, "expression", "", "Expression ID (overrides active context)")
	expressionExportCmd.Flags().StringVar(&exportOutputPath, "output", "", "Output path (default: <expression-id>.<ext>)")
	expressionCmd.AddCommand(expressionListCmd, expressionShowCmd, expressionPrintCmd, expressionExportCmd)
}
