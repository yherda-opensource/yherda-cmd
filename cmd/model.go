package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// modelCmd is the root of the Subject-generic command family — it operates on
// the platform's Subject base class rather than per-type (Person/Place/Thing)
// commands, taking a bare Subject id and calling a Subject-generic capability.
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Operate on any Subject by id",
	Long:  "The model command family operates on the platform's Subject base class, working the same way regardless of concrete subtype (Person/Place/Thing/Identity/Disposition/Goal/Step/Belief).",
}

// resolveSubjectID resolves the target Subject id for a bare-Subject-id
// command: an explicit arg always wins, but also resets the breadcrumb
// stack to just that id (a one-off override starts a fresh trail rather
// than drilling deeper into whatever trail 'model use' had built up).
// With no arg, it reads the top of the stack, unchanged.
func resolveSubjectID(args []string) (string, error) {
	ctx, err := config.LoadContext()
	if err != nil {
		return "", err
	}
	if len(args) > 0 {
		ctx.ResetSubject(args[0])
		if err := config.SaveContext(ctx); err != nil {
			return "", err
		}
		return args[0], nil
	}
	if ctx.Subject() == "" {
		return "", fmt.Errorf("no active subject — pass a subject id or run 'yherda model use <subject-id>'")
	}
	return ctx.Subject(), nil
}

var modelShowCmd = &cobra.Command{
	Use:   "show [<id>]",
	Short: "Show a Subject",
	Long:  "Shows a Subject's core fields, regardless of its concrete subtype. Uses the active subject unless an id is passed.",
	Example: `  yherda model show
  yherda model show 42`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID, err := resolveSubjectID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result map[string]any
		if err := client.Get("/subject/"+subjectID+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "name", "subject_type", "has_perspective", "has_self"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		return nil
	},
}

var modelListIdeaID string
var modelListType string
var modelListSearch string

var modelListCmd = &cobra.Command{
	Use:   "list [<idea-id>]",
	Short: "List Subjects for an idea",
	Long:  "Lists every Subject belonging to an idea, regardless of concrete subtype. Uses the active idea unless <idea-id> is passed.",
	Example: `  yherda model list
  yherda model list 42
  yherda model list 42 --type Belief --search "king"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := ""
		if len(args) > 0 {
			ideaID = args[0]
		}
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			fmt.Println("No active idea — showing ideas instead. Run 'yherda ideas use <id>' to select one.")
			fmt.Println("Note: the ids below are IDEA ids, not Subject ids — they are a different id space and will not work with 'model show' or other model commands.")
			client := mustClient()
			var ideas []map[string]any
			if err := client.Get("/idea/", &ideas); err != nil {
				return err
			}
			if jsonOutput {
				printJSON(ideas)
				return nil
			}
			w := newTabWriter()
			fmt.Fprintln(w, "IDEA ID\tNAME\tABSTRACT")
			for _, row := range ideas {
				fmt.Fprintf(w, "%s\t%s\t%s\n", strField(row, "id"), strField(row, "name"), strField(row, "abstract"))
			}
			w.Flush()
			printContext()
			return nil
		}
		path := "/idea/" + ideaID + "/subjects/"
		q := url.Values{}
		if modelListType != "" {
			q.Set("subject_type", modelListType)
		}
		if modelListSearch != "" {
			q.Set("search", modelListSearch)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get(path, &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tHAS PERSPECTIVE\tHAS SELF")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				strField(row, "id"), strField(row, "name"), strField(row, "subject_type"),
				strField(row, "has_perspective"), strField(row, "has_self"))
		}
		w.Flush()
		return nil
	},
}

var modelUseCmd = &cobra.Command{
	Use:   "use [<subject-id>]",
	Short: "Set the active Subject, or show the current breadcrumb stack",
	Long: "Pushes a Subject onto the active-Subject breadcrumb stack, making it active. Does not clear active " +
		"person/place/thing/state/goal — the active Subject is orthogonal to those, not a replacement for any of " +
		"them. With no id, prints the current breadcrumb stack instead. Use 'model back' to pop back to the " +
		"previous Subject.",
	Example: `  yherda model use 42
  yherda model use`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			if len(ctx.SubjectStack) == 0 {
				fmt.Println("No active subject stack.")
				return nil
			}
			fmt.Printf("Subject stack: %s (active)\n", breadcrumbTrail(ctx.SubjectStack))
			return nil
		}
		ctx.PushSubject(args[0])
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active subject set to %s\n", args[0])
		return nil
	},
}

// breadcrumbTrail renders a Subject stack bottom-to-top, e.g. "1 -> 7 -> 42".
func breadcrumbTrail(stack []string) string {
	trail := stack[0]
	for _, id := range stack[1:] {
		trail += " -> " + id
	}
	return trail
}

var modelBackCmd = &cobra.Command{
	Use:   "back",
	Short: "Pop the active Subject, returning to the previous one",
	Long: "Pops the top of the active-Subject breadcrumb stack, returning to whichever Subject was active before " +
		"the last 'model use'. Errors if there is nothing to go back to.",
	Example: `  yherda model back`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		popped, ok := ctx.PopSubject()
		if !ok {
			return fmt.Errorf("no previous subject to go back to")
		}
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		if newTop := ctx.Subject(); newTop != "" {
			fmt.Printf("Left subject %s, active subject is now %s\n", popped, newTop)
		} else {
			fmt.Printf("Left subject %s, no active subject remains\n", popped)
		}
		return nil
	},
}

// --- dispositions ---

var modelDispositionsCmd = &cobra.Command{
	Use:   "dispositions",
	Short: "Manage a Subject's dispositions",
}

var modelDispositionsListCmd = &cobra.Command{
	Use:   "list [<id>]",
	Short: "List dispositions for a Subject",
	Long:  "Lists dispositions for a Subject. Uses the active subject unless an id is passed.",
	Example: `  yherda model dispositions list
  yherda model dispositions list 42`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID, err := resolveSubjectID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/subject/"+subjectID+"/dispositions/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tTYPE\tNAME")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\n", strField(row, "id"), strField(row, "type"), strField(row, "name"))
		}
		w.Flush()
		printContextWithSubject(client, subjectID)
		return nil
	},
}

// --- states ---

var modelStatesCmd = &cobra.Command{
	Use:   "states",
	Short: "Manage a Subject's states",
}

var modelStatesListCmd = &cobra.Command{
	Use:   "list [<id>]",
	Short: "List states for a Subject",
	Long:  "Lists states for a Subject. Uses the active subject unless an id is passed.",
	Example: `  yherda model states list
  yherda model states list 42`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID, err := resolveSubjectID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/subject/"+subjectID+"/states/", &result); err != nil {
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
		printContextWithSubject(client, subjectID)
		return nil
	},
}

var modelStatesUseCmd = &cobra.Command{
	Use:     "use <state-id>",
	Short:   "Set the active state",
	Long:    "Sets the active state in context. Does not clear active person/place/thing — a state is scoped to whichever Subject is already active, not a replacement for it.",
	Example: `  yherda model states use 7`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.State = args[0]
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active state set to %s\n", args[0])
		return nil
	},
}

// --- states dispositions ---

var modelStatesDispositionsCmd = &cobra.Command{
	Use:   "dispositions",
	Short: "Manage which dispositions are bundled into a state",
}

func resolveStateID(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	ctx, err := config.LoadContext()
	if err != nil {
		return "", err
	}
	if ctx.State == "" {
		return "", fmt.Errorf("no active state — pass a state id or run 'yherda model states use <state-id>'")
	}
	return ctx.State, nil
}

var modelStatesDispositionsListCmd = &cobra.Command{
	Use:   "list [<state-id>]",
	Short: "List dispositions bundled into a state",
	Long:  "Lists dispositions bundled into a state. Uses the active state unless a state id is passed.",
	Example: `  yherda model states dispositions list
  yherda model states dispositions list 7`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stateID, err := resolveStateID(args)
		if err != nil {
			return err
		}
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/state/"+stateID+"/dispositions/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		fmt.Fprintln(w, "ID\tTYPE\tNAME")
		for _, row := range result {
			fmt.Fprintf(w, "%s\t%s\t%s\n", strField(row, "id"), strField(row, "type"), strField(row, "name"))
		}
		w.Flush()
		return nil
	},
}

func init() {
	modelListCmd.Flags().StringVar(&modelListType, "type", "", "Filter by subject_type (optional)")
	modelListCmd.Flags().StringVar(&modelListSearch, "search", "", "Case-insensitive substring match on name (optional)")

	modelDispositionsCmd.AddCommand(modelDispositionsListCmd)

	modelStatesDispositionsCmd.AddCommand(modelStatesDispositionsListCmd)
	modelStatesCmd.AddCommand(modelStatesListCmd, modelStatesUseCmd, modelStatesDispositionsCmd)

	modelCmd.AddCommand(modelShowCmd, modelListCmd, modelUseCmd, modelBackCmd, modelDispositionsCmd, modelStatesCmd)
}
