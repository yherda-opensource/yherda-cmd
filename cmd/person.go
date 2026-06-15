package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var personCmd = &cobra.Command{
	Use:   "person",
	Short: "Manage persons (roles)",
}

var personIdeaID string

func listPersons(client *api.Client, ideaID string) error {
	var result []map[string]any
	if err := client.Get("/storyline/"+ideaID+"/roles/", &result); err != nil {
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
	return nil
}

var personListCmd = &cobra.Command{
	Use:   "list",
	Short: "List persons (roles) for an idea",
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID := personIdeaID
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
		return listPersons(mustClient(), ideaID)
	},
}

var personUseCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set the active person (role)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.ActivePerson = args[0]
		cfg.ActiveArc = ""
		cfg.ActivePlace = ""
		cfg.ActiveThing = ""
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Active person set to %s\n", args[0])
		return nil
	},
}

func init() {
	personListCmd.Flags().StringVar(&personIdeaID, "idea", "", "Idea ID (overrides active context)")
	personCmd.AddCommand(personListCmd, personUseCmd)
}
