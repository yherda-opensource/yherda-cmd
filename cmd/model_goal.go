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
	Use:   "list [<subject-id>]",
	Short: "List Goals for a Subject",
	Long:  "Lists Goals for a Subject. Uses the active subject unless a subject id is passed.",
	Example: `  yherda model goals list
  yherda model goals list 42`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID, err := resolveSubjectID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/subject/"+subjectID+"/goals/", &result); err != nil {
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
		printContextWithSubject(client, subjectID)
		return nil
	},
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

func init() {
	modelGoalsCmd.AddCommand(modelGoalsListCmd)

	modelStepsCmd.AddCommand(modelStepsListCmd)

	modelGoalCmd.AddCommand(modelGoalUseCmd)

	modelCmd.AddCommand(modelGoalsCmd, modelGoalCmd, modelStepsCmd)
}
