package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// modelPerspectiveCmd operates on a Subject's Perspective — Subject-generic,
// per YOS-73's governing principle. Perspective materializes lazily on first
// 'get' call (idempotent get_or_create on the server).
var modelPerspectiveCmd = &cobra.Command{
	Use:   "perspective",
	Short: "Manage a Subject's Perspective",
}

// --- perspective contexts ---

var modelPerspectiveContextsCmd = &cobra.Command{
	Use:   "contexts",
	Short: "Manage the Contexts attached to a Perspective",
	Long: "Manages ContextPerspective — a Perspective's own ranked view of Contexts, independent of " +
		"SubjectContext.priority.",
}

var modelPerspectiveContextsListCmd = &cobra.Command{
	Use:     "list <perspective-id>",
	Short:   "List Contexts attached to a Perspective",
	Example: `  yherda model perspective contexts list 9`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/perspective/"+args[0]+"/contexts/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "CONTEXT ID\tPRIORITY")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\n", strField(row, "context"), strField(row, "priority"))
		}
		w.Flush()
		return nil
	},
}

func init() {
	modelPerspectiveContextsCmd.AddCommand(modelPerspectiveContextsListCmd)
	modelPerspectiveCmd.AddCommand(modelPerspectiveContextsCmd)

	modelCmd.AddCommand(modelPerspectiveCmd)
}
