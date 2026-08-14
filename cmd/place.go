package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var placeCmd = &cobra.Command{
	Use:   "place",
	Short: "Manage places",
	Long:  "A place is a physical or notional location in your story. Places hold settings (a specific configuration of a place at a point in time) and belong to an idea.",
}

var placeIdeaID string

func listPlaces(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/idea/"+ideaID+"/places/", &result); err != nil {
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

var placeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List places for an idea",
	Long:  "Lists places for an idea. Uses the active idea unless --idea is passed.",
	Example: `  yherda place list
  yherda place list --idea 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := placeIdeaID
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
		return listPlaces(mustClient(), ideaID)
	},
}

var placeUseCmd = &cobra.Command{
	Use:     "use <id>",
	Short:   "Set the active place",
	Long:    "Sets the active place. Clears any active person/thing, since those aren't scoped under a place.",
	Example: `  yherda place use 12`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := config.LoadContext()
		if err != nil {
			return err
		}
		ctx.Place = args[0]
		ctx.Person = ""
		ctx.Thing = ""
		if err := config.SaveContext(ctx); err != nil {
			return err
		}
		fmt.Printf("Active place set to %s\n", args[0])
		return nil
	},
}

func init() {
	placeListCmd.Flags().StringVar(&placeIdeaID, "idea", "", "Idea ID (overrides active context)")
	placeCmd.AddCommand(placeListCmd, placeUseCmd)
}
