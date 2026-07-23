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

var modelPerspectiveGetCmd = &cobra.Command{
	Use:   "get <subject-id>",
	Short: "Get (or lazily materialize) a Subject's Perspective",
	Long: "Gets a Subject's Perspective, materializing it on first call. This is a POST under the hood " +
		"despite being a 'get' from the CLI's perspective — the server's get_or_create is idempotent, " +
		"so repeated calls return the same Perspective.",
	Example: `  yherda model perspective get 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Post("/subject/"+args[0]+"/perspective/", nil, &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "name"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		printContextWithSubject(client, args[0])
		return nil
	},
}

// --- perspective contexts ---

var modelPerspectiveContextsCmd = &cobra.Command{
	Use:   "contexts",
	Short: "Manage the Contexts attached to a Perspective",
	Long: "Manages ContextPerspective — a Perspective's own ranked view of Contexts, independent of " +
		"SubjectContext.priority.",
}

var modelPerspectiveContextID string
var modelPerspectiveContextPriority int

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

var modelPerspectiveContextsAddCmd = &cobra.Command{
	Use:   "add <perspective-id>",
	Short: "Attach a Context to a Perspective",
	Example: `  yherda model perspective contexts add 9 --context 3
  yherda model perspective contexts add 9 --context 3 --priority 1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelPerspectiveContextID == "" {
			return fmt.Errorf("--context is required")
		}
		body := map[string]any{"context": modelPerspectiveContextID}
		if cmd.Flags().Changed("priority") {
			body["priority"] = modelPerspectiveContextPriority
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/perspective/"+args[0]+"/contexts/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var modelPerspectiveContextsUpdateCmd = &cobra.Command{
	Use:     "update <perspective-id>",
	Short:   "Update a Context's priority on a Perspective",
	Example: `  yherda model perspective contexts update 9 --context 3 --priority 2`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelPerspectiveContextID == "" {
			return fmt.Errorf("--context is required")
		}
		if !cmd.Flags().Changed("priority") {
			return fmt.Errorf("--priority is required")
		}
		body := map[string]any{"context": modelPerspectiveContextID, "priority": modelPerspectiveContextPriority}
		client := mustClient()
		var result map[string]any
		if err := client.Patch("/perspective/"+args[0]+"/contexts/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var modelPerspectiveContextsRemoveCmd = &cobra.Command{
	Use:     "remove <perspective-id>",
	Short:   "Detach a Context from a Perspective",
	Example: `  yherda model perspective contexts remove 9 --context 3`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelPerspectiveContextID == "" {
			return fmt.Errorf("--context is required")
		}
		client := mustClient()
		body := map[string]string{"context": modelPerspectiveContextID}
		if err := client.Delete("/perspective/"+args[0]+"/contexts/", body); err != nil {
			return err
		}
		fmt.Printf("Context %s detached\n", modelPerspectiveContextID)
		return nil
	},
}

func init() {
	modelPerspectiveContextsAddCmd.Flags().StringVar(&modelPerspectiveContextID, "context", "", "Context ID (required)")
	modelPerspectiveContextsAddCmd.Flags().IntVar(&modelPerspectiveContextPriority, "priority", 0, "Priority (optional)")
	modelPerspectiveContextsUpdateCmd.Flags().StringVar(&modelPerspectiveContextID, "context", "", "Context ID (required)")
	modelPerspectiveContextsUpdateCmd.Flags().IntVar(&modelPerspectiveContextPriority, "priority", 0, "Priority (required)")
	modelPerspectiveContextsRemoveCmd.Flags().StringVar(&modelPerspectiveContextID, "context", "", "Context ID (required)")

	modelPerspectiveContextsCmd.AddCommand(
		modelPerspectiveContextsListCmd,
		modelPerspectiveContextsAddCmd,
		modelPerspectiveContextsUpdateCmd,
		modelPerspectiveContextsRemoveCmd,
	)
	modelPerspectiveCmd.AddCommand(modelPerspectiveGetCmd, modelPerspectiveContextsCmd)

	modelCmd.AddCommand(modelPerspectiveCmd)
}
