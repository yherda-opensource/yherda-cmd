package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var arcCmd = &cobra.Command{
	Use:   "arc",
	Short: "Manage arcs",
	Long:  "An arc is a person's want and the sequence of beats that pursue it. Arcs belong to a person.",
}

var arcPersonID string
var arcIdeaID string

func listArcs(client *api.Client, personID string) error {
	var result []map[string]any
	if err := client.Get("/role/"+personID+"/arcs/", &result); err != nil {
		return err
	}
	if jsonOutput {
		printJSON(result)
		return nil
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tWANT")
	for _, row := range result {
		fmt.Fprintf(w, "%s\t%s\n", strField(row, "id"), strField(row, "want"))
	}
	w.Flush()
	printContext()
	return nil
}

var arcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List arcs for a person or idea",
	Long:  "Lists arcs for a person (default, using the active person unless --person is passed) or, with --idea, every arc across every person in that idea.",
	Example: `  yherda arc list
  yherda arc list --person 7
  yherda arc list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if arcIdeaID != "" {
			client := mustClient()
			var result []map[string]any
			if err := client.Get("/storyline/"+arcIdeaID+"/arcs/", &result); err != nil {
				return err
			}
			if jsonOutput {
				printJSON(result)
				return nil
			}
			w := newTabWriter()
			fmt.Fprintln(w, "ID\tWANT")
			for _, row := range result {
				fmt.Fprintf(w, "%s\t%s\n", strField(row, "id"), strField(row, "want"))
			}
			w.Flush()
			printContext()
			return nil
		}

		personID := arcPersonID
		if personID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			personID = ctx.Person
		}
		if personID == "" {
			fmt.Println("No active person — showing persons instead. Run 'yherda person use <id>' to select one.")
			return personListCmd.RunE(cmd, args)
		}
		if cmd.Flags().Changed("person") {
			useParent(func(ctx *config.Context, id string) { ctx.Person = id }, personID)
		}
		return listArcs(mustClient(), personID)
	},
}

var arcUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active arc",
	Long:    "Sets the active arc, and the active person to whichever person owns it. Clears any active place/thing.",
	Example: `  yherda arc use 9`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arcID := args[0]

		client := mustClient()
		var arc map[string]any
		if err := client.Get("/arc/"+arcID+"/", &arc); err != nil {
			return err
		}

		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Arc = arcID
		if roleID := strField(arc, "role"); roleID != "" {
			ctx.Person = roleID
		}
		ctx.Place = ""
		ctx.Thing = ""
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active arc set to %s\n", arcID)
		return nil
	},
}

var arcCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new arc for a person",
	Long:    "Creates a new arc for a person. Uses the active person unless --person is passed.",
	Example: `  yherda arc create --want "to be believed" --person 7`,
	RunE: func(cmd *cobra.Command, args []string) error {
		personID, _ := cmd.Flags().GetString("person")
		if personID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			personID = ctx.Person
		}
		if personID == "" {
			return fmt.Errorf("--person is required (or set active person with 'yherda person use <id>')")
		}
		if cmd.Flags().Changed("person") {
			useParent(func(ctx *config.Context, id string) { ctx.Person = id }, personID)
		}
		want, _ := cmd.Flags().GetString("want")
		if want == "" {
			return fmt.Errorf("--want is required")
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/role/"+personID+"/arcs/", map[string]string{"want": want}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	arcListCmd.Flags().StringVar(&arcPersonID, "person", "", "Person (role) ID (overrides active context)")
	arcListCmd.Flags().StringVar(&arcIdeaID, "idea", "", "Idea ID — lists all arcs across all persons")
	arcCreateCmd.Flags().String("person", "", "Person (role) ID (overrides active context)")
	arcCreateCmd.Flags().String("want", "", "Want statement for the arc (required)")
	arcCmd.AddCommand(arcListCmd, arcUseCmd, arcCreateCmd)
}
