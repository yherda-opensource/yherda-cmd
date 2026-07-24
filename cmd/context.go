package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// contextCmd manages Context — a Subject subtype independent of any one
// Perspective/Identity, used to attach and rank Beliefs (BeliefContext) and
// to weight what a Perspective surfaces first (ContextPerspective/
// SubjectContext). Contexts are created nested under Idea, the same pattern
// Identity uses nested under Person — there is no standalone POST /context/.
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage contexts",
	Long:  "A context is a topic or thread of relevance that Beliefs attach to and Perspectives rank. Contexts belong to an idea.",
}

var contextIdeaID string

func listContexts(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/idea/"+ideaID+"/contexts/", &result); err != nil {
		return err
	}
	if jsonOutput {
		printJSON(result)
		return nil
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME")
	for _, row := range result {
		fmt.Fprintf(w, "%s\t%s\n", strField(row, "id"), strField(row, "name"))
	}
	w.Flush()
	printContext()
	return nil
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contexts for an idea",
	Long:  "Lists contexts for an idea. Uses the active idea unless --idea is passed.",
	Example: `  yherda context list
  yherda context list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := contextIdeaID
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			return fmt.Errorf("--idea is required (or set active idea with 'yherda ideas use <id>')")
		}
		if cmd.Flags().Changed("idea") {
			useParent(func(ctx *config.Context, id string) { ctx.Idea = id }, ideaID)
		}
		return listContexts(mustClient(), ideaID)
	},
}

var contextUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active context",
	Long:    "Sets the active context, stored in the .yherda context file.",
	Example: `  yherda context use 9`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Context = args[0]
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active context set to %s\n", args[0])
		return nil
	},
}

var contextCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new context for an idea",
	Long:    "Creates a new context for an idea and sets it active. Uses the active idea unless --idea is passed.",
	Example: `  yherda context create --name "The King's Betrayal" --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID, _ := cmd.Flags().GetString("idea")
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			return fmt.Errorf("--idea is required (or set active idea with 'yherda ideas use <id>')")
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/idea/"+ideaID+"/contexts/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		if id := strField(result, "id"); id != "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ctx.Idea = ideaID
			ctx.Context = id
			_ = config.SaveContext(ctx)
		}
		printJSON(result)
		return nil
	},
}

// --- context belief ---

var contextBeliefCmd = &cobra.Command{
	Use:   "belief",
	Short: "Link Beliefs to a context",
	Long:  "Context-first wrapper around BeliefContext, the same join 'model belief contexts' manages from the Belief side.",
}

var contextBeliefContextID string
var contextBeliefID string
var contextBeliefStatus string
var contextBeliefMode string

func resolveContextID(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	ctx, err := config.LoadContext()
	if err != nil {
		return "", err
	}
	if ctx.Context == "" {
		return "", fmt.Errorf("--context is required (or set active context with 'yherda context use <id>')")
	}
	return ctx.Context, nil
}

var contextBeliefAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Attach a Belief to a context",
	Long:  "Attaches a Belief to a context. Uses the active context unless --context is passed. --status defaults to Active, --mode defaults to Surfacing on the server when omitted.",
	Example: `  yherda context belief add --belief 12
  yherda context belief add --context 3 --belief 12 --status Emerging --mode Masking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		contextID, err := resolveContextID(contextBeliefContextID)
		if err != nil {
			return err
		}
		if contextBeliefID == "" {
			return fmt.Errorf("--belief is required")
		}
		body := map[string]string{"belief": contextBeliefID}
		if contextBeliefStatus != "" {
			body["status"] = contextBeliefStatus
		}
		if contextBeliefMode != "" {
			body["mode"] = contextBeliefMode
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/context/"+contextID+"/beliefs/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var contextBeliefListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Beliefs attached to a context",
	Long:  "Lists Beliefs attached to a context. Uses the active context unless --context is passed.",
	Example: `  yherda context belief list
  yherda context belief list --context 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		contextID, err := resolveContextID(contextBeliefContextID)
		if err != nil {
			return err
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/context/"+contextID+"/beliefs/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tBELIEF\tSTATEMENT\tSTATUS\tMODE")
		for _, row := range result {
			belief, _ := row["belief"].(map[string]any)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				strField(row, "id"), strField(row, "belief"), strField(belief, "statement"),
				strField(row, "status"), strField(row, "mode"))
		}
		w.Flush()
		return nil
	},
}

var contextBeliefRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Detach a Belief from a context",
	Long:    "Detaches a Belief from a context. Uses the active context unless --context is passed.",
	Example: `  yherda context belief remove --belief 12`,
	RunE: func(cmd *cobra.Command, args []string) error {
		contextID, err := resolveContextID(contextBeliefContextID)
		if err != nil {
			return err
		}
		if contextBeliefID == "" {
			return fmt.Errorf("--belief is required")
		}
		client := mustClient()
		body := map[string]string{"belief": contextBeliefID}
		if err := client.Delete("/context/"+contextID+"/beliefs/", body); err != nil {
			return err
		}
		fmt.Printf("Belief %s detached from context %s\n", contextBeliefID, contextID)
		return nil
	},
}

func init() {
	contextListCmd.Flags().StringVar(&contextIdeaID, "idea", "", "Idea ID (overrides active context)")
	contextCreateCmd.Flags().String("idea", "", "Idea ID (overrides active context)")
	contextCreateCmd.Flags().String("name", "", "Name of the context (required)")

	contextBeliefAddCmd.Flags().StringVar(&contextBeliefContextID, "context", "", "Context ID (overrides active context)")
	contextBeliefAddCmd.Flags().StringVar(&contextBeliefID, "belief", "", "Belief ID (required)")
	contextBeliefAddCmd.Flags().StringVar(&contextBeliefStatus, "status", "", "Emerging, Active, Strained, or Former (defaults to Active on the server)")
	contextBeliefAddCmd.Flags().StringVar(&contextBeliefMode, "mode", "", "Masking or Surfacing (defaults to Surfacing on the server)")

	contextBeliefListCmd.Flags().StringVar(&contextBeliefContextID, "context", "", "Context ID (overrides active context)")

	contextBeliefRemoveCmd.Flags().StringVar(&contextBeliefContextID, "context", "", "Context ID (overrides active context)")
	contextBeliefRemoveCmd.Flags().StringVar(&contextBeliefID, "belief", "", "Belief ID (required)")

	contextBeliefCmd.AddCommand(contextBeliefAddCmd, contextBeliefListCmd, contextBeliefRemoveCmd)
	contextCmd.AddCommand(contextListCmd, contextCreateCmd, contextUseCmd, contextBeliefCmd)
}
