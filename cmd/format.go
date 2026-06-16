package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Manage expression formats",
	Long:  "An expression format defines the segment types and structure an expression template can use — e.g. a screenplay format with scene/action/dialogue segment types.",
}

var formatListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all expression formats",
	Example: `  yherda format list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/expressionformat/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME\tVERSION\tAPPROVED")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				strField(row, "id"),
				strField(row, "name"),
				strField(row, "version"),
				strField(row, "approved"),
			)
		}
		w.Flush()
		printContext()
		return nil
	},
}

var formatShowCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show an expression format and its segment types",
	Example: `  yherda format show 2`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/expressionformat/"+args[0]+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "name", "version", "approved", "owner", "created"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()

		if types, ok := result["types"].([]any); ok && len(types) > 0 {
			fmt.Println()
			tw := newTabWriter()
			fmt.Fprintln(tw, "TYPE ID\tNAME\tALLOWS MOMENT\tALLOWS ACTOR\tIS REFERENCE\tCHILDREN")
			for _, t := range types {
				if row, ok := t.(map[string]any); ok {
					children := ""
					if ch, ok := row["allowed_child_names"].([]any); ok {
						for i, c := range ch {
							if i > 0 {
								children += ", "
							}
							children += fmt.Sprintf("%v", c)
						}
					}
					fmt.Fprintf(tw, "%s\t%s\t%v\t%v\t%v\t%s\n",
						strField(row, "id"),
						strField(row, "name"),
						strField(row, "allows_moment"),
						strField(row, "allows_actor"),
						strField(row, "is_reference"),
						children,
					)
				}
			}
			tw.Flush()
		}
		return nil
	},
}

func init() {
	formatCmd.AddCommand(formatListCmd, formatShowCmd)
}
