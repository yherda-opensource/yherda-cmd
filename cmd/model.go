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

var modelShowCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show a Subject",
	Long:    "Shows a Subject's core fields, regardless of its concrete subtype.",
	Example: `  yherda model show 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/subject/"+args[0]+"/", &result); err != nil {
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
	Use:   "use <subject-id>",
	Short: "Set the active Subject",
	Long: "Sets the active Subject in context. Does not clear active person/place/thing/state/goal — the active " +
		"Subject is orthogonal to those, not a replacement for any of them.",
	Example: `  yherda model use 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Subject = args[0]
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active subject set to %s\n", args[0])
		return nil
	},
}

// --- dispositions ---

var modelDispositionsCmd = &cobra.Command{
	Use:   "dispositions",
	Short: "Manage a Subject's dispositions",
}

var modelDispositionType string
var modelDispositionName string
var modelDispositionDeleteID string

var modelDispositionsListCmd = &cobra.Command{
	Use:     "list <id>",
	Short:   "List dispositions for a Subject",
	Example: `  yherda model dispositions list 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/subject/"+args[0]+"/dispositions/", &result); err != nil {
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
		printContextWithSubject(client, args[0])
		return nil
	},
}

var modelDispositionsCreateCmd = &cobra.Command{
	Use:     "create <id>",
	Short:   "Create a disposition on a Subject",
	Long:    "Creates a disposition on a Subject. --type must be one of physical, emotional, mental, spiritual.",
	Example: `  yherda model dispositions create 42 --type emotional --name "Grieving"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch modelDispositionType {
		case "physical", "emotional", "mental", "spiritual":
		default:
			return fmt.Errorf("--type must be one of physical, emotional, mental, spiritual")
		}
		if modelDispositionName == "" {
			return fmt.Errorf("--name is required")
		}
		client := mustClient()
		var result map[string]any
		body := map[string]string{"type": modelDispositionType, "name": modelDispositionName}
		if err := client.Post("/subject/"+args[0]+"/dispositions/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		printContextWithSubject(client, args[0])
		return nil
	},
}

var modelDispositionsDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a disposition from a Subject",
	Example: `  yherda model dispositions delete 42 --disposition 7`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelDispositionDeleteID == "" {
			return fmt.Errorf("--disposition is required")
		}
		client := mustClient()
		body := map[string]string{"disposition": modelDispositionDeleteID}
		if err := client.Delete("/subject/"+args[0]+"/dispositions/", body); err != nil {
			return err
		}
		fmt.Printf("Disposition %s deleted\n", modelDispositionDeleteID)
		printContextWithSubject(client, args[0])
		return nil
	},
}

// --- states ---

var modelStatesCmd = &cobra.Command{
	Use:   "states",
	Short: "Manage a Subject's states",
}

var modelStateName string
var modelStateDeleteID string

var modelStatesListCmd = &cobra.Command{
	Use:     "list <id>",
	Short:   "List states for a Subject",
	Example: `  yherda model states list 42`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result []map[string]any
		if err := client.Get("/subject/"+args[0]+"/states/", &result); err != nil {
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
		printContextWithSubject(client, args[0])
		return nil
	},
}

var modelStatesCreateCmd = &cobra.Command{
	Use:     "create <id>",
	Short:   "Create a state on a Subject",
	Example: `  yherda model states create 42 --name "Act Two"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelStateName == "" {
			return fmt.Errorf("--name is required")
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/subject/"+args[0]+"/states/", map[string]string{"name": modelStateName}, &result); err != nil {
			return err
		}
		printJSON(result)
		printContextWithSubject(client, args[0])
		return nil
	},
}

var modelStatesDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a state from a Subject",
	Example: `  yherda model states delete 42 --state 7`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelStateDeleteID == "" {
			return fmt.Errorf("--state is required")
		}
		client := mustClient()
		body := map[string]string{"state": modelStateDeleteID}
		if err := client.Delete("/subject/"+args[0]+"/states/", body); err != nil {
			return err
		}
		fmt.Printf("State %s deleted\n", modelStateDeleteID)
		printContextWithSubject(client, args[0])
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

var modelStatesDispositionsSetCmd = &cobra.Command{
	Use:   "set <disposition-id> [<state-id>]",
	Short: "Bundle a disposition into a state",
	Long: "Bundles a Disposition into a State. The Disposition must already exist on the same Subject as the State " +
		"(via 'model dispositions create'), and the server enforces at most one Disposition per type per State — " +
		"the CLI surfaces that validation error directly rather than pre-checking client-side.",
	Example: `  yherda model states dispositions set 12
  yherda model states dispositions set 12 7`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		stateID, err := resolveStateID(args[1:])
		if err != nil {
			return err
		}
		client := mustClient()
		var result map[string]any
		body := map[string]string{"disposition": args[0]}
		if err := client.Post("/state/"+stateID+"/dispositions/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var modelStatesDispositionsUnsetCmd = &cobra.Command{
	Use:   "unset <disposition-id> [<state-id>]",
	Short: "Remove a disposition from a state",
	Example: `  yherda model states dispositions unset 12
  yherda model states dispositions unset 12 7`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		stateID, err := resolveStateID(args[1:])
		if err != nil {
			return err
		}
		client := mustClient()
		body := map[string]string{"disposition": args[0]}
		if err := client.Delete("/state/"+stateID+"/dispositions/", body); err != nil {
			return err
		}
		fmt.Printf("Disposition %s unset from state %s\n", args[0], stateID)
		return nil
	},
}

func init() {
	modelListCmd.Flags().StringVar(&modelListType, "type", "", "Filter by subject_type (optional)")
	modelListCmd.Flags().StringVar(&modelListSearch, "search", "", "Case-insensitive substring match on name (optional)")

	modelDispositionsCreateCmd.Flags().StringVar(&modelDispositionType, "type", "", "Disposition type: physical, emotional, mental, spiritual (required)")
	modelDispositionsCreateCmd.Flags().StringVar(&modelDispositionName, "name", "", "Name of the disposition (required)")
	modelDispositionsDeleteCmd.Flags().StringVar(&modelDispositionDeleteID, "disposition", "", "Disposition ID to delete (required)")
	modelDispositionsCmd.AddCommand(modelDispositionsListCmd, modelDispositionsCreateCmd, modelDispositionsDeleteCmd)

	modelStatesCreateCmd.Flags().StringVar(&modelStateName, "name", "", "Name of the state (required)")
	modelStatesDeleteCmd.Flags().StringVar(&modelStateDeleteID, "state", "", "State ID to delete (required)")
	modelStatesDispositionsCmd.AddCommand(modelStatesDispositionsListCmd, modelStatesDispositionsSetCmd, modelStatesDispositionsUnsetCmd)
	modelStatesCmd.AddCommand(modelStatesListCmd, modelStatesCreateCmd, modelStatesDeleteCmd, modelStatesUseCmd, modelStatesDispositionsCmd)

	modelCmd.AddCommand(modelShowCmd, modelListCmd, modelUseCmd, modelDispositionsCmd, modelStatesCmd)
}
