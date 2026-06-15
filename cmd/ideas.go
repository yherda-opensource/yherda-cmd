package cmd

import (
	"github.com/spf13/cobra"
)

var ideasCmd = &cobra.Command{
	Use:   "ideas",
	Short: "Manage ideas",
	Long:  "List, show, and create ideas. Ideas are the top-level container in Yherda (API path: /storyline/).",
}

var ideasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ideas in the active workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result any
		if err := client.Get("/storyline/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var ideasShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single idea",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result any
		if err := client.Get("/storyline/"+args[0]+"/", &result); err != nil {
			return err
		}
		printJSON(result)
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
		var result any
		if err := client.Post("/storyline/", map[string]string{"name": name}, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	ideasCreateCmd.Flags().String("name", "", "Name of the idea (required)")
	ideasCmd.AddCommand(ideasListCmd, ideasShowCmd, ideasCreateCmd)
}
