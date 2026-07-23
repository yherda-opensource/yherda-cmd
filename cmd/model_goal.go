package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// modelGoalsCmd manages Goals on a Subject. Creating a Goal on a Subject with
// no existing Self cascades Self -> default Identity -> that Identity's own
// Perspective/Disposition -> Purpose -> Goal — a much bigger side effect than
// the command name alone suggests, so 'goals create' confirms before doing
// that cascade (see insight_cli_capability_grant_confirmation_pattern.md).
var modelGoalsCmd = &cobra.Command{
	Use:   "goals",
	Short: "Manage a Subject's Goals",
}

var modelGoalsListCmd = &cobra.Command{
	Use:     "list <subject-id>",
	Short:   "List Goals for a Subject",
	Example: `  yherda model goals list 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/subject/"+args[0]+"/goals/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tWANT\tNEED\tTRAGEDY")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				strField(row, "id"), strField(row, "want"), strField(row, "need"), strField(row, "tragedy"))
		}
		w.Flush()
		printContextWithSubject(client, args[0])
		return nil
	},
}

var modelGoalWant string
var modelGoalNeed string
var modelGoalTragedy bool
var modelGoalDescription string
var modelGoalsCreateSkipConfirm bool

var modelGoalsCreateCmd = &cobra.Command{
	Use:   "create <subject-id>",
	Short: "Create a Goal on a Subject",
	Long: "Creates a Goal on a Subject. If the Subject has no Self yet, this also cascades a default Identity " +
		"and that Identity's own Perspective/Disposition into existence — you'll be asked to confirm before that " +
		"happens unless --yes is passed.",
	Example: `  yherda model goals create 42 --want "To find her father"
  yherda model goals create 42 --want "To find her father" --need "To let go of guilt" --tragedy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createGoalOnSubject(args[0], modelGoalsCreateSkipConfirm)
	},
}

// createGoalOnSubject creates a Goal on the given Subject, confirming first
// (unless skipConfirm) when the Subject has no Self yet — shared by
// 'model goals create <subject-id>' and 'model add goal [<subject-id>]',
// which differ only in how they resolve the target Subject id.
func createGoalOnSubject(subjectID string, skipConfirm bool) error {
	if modelGoalWant == "" {
		fmt.Println("Warning: --want is empty. A Goal without a want is technically valid but not very useful.")
	}
	client := mustClient()
	if !skipConfirm {
		var subject map[string]any
		if err := client.Get("/subject/"+subjectID+"/", &subject); err != nil {
			return err
		}
		hasSelf := subject["has_self"] == true
		if !hasSelf {
			name := strField(subject, "name")
			subjectType := strField(subject, "subject_type")
			prompt := fmt.Sprintf(
				"Subject #%s %q (%s) has no Self yet — creating a Goal will also create its Self, a default Identity, and that Identity's own Perspective/Disposition. Continue? [y/N] ",
				subjectID, name, subjectType,
			)
			if !confirm(prompt) {
				return fmt.Errorf("aborted")
			}
		}
	}
	body := map[string]any{
		"want":        modelGoalWant,
		"need":        modelGoalNeed,
		"tragedy":     modelGoalTragedy,
		"description": modelGoalDescription,
	}
	var result map[string]any
	if err := client.Post("/subject/"+subjectID+"/goals/", body, &result); err != nil {
		return err
	}
	printJSON(result)
	printContextWithSubject(client, subjectID)
	return nil
}

// modelGoalCmd is the singular 'goal' command, distinct from the plural
// 'goals' resource-management command above — 'goal use' mirrors the
// 'person use'/'place use' shape of the other Subject-bearing resources.
var modelGoalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage the active Goal",
}

var modelGoalUseCmd = &cobra.Command{
	Use:   "use <goal-id>",
	Short: "Set the active goal",
	Long: "Sets the active goal in context. Does not clear active person/place/thing/state — a Goal attaches " +
		"to a Subject's Purpose, not directly to Person, so which Subject has the active Goal is orthogonal to " +
		"which Person/Place/Thing is active.",
	Example: `  yherda model goal use 15`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Goal = args[0]
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active goal set to %s\n", args[0])
		return nil
	},
}

// --- steps ---

var modelStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Manage a Goal's Steps",
}

func resolveGoalID(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	ctx, err := config.LoadContext()
	if err != nil {
		return "", err
	}
	if ctx.Goal == "" {
		return "", fmt.Errorf("no active goal — pass a goal id or run 'yherda model goal use <goal-id>'")
	}
	return ctx.Goal, nil
}

var modelStepsListCmd = &cobra.Command{
	Use:   "list [<goal-id>]",
	Short: "List Steps for a Goal",
	Long:  "Lists Steps for a Goal, in server order (by number). Uses the active goal unless a goal id is passed.",
	Example: `  yherda model steps list
  yherda model steps list 15`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		goalID, err := resolveGoalID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/goal/"+goalID+"/steps/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNUMBER\tDESCRIPTION")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\n", strField(row, "id"), strField(row, "number"), strField(row, "description"))
		}
		w.Flush()
		return nil
	},
}

var modelStepDescription string
var modelStepNumber int

var modelStepsCreateCmd = &cobra.Command{
	Use:   "create [<goal-id>]",
	Short: "Create a Step on a Goal",
	Long:  "Creates a Step on a Goal. Uses the active goal unless a goal id is passed.",
	Example: `  yherda model steps create --description "Ask the innkeeper"
  yherda model steps create 15 --description "Ask the innkeeper" --number 1`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		goalID, err := resolveGoalID(args)
		if err != nil {
			return err
		}
		if modelStepDescription == "" {
			return fmt.Errorf("--description is required")
		}
		body := map[string]any{"description": modelStepDescription}
		if cmd.Flags().Changed("number") {
			body["number"] = modelStepNumber
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/goal/"+goalID+"/steps/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	modelGoalsCreateCmd.Flags().StringVar(&modelGoalWant, "want", "", "What the Subject wants")
	modelGoalsCreateCmd.Flags().StringVar(&modelGoalNeed, "need", "", "What the Subject actually needs")
	modelGoalsCreateCmd.Flags().BoolVar(&modelGoalTragedy, "tragedy", false, "Whether achieving the want costs the need")
	modelGoalsCreateCmd.Flags().StringVar(&modelGoalDescription, "description", "", "Free-form description")
	modelGoalsCreateCmd.Flags().BoolVarP(&modelGoalsCreateSkipConfirm, "yes", "y", false, "Skip the confirmation prompt when a Self/Identity cascade would be created")
	modelGoalsCmd.AddCommand(modelGoalsListCmd, modelGoalsCreateCmd)

	modelStepsCreateCmd.Flags().StringVar(&modelStepDescription, "description", "", "Step description (required)")
	modelStepsCreateCmd.Flags().IntVar(&modelStepNumber, "number", 0, "Step order number (optional)")
	modelStepsCmd.AddCommand(modelStepsListCmd, modelStepsCreateCmd)

	modelGoalCmd.AddCommand(modelGoalUseCmd)

	modelCmd.AddCommand(modelGoalsCmd, modelGoalCmd, modelStepsCmd)
}
