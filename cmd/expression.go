package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var expressionCmd = &cobra.Command{
	Use:   "expression",
	Short: "Manage expressions",
}

var expressionIdeaID string

var expressionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List expressions for an idea",
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
		if err := client.Get("/expression/?idea="+ideaID, &result); err != nil {
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
				strField(row, "template"),
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
	Use:   "show <id>",
	Short: "Show an expression",
	Args:  cobra.ExactArgs(1),
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
		for _, key := range []string{"id", "idea", "template", "language", "created"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		return nil
	},
}

var expressionPrintCmd = &cobra.Command{
	Use:   "print <id>",
	Short: "Print expression content in reading order (depth-first)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()

		var expr map[string]any
		if err := client.Get("/expression/"+args[0]+"/", &expr); err != nil {
			return err
		}

		templateID := strField(expr, "template")
		if templateID == "" {
			return fmt.Errorf("expression has no template")
		}

		// Fetch root segments with their full nested children tree in one call.
		var allSegments []map[string]any
		if err := client.Get("/segment/?template="+templateID+"&root=true", &allSegments); err != nil {
			return fmt.Errorf("could not fetch segments: %w", err)
		}

		// Find roots (parent == nil) and walk depth-first
		if jsonOutput {
			type segEntry struct {
				Type    string `json:"type"`
				Number  string `json:"number"`
				Depth   int    `json:"depth"`
				Content string `json:"content"`
			}
			var entries []segEntry
			var walk func(segs []map[string]any, depth int)
			walk = func(segs []map[string]any, depth int) {
				for _, seg := range segs {
					entries = append(entries, segEntry{
						Type:    strField(seg, "type_name"),
						Number:  strField(seg, "number"),
						Depth:   depth,
						Content: strField(seg, "segment_data"),
					})
					if children, ok := seg["children"].([]any); ok {
						var childMaps []map[string]any
						for _, c := range children {
							if m, ok := c.(map[string]any); ok {
								childMaps = append(childMaps, m)
							}
						}
						walk(childMaps, depth+1)
					}
				}
			}
			walk(allSegments, 0)
			printJSON(entries)
			return nil
		}

		var walk func(segs []map[string]any, depth int)
		walk = func(segs []map[string]any, depth int) {
			indent := strings.Repeat("  ", depth)
			for _, seg := range segs {
				typeName := strField(seg, "type_name")
				number := strField(seg, "number")
				content := strField(seg, "segment_data")
				if content != "" {
					fmt.Printf("%s[%s %s] %s\n", indent, typeName, number, content)
				} else {
					fmt.Printf("%s[%s %s]\n", indent, typeName, number)
				}
				if children, ok := seg["children"].([]any); ok {
					var childMaps []map[string]any
					for _, c := range children {
						if m, ok := c.(map[string]any); ok {
							childMaps = append(childMaps, m)
						}
					}
					walk(childMaps, depth+1)
				}
			}
		}
		walk(allSegments, 0)
		printContext()
		return nil
	},
}

func init() {
	expressionListCmd.Flags().StringVar(&expressionIdeaID, "idea", "", "Idea ID (overrides active context)")
	expressionCmd.AddCommand(expressionListCmd, expressionShowCmd, expressionPrintCmd)
}
