package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// modelAddCmd is the CLI's expressive layer for exercising the platform's
// capability-granting cascades against whatever Subject is active via
// 'model use' — 'model add perspective'/'model add goal' resolve their
// target from ctx.Subject by default, with an optional positional override,
// same fallback shape as resolveGoalID/resolveStateID.
//
// 'model add identity' is deliberately not here: Identity creation is
// hardcoded to Person server-side today (no Subject-generic path exists —
// tracked as GEN-576), and beyond that gap, Identity is the character layer
// that only becomes load-bearing once a Subject needs a Goal (which already
// cascades Self -> Identity -> Perspective on its own via 'model add goal').
// A standalone "give this Subject an Identity" command may not be a real
// need at all — see insight_perspective_vs_identity_capability_boundary.md.
var modelAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a capability to the active Subject",
	Long:  "Adds a capability (Perspective, Goal) to the Subject set via 'model use', with an optional positional id override.",
}

func resolveSubjectID(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	ctx, err := config.LoadContext()
	if err != nil {
		return "", err
	}
	if ctx.Subject == "" {
		return "", fmt.Errorf("no active subject — pass a subject id or run 'yherda model use <subject-id>'")
	}
	return ctx.Subject, nil
}

var modelAddPerspectiveCmd = &cobra.Command{
	Use:   "perspective [<subject-id>]",
	Short: "Add (or lazily materialize) a Perspective on the active Subject",
	Long: "Adds a Perspective on the Subject set via 'model use', or an explicit subject id if passed. " +
		"Idempotent — repeated calls return the same Perspective.",
	Example: `  yherda model add perspective
  yherda model add perspective 42`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID, err := resolveSubjectID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/subject/"+subjectID+"/perspective/", nil, &result); err != nil {
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
		printContextWithSubject(client, subjectID)
		return nil
	},
}

var modelAddGoalSkipConfirm bool

var modelAddGoalCmd = &cobra.Command{
	Use:   "goal [<subject-id>]",
	Short: "Add a Goal to the active Subject",
	Long: "Adds a Goal to the Subject set via 'model use', or an explicit subject id if passed. If the Subject " +
		"has no Self yet, this also cascades a default Identity and that Identity's own Perspective/Disposition " +
		"into existence — you'll be asked to confirm before that happens unless --yes is passed.",
	Example: `  yherda model add goal --want "To find her father"
  yherda model add goal 42 --want "To find her father" --need "To let go of guilt" --tragedy`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID, err := resolveSubjectID(args)
		if err != nil {
			return err
		}
		return createGoalOnSubject(subjectID, modelAddGoalSkipConfirm)
	},
}

func init() {
	modelAddGoalCmd.Flags().StringVar(&modelGoalWant, "want", "", "What the Subject wants")
	modelAddGoalCmd.Flags().StringVar(&modelGoalNeed, "need", "", "What the Subject actually needs")
	modelAddGoalCmd.Flags().BoolVar(&modelGoalTragedy, "tragedy", false, "Whether achieving the want costs the need")
	modelAddGoalCmd.Flags().StringVar(&modelGoalDescription, "description", "", "Free-form description")
	modelAddGoalCmd.Flags().BoolVarP(&modelAddGoalSkipConfirm, "yes", "y", false, "Skip the confirmation prompt when a Self/Identity cascade would be created")

	modelAddCmd.AddCommand(modelAddPerspectiveCmd, modelAddGoalCmd)
	modelCmd.AddCommand(modelAddCmd)
}
