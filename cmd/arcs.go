package cmd

import (
	"github.com/spf13/cobra"
)

var arcsCmd = &cobra.Command{
	Use:   "arcs",
	Short: "Manage arcs",
}

var arcsIdeaID string

var arcsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List arcs for an idea",
	RunE: func(cmd *cobra.Command, args []string) error {
		if arcsIdeaID == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result any
		if err := client.Get("/storyline/"+arcsIdeaID+"/arcs/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var arcsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single arc",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		var result any
		if err := client.Get("/arcs/"+args[0]+"/", &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

var arcsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new arc",
	RunE: func(cmd *cobra.Command, args []string) error {
		storylineID, _ := cmd.Flags().GetString("storyline")
		name, _ := cmd.Flags().GetString("name")
		if storylineID == "" || name == "" {
			return cmd.Usage()
		}
		client := mustClient()
		var result any
		body := map[string]string{"name": name, "storyline": storylineID}
		if err := client.Post("/arcs/", body, &result); err != nil {
			return err
		}
		printJSON(result)
		return nil
	},
}

func init() {
	arcsListCmd.Flags().StringVar(&arcsIdeaID, "idea", "", "Idea ID (required)")
	arcsCreateCmd.Flags().String("storyline", "", "Storyline ID (required)")
	arcsCreateCmd.Flags().String("name", "", "Arc name (required)")
	arcsCmd.AddCommand(arcsListCmd, arcsShowCmd, arcsCreateCmd)
}
