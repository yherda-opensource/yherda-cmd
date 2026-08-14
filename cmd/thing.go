package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var thingCmd = &cobra.Command{
	Use:   "thing",
	Short: "Manage things",
	Long:  "A thing is a notable object in your story. Things hold dispositions (their state at a point in the story) and belong to an idea.",
}

var thingIdeaID string

func listThings(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/idea/"+ideaID+"/things/", &result); err != nil {
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

var thingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List things for an idea",
	Long:  "Lists things for an idea. Uses the active idea unless --idea is passed.",
	Example: `  yherda thing list
  yherda thing list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := thingIdeaID
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
		return listThings(mustClient(), ideaID)
	},
}

var thingUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active thing",
	Long:    "Sets the active thing. Clears any active person/place, since those aren't scoped under a thing.",
	Example: `  yherda thing use 5`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Thing = args[0]
		ctx.Person = ""
		ctx.Place = ""
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active thing set to %s\n", args[0])
		return nil
	},
}

func init() {
	thingListCmd.Flags().StringVar(&thingIdeaID, "idea", "", "Idea ID (overrides active context)")
	thingCmd.AddCommand(thingListCmd, thingUseCmd)
}
