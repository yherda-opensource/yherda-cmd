package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var personCmd = &cobra.Command{
	Use:   "person",
	Short: "Manage persons",
	Long:  "A person is the slot a character fills in your story. Persons hold identities (which can change) and goals. Persons belong to an idea.",
}

var personIdeaID string

func listPersons(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/idea/"+ideaID+"/persons/", &result); err != nil {
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

var personListCmd = &cobra.Command{
	Use:   "list",
	Short: "List persons for an idea",
	Long:  "Lists persons for an idea. Uses the active idea unless --idea is passed.",
	Example: `  yherda person list
  yherda person list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := personIdeaID
		if ideaID == "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ideaID = ctx.Idea
		}
		if ideaID == "" {
			fmt.Println("No active idea — showing ideas instead. Run 'yherda ideas use <id>' to select one.")
			return ideasListCmd.RunE(cmd, args)
		}
		if cmd.Flags().Changed("idea") {
			useParent(func(ctx *config.Context, id string) { ctx.Idea = id }, ideaID)
		}
		return listPersons(mustClient(), ideaID)
	},
}

var personUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active person",
	Long:    "Sets the active person. Clears any active place/thing, since those belong to a specific person.",
	Example: `  yherda person use 7`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Person = args[0]
		ctx.Place = ""
		ctx.Thing = ""
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active person set to %s\n", args[0])
		return nil
	},
}

var personCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new person for an idea",
	Long:    "Creates a new person for an idea and sets it active. Uses the active idea unless --idea is passed.",
	Example: `  yherda person create --name "Detective Marlowe" --idea 42`,
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
		if err := client.Post("/idea/"+ideaID+"/persons/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		if id := strField(result, "id"); id != "" {
			ctx, err := config.LoadContext()
			if err != nil {
				return err
			}
			ctx.Idea = ideaID
			ctx.Person = id
			ctx.Place = ""
			ctx.Thing = ""
			_ = config.SaveContext(ctx)
		}
		printJSON(result)
		return nil
	},
}

func init() {
	personListCmd.Flags().StringVar(&personIdeaID, "idea", "", "Idea ID (overrides active context)")
	personCreateCmd.Flags().String("idea", "", "Idea ID (overrides active context)")
	personCreateCmd.Flags().String("name", "", "Name of the person (required)")
	personCmd.AddCommand(personListCmd, personUseCmd, personCreateCmd)
}
