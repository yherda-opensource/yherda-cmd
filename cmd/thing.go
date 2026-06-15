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
}

var thingIdeaID string

func listThings(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/storyline/"+ideaID+"/things/", &result); err != nil {
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := thingIdeaID
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
		return listThings(mustClient(), ideaID)
	},
}

var thingUseCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set the active thing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.ActiveThing = args[0]
		cfg.ActivePerson = ""
		cfg.ActiveArc = ""
		cfg.ActivePlace = ""
		if err := config.SaveConfig(cfg); err != nil {
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
