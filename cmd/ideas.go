package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yherda-opensource/yherda-cmd/internal/api"
	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

var ideasCmd = &cobra.Command{
	Use:   "ideas",
	Short: "Manage ideas",
}

func listIdeas(client *api.Client) error {
	var result []map[string]any
	if err := client.Get("/storyline/", &result); err != nil {
		return err
	}
	if jsonOutput {
		printJSON(result)
		return nil
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tABSTRACT")
	for _, row := range result {
		fmt.Fprintf(w, "%s\t%s\t%s\n", strField(row, "id"), strField(row, "name"), strField(row, "abstract"))
	}
	w.Flush()
	return nil
}

var ideasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ideas in the active workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listIdeas(mustClient())
	},
}

var ideasShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single idea",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result map[string]any
		if err := client.Get("/storyline/"+args[0]+"/", &result); err != nil {
			return err
		}
		if jsonOutput {
			printJSON(result)
			return nil
		}
		w := newTabWriter()
		for _, key := range []string{"id", "name", "abstract", "idea_type"} {
			fmt.Fprintf(w, "%s\t%s\n", key, strField(result, key))
		}
		w.Flush()
		return nil
	},
}

var ideasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new idea",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result map[string]any
		if err := client.Post("/storyline/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var ideasUseCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set the active idea",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cfg.ActiveIdea = args[0]
		cfg.ActivePerson = ""
		cfg.ActiveArc = ""
		cfg.ActivePlace = ""
		cfg.ActiveThing = ""
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Active idea set to %s\n", args[0])
		return nil
	},
}

func init() {
	ideasCreateCmd.Flags().String("name", "", "Name of the idea (required)")
	ideasCmd.AddCommand(ideasListCmd, ideasShowCmd, ideasCreateCmd, ideasUseCmd)
}
