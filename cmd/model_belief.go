package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// modelBeliefCmd manages Belief — itself a Subject subtype — and its
// attachment to Context via BeliefContext. Belief creation is idea-scoped
// root create (like Person/Place/Thing), not created under /subject/.
//
// Attachment specificity note (see insight_marketing_idea_voice_as_editor_license_context.md):
// unqualified beliefs generally attach at a Subject's own (or base) Perspective;
// fully-qualified ones ("X, in identity Y, when Z") attach at the specific
// Identity/Disposition's own Perspective. The CLI doesn't enforce this — it's
// guidance for whoever is typing the command.
var modelBeliefCmd = &cobra.Command{
	Use:   "belief",
	Short: "Create and manage Beliefs",
}

var modelBeliefIdeaID string
var modelBeliefStatement string
var modelBeliefSubjectID string

var modelBeliefCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Belief",
	Long: "Creates a Belief for an idea. Uses the active idea unless --idea is passed. --subject optionally " +
		"targets another Subject that the belief is about (a separate relationship from Context attachment, " +
		"which is done afterwards via 'model belief contexts add').",
	Example: `  yherda model belief create --statement "The king is not to be trusted"
  yherda model belief create --idea 42 --statement "She loved him once" --subject 7`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := modelBeliefIdeaID
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
		if modelBeliefStatement == "" {
			return fmt.Errorf("--statement is required")
		}
		body := map[string]string{"statement": modelBeliefStatement}
		if modelBeliefSubjectID != "" {
			body["subject"] = modelBeliefSubjectID
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/idea/"+ideaID+"/beliefs/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

// --- belief contexts ---

var modelBeliefContextsCmd = &cobra.Command{
	Use:   "contexts",
	Short: "Manage a Belief's attachment to Contexts",
	Long:  "Manages BeliefContext — a Belief's attachment to a Context, with a status and mode.",
}

var modelBeliefContextID string
var modelBeliefID string
var modelBeliefStatus string
var modelBeliefMode string

var modelBeliefContextsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Attach a Belief to a Context",
	Long:  "Attaches a Belief to a Context. --status defaults to Active, --mode defaults to Surfacing on the server when omitted.",
	Example: `  yherda model belief contexts add --context 3 --belief 12
  yherda model belief contexts add --context 3 --belief 12 --status Emerging --mode Masking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelBeliefContextID == "" {
			return fmt.Errorf("--context is required")
		}
		if modelBeliefID == "" {
			return fmt.Errorf("--belief is required")
		}
		body := map[string]string{"belief": modelBeliefID}
		if modelBeliefStatus != "" {
			body["status"] = modelBeliefStatus
		}
		if modelBeliefMode != "" {
			body["mode"] = modelBeliefMode
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/context/"+modelBeliefContextID+"/beliefs/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var modelBeliefContextsUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update a Belief's status/mode on a Context",
	Example: `  yherda model belief contexts update --context 3 --belief 12 --status Strained`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelBeliefContextID == "" {
			return fmt.Errorf("--context is required")
		}
		if modelBeliefID == "" {
			return fmt.Errorf("--belief is required")
		}
		body := map[string]string{"belief": modelBeliefID}
		if modelBeliefStatus != "" {
			body["status"] = modelBeliefStatus
		}
		if modelBeliefMode != "" {
			body["mode"] = modelBeliefMode
		}
		client := mustClient()
		var result map[string]any
		if err := client.Patch("/context/"+modelBeliefContextID+"/beliefs/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var modelBeliefContextsRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Detach a Belief from a Context",
	Example: `  yherda model belief contexts remove --context 3 --belief 12`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelBeliefContextID == "" {
			return fmt.Errorf("--context is required")
		}
		if modelBeliefID == "" {
			return fmt.Errorf("--belief is required")
		}
		client := mustClient()
		body := map[string]string{"belief": modelBeliefID}
		if err := client.Delete("/context/"+modelBeliefContextID+"/beliefs/", body); err != nil {
			return err
		}
		fmt.Printf("Belief %s detached from context %s\n", modelBeliefID, modelBeliefContextID)
		return nil
	},
}

func init() {
	modelBeliefCreateCmd.Flags().StringVar(&modelBeliefIdeaID, "idea", "", "Idea ID (overrides active context)")
	modelBeliefCreateCmd.Flags().StringVar(&modelBeliefStatement, "statement", "", "Belief statement (required)")
	modelBeliefCreateCmd.Flags().StringVar(&modelBeliefSubjectID, "subject", "", "Subject ID this belief is about (optional)")

	modelBeliefContextsAddCmd.Flags().StringVar(&modelBeliefContextID, "context", "", "Context ID (required)")
	modelBeliefContextsAddCmd.Flags().StringVar(&modelBeliefID, "belief", "", "Belief ID (required)")
	modelBeliefContextsAddCmd.Flags().StringVar(&modelBeliefStatus, "status", "", "Emerging, Active, Strained, or Former (defaults to Active on the server)")
	modelBeliefContextsAddCmd.Flags().StringVar(&modelBeliefMode, "mode", "", "Masking or Surfacing (defaults to Surfacing on the server)")

	modelBeliefContextsUpdateCmd.Flags().StringVar(&modelBeliefContextID, "context", "", "Context ID (required)")
	modelBeliefContextsUpdateCmd.Flags().StringVar(&modelBeliefID, "belief", "", "Belief ID (required)")
	modelBeliefContextsUpdateCmd.Flags().StringVar(&modelBeliefStatus, "status", "", "Emerging, Active, Strained, or Former")
	modelBeliefContextsUpdateCmd.Flags().StringVar(&modelBeliefMode, "mode", "", "Masking or Surfacing")

	modelBeliefContextsRemoveCmd.Flags().StringVar(&modelBeliefContextID, "context", "", "Context ID (required)")
	modelBeliefContextsRemoveCmd.Flags().StringVar(&modelBeliefID, "belief", "", "Belief ID (required)")

	modelBeliefContextsCmd.AddCommand(modelBeliefContextsAddCmd, modelBeliefContextsUpdateCmd, modelBeliefContextsRemoveCmd)
	modelBeliefCmd.AddCommand(modelBeliefCreateCmd, modelBeliefContextsCmd)

	modelCmd.AddCommand(modelBeliefCmd)
}
