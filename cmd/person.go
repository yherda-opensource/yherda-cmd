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
	printContext()
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

var personCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new person for an idea",
	RunE: func(cmd *cobra.Command, args []string) error {
		ideaID, _ := cmd.Flags().GetString("idea")
		if ideaID == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ideaID = cfg.ActiveIdea
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
		if err := client.Post("/storyline/"+ideaID+"/roles/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		if id := strField(result, "id"); id != "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cfg.ActivePerson = id
			cfg.ActiveArc = ""
			cfg.ActivePlace = ""
			cfg.ActiveThing = ""
			_ = config.SaveConfig(cfg)
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
