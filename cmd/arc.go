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
	printContext()
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
			printContext()
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
		if cmd.Flags().Changed("person") {
			useParent(func(cfg *config.Config, id string) { cfg.ActivePerson = id }, personID)
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

var arcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new arc for a person",
	RunE: func(cmd *cobra.Command, args []string) error {
		personID, _ := cmd.Flags().GetString("person")
		if personID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			personID = cfg.ActivePerson
		}
		if personID == "" {
			return fmt.Errorf("--person is required (or set active person with 'yherda person use <id>')")
		}
		if cmd.Flags().Changed("person") {
			useParent(func(cfg *config.Config, id string) { cfg.ActivePerson = id }, personID)
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
