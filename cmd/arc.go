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
	return nil
}

var arcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List arcs for a person or idea",
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
			return nil
		}

		personID := arcPersonID
		if personID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			personID = cfg.ActivePerson
		}
		if personID == "" {
			fmt.Println("No active person — showing persons instead. Run 'yherda person use <id>' to select one.")
			return personListCmd.RunE(cmd, args)
		}
		return listArcs(mustClient(), personID)
	},
}

var arcUseCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set the active arc",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arcID := args[0]

		client := mustClient()
		var arc map[string]any
		if err := client.Get("/arc/"+arcID+"/", &arc); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.ActiveArc = arcID
		if roleID := strField(arc, "role"); roleID != "" {
			cfg.ActivePerson = roleID
		}
		cfg.ActivePlace = ""
		cfg.ActiveThing = ""
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Active arc set to %s\n", arcID)
		return nil
	},
}

func init() {
	arcListCmd.Flags().StringVar(&arcPersonID, "person", "", "Person (role) ID (overrides active context)")
	arcListCmd.Flags().StringVar(&arcIdeaID, "idea", "", "Idea ID — lists all arcs across all persons")
	arcCmd.AddCommand(arcListCmd, arcUseCmd)
}
