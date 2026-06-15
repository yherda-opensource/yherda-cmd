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
}

var placeIdeaID string

func listPlaces(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/storyline/"+ideaID+"/places/", &result); err != nil {
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := placeIdeaID
		if ideaID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ideaID = cfg.ActiveIdea
		}
		if ideaID == "" {
			fmt.Println("No active idea — showing ideas instead. Run 'yherda ideas use <id>' to select one.")
			return ideasListCmd.RunE(cmd, args)
		}
		return listPlaces(mustClient(), ideaID)
	},
}

var placeUseCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set the active place",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.ActivePlace = args[0]
		cfg.ActivePerson = ""
		cfg.ActiveArc = ""
		cfg.ActiveThing = ""
		if err := config.SaveConfig(cfg); err != nil {
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
